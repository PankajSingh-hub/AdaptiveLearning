package coding

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type LangInfo struct{ ID, Label string }

type RunResult struct {
	Stdout, Stderr string
	Millis         int
	TimedOut       bool
	ExitErr        bool
}

type CompileError struct{ Output string }

func (e *CompileError) Error() string { return e.Output }

// Program is a compiled submission that can be run against many inputs.
type Program interface {
	Run(input string, limit time.Duration) RunResult
	Close()
}

// Executor is the seam for swapping in an isolated judge (Judge0, Piston, a container pool).
type Executor interface {
	Languages() []LangInfo
	Prepare(lang, source string, lineOffset int) (Program, error)
	Sandboxed() bool
}

// LocalExecutor runs code on THIS machine. It is NOT a security sandbox: use it for local
// demos only. It adds timeouts, an output cap, a stripped environment and a concurrency limit.
type LocalExecutor struct {
	langs []LangInfo
	sem   chan struct{}
	
}
func (e *LocalExecutor) Sandboxed() bool { return false }
var probes = []struct {
	info LangInfo
	cmd  []string
}{
	{LangInfo{"cpp", "C++"}, []string{"g++", "--version"}},
	{LangInfo{"java", "Java"}, []string{"javac", "-version"}},
	{LangInfo{"python", "Python"}, []string{"python3", "--version"}},
}

func NewLocalExecutor(parallel int) *LocalExecutor {
	e := &LocalExecutor{sem: make(chan struct{}, parallel)}
	for _, p := range probes {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		err := exec.CommandContext(ctx, p.cmd[0], p.cmd[1:]...).Run()
		cancel()
		if err != nil {
			log.Printf("coding: %s unavailable (%s failed)", p.info.Label, p.cmd[0])
			continue
		}
		e.langs = append(e.langs, p.info)
	}
	log.Printf("coding: available languages: %v", e.langs)
	return e
}

func (e *LocalExecutor) Languages() []LangInfo { return append([]LangInfo(nil), e.langs...) }

func (e *LocalExecutor) has(lang string) bool {
	for _, l := range e.langs {
		if l.ID == lang {
			return true
		}
	}
	return false
}

func (e *LocalExecutor) Prepare(lang, source string, offset int) (Program, error) {
	if !e.has(lang) {
		return nil, fmt.Errorf("language %q is not available on this server", lang)
	}
	dir, err := os.MkdirTemp("", "judge-*")
	if err != nil {
		return nil, err
	}
	var file string
	var compile, run []string
	switch lang {
	case "cpp":
		file, compile, run = "Main.cpp", []string{"g++", "-O2", "-std=c++17", "-o", "prog", "Main.cpp"}, []string{"./prog"}
	case "java":
		file, compile, run = "Main.java", []string{"javac", "-J-Xmx256m", "Main.java"}, []string{"java", "-Xmx256m", "-Xss64m", "Main"}
	default:
		file, compile, run = "main.py", []string{"python3", "-m", "py_compile", "main.py"}, []string{"python3", "main.py"}
	}
	if err := os.WriteFile(filepath.Join(dir, file), []byte(source), 0o600); err != nil {
		os.RemoveAll(dir)
		return nil, err
	}

	e.sem <- struct{}{}
	defer func() { <-e.sem }()
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, compile[0], compile[1:]...)
	cmd.Dir, cmd.Env = dir, minimalEnv(dir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		os.RemoveAll(dir)
		msg := cleanOutput(string(out), dir, offset)
		switch {
		case ctx.Err() != nil:
			msg = "Compilation took too long."
		case strings.TrimSpace(msg) == "":
			msg = err.Error()
		}
		return nil, &CompileError{Output: msg}
	}
	return &localProgram{e: e, dir: dir, cmd: run, offset: offset}, nil
}

type localProgram struct {
	e      *LocalExecutor
	dir    string
	cmd    []string
	offset int
}

func (p *localProgram) Close() { os.RemoveAll(p.dir) }

func (p *localProgram) Run(input string, limit time.Duration) RunResult {
	p.e.sem <- struct{}{}
	defer func() { <-p.e.sem }()
	ctx, cancel := context.WithTimeout(context.Background(), limit)
	defer cancel()
	cmd := exec.CommandContext(ctx, p.cmd[0], p.cmd[1:]...)
	cmd.Dir, cmd.Env = p.dir, minimalEnv(p.dir)
	cmd.Stdin = strings.NewReader(input)
	cmd.WaitDelay = time.Second
	out, errBuf := &capped{max: 4 << 20}, &capped{max: 16 << 10}
	cmd.Stdout, cmd.Stderr = out, errBuf

	start := time.Now()
	err := cmd.Run()
	r := RunResult{Stdout: out.String(), Millis: int(time.Since(start).Milliseconds())}
	r.Stderr = cleanOutput(errBuf.String(), "", p.offset)
	if ctx.Err() == context.DeadlineExceeded {
		r.TimedOut = true
	} else if err != nil {
		r.ExitErr = true
	}
	return r
}

// minimalEnv keeps secrets such as JWT_SECRET and DB_PASSWORD out of the candidate's process.
func minimalEnv(dir string) []string {
	env := []string{"HOME=" + dir, "TMPDIR=" + dir, "LANG=C.UTF-8"}
	for _, k := range []string{"PATH", "JAVA_HOME", "DEVELOPER_DIR", "SDKROOT"} {
		if v := os.Getenv(k); v != "" {
			env = append(env, k+"="+v)
		}
	}
	return env
}

type capped struct {
	buf bytes.Buffer
	max int
}

func (c *capped) Write(p []byte) (int, error) {
	if room := c.max - c.buf.Len(); room > 0 {
		if len(p) > room {
			p = p[:room]
		}
		c.buf.Write(p)
	}
	return len(p), nil
}

func (c *capped) String() string { return c.buf.String() }

var (
	cppJavaLine = regexp.MustCompile(`(?:Main\.cpp|Main\.java):(\d+)(?::\d+)?`)
	pyLine      = regexp.MustCompile(`File "[^"]*main\.py", line (\d+)`)
)

// cleanOutput hides temp paths and shifts line numbers so they match the candidate's editor.
func cleanOutput(out, dir string, offset int) string {
	if dir != "" {
		out = strings.ReplaceAll(out, dir+"/", "")
		out = strings.ReplaceAll(out, dir, "")
	}
	shift := func(n string) int { v, _ := strconv.Atoi(n); return v - offset }
	out = cppJavaLine.ReplaceAllStringFunc(out, func(m string) string {
		return fmt.Sprintf("line %d", shift(cppJavaLine.FindStringSubmatch(m)[1]))
	})
	out = pyLine.ReplaceAllStringFunc(out, func(m string) string {
		return fmt.Sprintf(`File "solution", line %d`, shift(pyLine.FindStringSubmatch(m)[1]))
	})
	out = strings.TrimSpace(out)
	if len(out) > 1500 {
		out = out[:1500] + "\n..."
	}
	return out
}