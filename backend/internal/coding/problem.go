package coding

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
)

// Param types: "int", "int[]", "string". Return types: "int", "bool", "int[]".
type Param struct{ Name, Type string }

type Test struct {
	Tag  string // "sample" for visible examples, otherwise the edge case this test checks
	Args []any
}

type Problem struct {
	Slug, Title, Difficulty, Description string
	FuncName                             string
	Params                               []Param
	Ret                                  string
	SortedOutput                         bool // int[] answers may come back in any order
	TimeOpt, SpaceOpt, Approach          string
	Hints                                [5]string
	Tests                                []Test // the first SampleCount tests are the visible examples
	Stress                               func() []any
	StressDesc                           string
	Ref                                  func(a []any) any // reference solution; expected outputs come from here
}

// Stress is nil when runtime cannot distinguish a good approach from a bad one.

const SampleCount = 2

var Catalog []Problem

func init() { Catalog = append(problemsA(), problemsB()...) }

func Find(slug string) (Problem, bool) {
	for _, p := range Catalog {
		if p.Slug == slug {
			return p, true
		}
	}
	return Problem{}, false
}

func T(tag string, args ...any) Test { return Test{Tag: tag, Args: args} }

func rng(seed int64) *rand.Rand { return rand.New(rand.NewSource(seed)) }

// ---- stdin / stdout wire format ----
// int: one token. string: one token. int[]: "n a1 a2 ... an". Each argument on its own line.

func EncodeInput(args []any) string {
	var b strings.Builder
	for _, a := range args {
		switch v := a.(type) {
		case int:
			b.WriteString(strconv.Itoa(v))
		case string:
			b.WriteString(v)
		case []int:
			b.WriteString(strconv.Itoa(len(v)))
			for _, x := range v {
				b.WriteByte(' ')
				b.WriteString(strconv.Itoa(x))
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func FormatOutput(v any) string {
	switch x := v.(type) {
	case int:
		return strconv.Itoa(x)
	case bool:
		if x {
			return "true"
		}
		return "false"
	case []int:
		parts := make([]string, len(x))
		for i, n := range x {
			parts[i] = strconv.Itoa(n)
		}
		return strings.Join(parts, " ")
	}
	return ""
}

func Equal(p Problem, got, want string) bool {
	g, w := strings.Fields(got), strings.Fields(want)
	if len(g) != len(w) {
		return false
	}
	if p.SortedOutput {
		sortNumeric(g)
		sortNumeric(w)
	}
	for i := range g {
		if g[i] != w[i] {
			return false
		}
	}
	return true
}

func sortNumeric(s []string) {
	sort.Slice(s, func(i, j int) bool {
		a, _ := strconv.Atoi(s[i])
		b, _ := strconv.Atoi(s[j])
		return a < b
	})
}

// ---- human-readable rendering for the examples shown to candidates ----

func Render(p Problem, args []any) string {
	parts := make([]string, len(p.Params))
	for i, pr := range p.Params {
		switch v := args[i].(type) {
		case string:
			parts[i] = fmt.Sprintf("%s = %q", pr.Name, v)
		case []int:
			parts[i] = pr.Name + " = " + bracket(v)
		default:
			parts[i] = fmt.Sprintf("%s = %v", pr.Name, v)
		}
	}
	return strings.Join(parts, ", ")
}

func RenderOutput(v any) string {
	if x, ok := v.([]int); ok {
		return bracket(x)
	}
	return FormatOutput(v)
}

func bracket(v []int) string {
	return "[" + strings.Join(strings.Fields(FormatOutput(v)), ",") + "]"
}

// ---- starter code shown in the editor ----

func Starter(p Problem, lang string) string {
	switch lang {
	case "cpp":
		types := map[string]string{"int": "int", "int[]": "vector<int>&", "string": "string"}
		ret := map[string]string{"int": "int", "bool": "bool", "int[]": "vector<int>"}[p.Ret]
		return fmt.Sprintf("class Solution {\npublic:\n    %s %s(%s) {\n        \n    }\n};\n",
			ret, p.FuncName, paramList(p, types))
	case "java":
		types := map[string]string{"int": "int", "int[]": "int[]", "string": "String"}
		ret := map[string]string{"int": "int", "bool": "boolean", "int[]": "int[]"}[p.Ret]
		return fmt.Sprintf("class Solution {\n    public %s %s(%s) {\n        \n    }\n}\n",
			ret, p.FuncName, paramList(p, types))
	default:
		parts := []string{"self"}
		pyT := map[string]string{"int": "int", "int[]": "List[int]", "string": "str"}
		for _, pr := range p.Params {
			parts = append(parts, pr.Name+": "+pyT[pr.Type])
		}
		ret := map[string]string{"int": "int", "bool": "bool", "int[]": "List[int]"}[p.Ret]
		return fmt.Sprintf("class Solution:\n    def %s(%s) -> %s:\n        pass\n",
			p.FuncName, strings.Join(parts, ", "), ret)
	}
}

func paramList(p Problem, types map[string]string) string {
	parts := make([]string, len(p.Params))
	for i, pr := range p.Params {
		parts[i] = types[pr.Type] + " " + pr.Name
	}
	return strings.Join(parts, ", ")
}