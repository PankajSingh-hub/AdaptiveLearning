package coding

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// OffExecutor disables code execution (used for deployments without an isolated judge).
type OffExecutor struct{}

func (OffExecutor) Languages() []LangInfo { return nil }
func (OffExecutor) Sandboxed() bool       { return false }
func (OffExecutor) Prepare(string, string, int) (Program, error) {
	return nil, errors.New("code execution is disabled on this server")
}

// DockerExecutor runs every compile and every test in a throwaway container:
// no network, capped memory/CPU/processes, read-only filesystem, hard kill timeout.
// The backend must run on a host with Docker (not inside a container).
type DockerExecutor struct {
	image    string
	sem      chan struct{}
	overhead time.Duration // container start-up time, subtracted from reported runtimes
}

func NewDockerExecutor(image string, parallel int) (*DockerExecutor, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if out, err := exec.CommandContext(ctx, "docker", "image", "inspect", image).CombinedOutput(); err != nil {
		return nil, fmt.Errorf("judge image %q not available (run: docker build -t %s ./judge): %s",
			image, image, strings.TrimSpace(string(out)))
	}
	e := &DockerExecutor{image: image, sem: make(chan struct{}, parallel)}
	e.overhead = e.measureOverhead()
	log.Printf("coding: sandboxed Docker judge ready (image %s, start-up overhead %v)", image, e.overhead)
	return e, nil
}

func (e *DockerExecutor) Languages() []LangInfo {
	return []LangInfo{{"cpp", "C++"}, {"java", "Java"}, {"python", "Python"}}
}

func (e *DockerExecutor) Sandboxed() bool { return true }

func containerName() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "judge-" + hex.EncodeToString(b)
}

func (e *DockerExecutor) args(name, dir string, readOnly, stdin bool, cmd ...string) []string {
	a := []string{"run", "--rm", "--name", name, "--network", "none",
		"--memory", "512m", "--memory-swap", "512m", "--cpus", "1", "--pids-limit", "256",
		"--cap-drop", "ALL", "--security-opt", "no-new-privileges"}
	if stdin {
		a = append(a, "-i")
	}
	if readOnly {
		a = append(a, "--read-only", "--tmpfs", "/tmp:rw,size=64m")
	}
	if dir != "" {
		mount := dir + ":/work"
		if readOnly {
			mount += ":ro"
		}
		a = append(a, "-v", mount, "-w", "/work")
	}
	a = append(a, e.image)
	return append(a, cmd...)
}

func (e *DockerExecutor) measureOverhead() time.Duration {
	best := time.Duration(math.MaxInt64)
	for i := 0; i < 3; i++ {
		start := time.Now()
		if err := exec.Command("docker", e.args(containerName(), "", true, false, "true")...).Run(); err == nil {
			if d := time.Since(start); d < best {
				best = d
			}
		}
	}
	if best == time.Duration(math.MaxInt64) {
		return 0
	}
	return best
}

func killContainer(name string) { _ = exec.Command("docker", "rm", "-f", name).Run() }

func (e *DockerExecutor) Prepare(lang, source string, offset int) (Program, error) {
	dir, err := os.MkdirTemp("", "judge-*")
	if err != nil {
		return nil, err
	}
	_ = os.Chmod(dir, 0o777) // the container user (uid 10001) must be able to write the compiled output
	var file string
	var compile, run []string
	switch lang {
	case "cpp":
		file, compile, run = "Main.cpp", []string{"g++", "-O2", "-std=c++17", "-o", "prog", "Main.cpp"}, []string{"./prog"}
	case "java":
		file, compile, run = "Main.java", []string{"javac", "-J-Xmx256m", "Main.java"},
			[]string{"java", "-XX:-UsePerfData", "-Xmx256m", "-Xss64m", "Main"}
	case "python":
		file, compile, run = "main.py", []string{"python3", "-m", "py_compile", "main.py"}, []string{"python3", "main.py"}
	default:
		os.RemoveAll(dir)
		return nil, fmt.Errorf("language %q is not supported", lang)
	}
	if err := os.WriteFile(filepath.Join(dir, file), []byte(source), 0o644); err != nil {
		os.RemoveAll(dir)
		return nil, err
	}

	e.sem <- struct{}{}
	defer func() { <-e.sem }()
	ctx, cancel := context.WithTimeout(context.Background(), 70*time.Second)
	defer cancel()
	name := containerName()
	cmd := append([]string{"timeout", "-s", "KILL", "40"}, compile...)
	out, err := exec.CommandContext(ctx, "docker", e.args(name, dir, false, false, cmd...)...).CombinedOutput()
	if err != nil {
		os.RemoveAll(dir)
		if ctx.Err() != nil {
			killContainer(name)
			return nil, &CompileError{Output: "Compilation took too long."}
		}
		msg := cleanOutput(string(out), "/work", offset)
		var ee *exec.ExitError
		if errors.As(err, &ee) && ee.ExitCode() == 125 { // docker itself failed, not the compiler
			return nil, fmt.Errorf("docker could not start the judge container: %s", msg)
		}
		if msg == "" {
			msg = err.Error()
		}
		return nil, &CompileError{Output: msg}
	}
	return &dockerProgram{e: e, dir: dir, cmd: run, offset: offset}, nil
}

type dockerProgram struct {
	e      *DockerExecutor
	dir    string
	cmd    []string
	offset int
}

func (p *dockerProgram) Close() { os.RemoveAll(p.dir) }

func (p *dockerProgram) Run(input string, limit time.Duration) RunResult {
	p.e.sem <- struct{}{}
	defer func() { <-p.e.sem }()
	secs := int(math.Ceil(limit.Seconds()))
	ctx, cancel := context.WithTimeout(context.Background(), limit+10*time.Second)
	defer cancel()
	name := containerName()
	cmd := exec.CommandContext(ctx, "docker",
		p.e.args(name, p.dir, true, true, append([]string{"timeout", "-s", "KILL", strconv.Itoa(secs)}, p.cmd...)...)...)
	cmd.Stdin = strings.NewReader(input)
	cmd.WaitDelay = time.Second
	out, errBuf := &capped{max: 4 << 20}, &capped{max: 16 << 10}
	cmd.Stdout, cmd.Stderr = out, errBuf

	start := time.Now()
	err := cmd.Run()
	elapsed := time.Since(start) - p.e.overhead
	if elapsed < 0 {
		elapsed = 0
	}
	r := RunResult{Stdout: out.String(), Millis: int(elapsed.Milliseconds())}
	r.Stderr = cleanOutput(errBuf.String(), "/work", p.offset)
	switch {
	case ctx.Err() == context.DeadlineExceeded:
		killContainer(name)
		r.TimedOut = true
	case err != nil:
		var ee *exec.ExitError
		if errors.As(err, &ee) && (ee.ExitCode() == 124 || ee.ExitCode() == 137) {
			r.TimedOut = true // killed by the time limit (an out-of-memory kill also looks like this)
		} else {
			r.ExitErr = true
		}
	}
	return r
}