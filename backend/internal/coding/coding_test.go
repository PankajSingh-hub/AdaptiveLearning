package coding

import (
	"strings"
	"testing"
)

func TestCatalogIsConsistent(t *testing.T) {
	if len(Catalog) != 20 {
		t.Fatalf("expected 20 problems, got %d", len(Catalog))
	}
	seen := map[string]bool{}
	for _, p := range Catalog {
		if seen[p.Slug] {
			t.Fatalf("duplicate slug %s", p.Slug)
		}
		seen[p.Slug] = true
		if len(p.Tests) < SampleCount+2 {
			t.Errorf("%s: needs at least %d tests", p.Slug, SampleCount+2)
		}
		for i, h := range p.Hints {
			if strings.TrimSpace(h) == "" {
				t.Errorf("%s: hint %d is empty", p.Slug, i+1)
			}
		}
		for _, tc := range p.Tests {
			if len(tc.Args) != len(p.Params) {
				t.Errorf("%s: test %q has %d args, want %d", p.Slug, tc.Tag, len(tc.Args), len(p.Params))
			}
			p.Ref(tc.Args) // must not panic
		}
		for _, lang := range []string{"cpp", "java", "python"} {
			starter := Starter(p, lang)
			if !strings.Contains(starter, p.FuncName) {
				t.Errorf("%s/%s: starter lacks function name", p.Slug, lang)
			}
			src, _ := BuildSource(p, lang, starter)
			if !strings.Contains(src, p.FuncName+"(") {
				t.Errorf("%s/%s: driver does not call the function", p.Slug, lang)
			}
		}
	}
}

func executorWith(t *testing.T, lang string) Executor {
	t.Helper()
	ex := NewLocalExecutor(1)
	for _, l := range ex.Languages() {
		if l.ID == lang {
			return ex
		}
	}
	t.Skipf("%s is not installed", lang)
	return nil
}

const twoSumPy = "class Solution:\n    def twoSum(self, nums, target):\n        seen = {}\n" +
	"        for i, v in enumerate(nums):\n            if target - v in seen:\n                return [seen[target - v], i]\n            seen[v] = i\n"

func TestJudgeAcceptsCorrectPython(t *testing.T) {
	p, _ := Find("two-sum")
	out := Judge(executorWith(t, "python"), p, "python", twoSumPy, ModeSubmit)
	if out.Verdict != Accepted {
		t.Fatalf("want ACCEPTED, got %s: %+v", out.Verdict, out)
	}
	if !out.StressRan {
		t.Fatal("stress test should have run")
	}
}

func TestJudgeRejectsWrongAnswer(t *testing.T) {
	p, _ := Find("two-sum")
	code := "class Solution:\n    def twoSum(self, nums, target):\n        return [0, 0]\n"
	out := Judge(executorWith(t, "python"), p, "python", code, ModeSubmit)
	if out.Verdict != WrongAnswer {
		t.Fatalf("want WRONG_ANSWER, got %s", out.Verdict)
	}
}

func TestJudgeReportsSyntaxErrorAsCompileError(t *testing.T) {
	p, _ := Find("two-sum")
	out := Judge(executorWith(t, "python"), p, "python", "class Solution:\n    def twoSum(self, nums target):\n        pass\n", ModeRun)
	if out.Verdict != CompileFail {
		t.Fatalf("want COMPILE_ERROR, got %s", out.Verdict)
	}
}

func TestJudgeReportsCrash(t *testing.T) {
	p, _ := Find("two-sum")
	code := "class Solution:\n    def twoSum(self, nums, target):\n        return nums[999999]\n"
	out := Judge(executorWith(t, "python"), p, "python", code, ModeRun)
	if out.Verdict != RuntimeError {
		t.Fatalf("want RUNTIME_ERROR, got %s", out.Verdict)
	}
}