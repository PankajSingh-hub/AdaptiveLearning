package coding

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

type Mode int

const (
	ModeRun    Mode = iota // visible examples only; free
	ModeSubmit             // every test plus the large-input test; counts toward hints
)

type Verdict string

const (
	Accepted     Verdict = "ACCEPTED"
	WrongAnswer  Verdict = "WRONG_ANSWER"
	RuntimeError Verdict = "RUNTIME_ERROR"
	TimeLimit    Verdict = "TIME_LIMIT"
	CompileFail  Verdict = "COMPILE_ERROR"
)

type Case struct {
	Tag                    string
	Visible, Passed        bool
	Status                 string // OK, WA, RE, TLE
	Input, Expected, Got   string // filled only for visible examples
	Millis                 int
}

type Outcome struct {
	Verdict       Verdict
	Passed, Total int
	Cases         []Case
	CompileOutput string
	Detail        string // stderr tail for crashes
	StressRan     bool
	StressMillis  int
	MaxMillis     int
}

var timeLimits = map[string]time.Duration{"cpp": 2 * time.Second, "java": 4 * time.Second, "python": 6 * time.Second}

type job struct {
	tag                          string
	visible, stress              bool
	input, expected              string
	display, shownExpected       string
}

func Judge(ex Executor, p Problem, lang, code string, mode Mode) Outcome {
	src, offset := BuildSource(p, lang, code)
	prog, err := ex.Prepare(lang, src, offset)
	if err != nil {
		var ce *CompileError
		if errors.As(err, &ce) {
			return Outcome{Verdict: CompileFail, CompileOutput: ce.Output}
		}
		return Outcome{Verdict: RuntimeError, Detail: "The judge could not run your code: " + err.Error()}
	}
	defer prog.Close()

	jobs := buildJobs(p, mode)
	out := Outcome{Verdict: Accepted, Total: len(jobs)}
	limit := timeLimits[lang]

	for _, j := range jobs {
		res := prog.Run(j.input, limit)
		c := Case{Tag: j.tag, Visible: j.visible, Millis: res.Millis}
		if j.visible {
			c.Input, c.Expected = j.display, j.shownExpected
		}
		switch {
		case res.TimedOut:
			c.Status = "TLE"
		case res.ExitErr:
			c.Status = "RE"
			out.Detail = tail(res.Stderr, 500)
		case !Equal(p, res.Stdout, j.expected):
			c.Status = "WA"
			if j.visible {
				c.Got = clip(strings.TrimSpace(res.Stdout))
			}
		default:
			c.Status, c.Passed = "OK", true
		}
		if j.stress {
			out.StressRan, out.StressMillis = true, res.Millis
		} else if res.Millis > out.MaxMillis {
			out.MaxMillis = res.Millis
		}
		out.Cases = append(out.Cases, c)
		if c.Passed {
			out.Passed++
			continue
		}
		out.Verdict = map[string]Verdict{"WA": WrongAnswer, "RE": RuntimeError, "TLE": TimeLimit}[c.Status]
		if mode == ModeSubmit {
			break // stop at the first failure
		}
	}
	return out
}

func buildJobs(p Problem, mode Mode) []job {
	tests := p.Tests
	if mode == ModeRun {
		tests = p.Tests[:SampleCount]
	}
	jobs := make([]job, 0, len(tests)+1)
	for i, t := range tests {
		exp := p.Ref(t.Args)
		jobs = append(jobs, job{
			tag: t.Tag, visible: i < SampleCount,
			input: EncodeInput(t.Args), expected: FormatOutput(exp),
			display: Render(p, t.Args), shownExpected: RenderOutput(exp),
		})
	}
	if mode == ModeSubmit && p.Stress != nil {
		in, exp := stressData(p)
		jobs = append(jobs, job{tag: "large input", stress: true, input: in, expected: exp})
	}
	return jobs
}

var stressCache sync.Map // slug -> [2]string{input, expected}

func stressData(p Problem) (string, string) {
	if v, ok := stressCache.Load(p.Slug); ok {
		s := v.([2]string)
		return s[0], s[1]
	}
	args := p.Stress()
	in, exp := EncodeInput(args), FormatOutput(p.Ref(args))
	stressCache.Store(p.Slug, [2]string{in, exp})
	return in, exp
}

func clip(s string) string {
	if len(s) > 300 {
		return s[:300] + "..."
	}
	return s
}

func tail(s string, n int) string {
	if len(s) > n {
		return "..." + s[len(s)-n:]
	}
	return s
}

// Analyze explains a failure WITHOUT revealing the solution.
func Analyze(p Problem, o Outcome) []string {
	if o.Verdict == Accepted {
		return nil
	}
	if o.Verdict == CompileFail {
		return []string{"Your code did not compile. Fix the errors below and try again. Compile errors do not use up a hint."}
	}
	var f *Case
	for i := range o.Cases {
		if !o.Cases[i].Passed {
			f = &o.Cases[i]
			break
		}
	}
	if f == nil {
		return []string{o.Detail}
	}
	var notes []string
	switch o.Verdict {
	case WrongAnswer:
		if f.Visible {
			notes = append(notes, "Your output differs from the expected output on a visible example.")
			if f.Got == "" {
				notes = append(notes, "Your function produced no output. Make sure every code path returns a value.")
			}
		} else {
			notes = append(notes, fmt.Sprintf("Your code passed %d earlier test(s) but failed a hidden test that checks: %s.", o.Passed, f.Tag))
		}
	case RuntimeError:
		notes = append(notes, fmt.Sprintf("Your program crashed on the test: %s.", f.Tag),
			"Common causes: index out of range, null or None access, division by zero, or infinite recursion.")
	case TimeLimit:
		if f.Tag == "large input" {
			notes = append(notes, fmt.Sprintf("Your code is correct on small inputs but too slow on a large one (%s).", p.StressDesc),
				fmt.Sprintf("The reference solution runs in %s. Look for a nested loop or repeated work.", p.TimeOpt))
		} else {
			notes = append(notes, fmt.Sprintf("Your code exceeded the time limit on the test: %s.", f.Tag))
		}
	}
	return notes
}

type Feedback struct {
	TimeOptimal, SpaceOptimal, Approach string
	Notes                               []string
	RuntimeMs                           int
}

// BuildFeedback is shown after an accepted submission. Complexity is NOT measured exactly:
// it combines the problem's reference complexity with the result of the large-input test.
func BuildFeedback(p Problem, o Outcome, failures int) Feedback {
	f := Feedback{TimeOptimal: p.TimeOpt, SpaceOptimal: p.SpaceOpt, Approach: p.Approach, RuntimeMs: o.MaxMillis}
	switch {
	case o.StressRan:
		f.Notes = append(f.Notes, fmt.Sprintf("Your solution handled the large input (%s) in %d ms, which is consistent with the optimal %s time.",
			p.StressDesc, o.StressMillis, p.TimeOpt))
		if o.StressMillis > 1000 {
			f.Notes = append(f.Notes, "It passed, but slowly. Extra passes or heavy constant factors may be worth trimming.")
		}
	default:
		f.Notes = append(f.Notes, "Runtime tests cannot tell approaches apart for this problem, so compare your approach with the reference below.")
	}
	if failures > 0 {
		f.Notes = append(f.Notes, fmt.Sprintf("You used %d hint(s). Re-check the edge cases before submitting next time.", failures))
	}
	return f
}