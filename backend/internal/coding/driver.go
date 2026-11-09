package coding

import (
	"fmt"
	"strings"
)

// The candidate's code is placed between a prelude (imports) and a generated main()
// that reads the test input from stdin, calls the function and prints the result.

const cppPrelude = "#include <iostream>\n#include <vector>\n#include <string>\n#include <algorithm>\n" +
	"#include <unordered_map>\n#include <unordered_set>\n#include <map>\n#include <set>\n#include <queue>\n" +
	"#include <stack>\n#include <deque>\n#include <cmath>\n#include <climits>\n#include <numeric>\n" +
	"#include <cstring>\n#include <functional>\nusing namespace std;\n"

const javaPrelude = "import java.util.*;\n"

const pyPrelude = "from typing import List\nimport sys, os, threading, traceback\n"

// BuildSource returns the full program and how many prelude lines were added,
// so error messages can be shown with the candidate's own line numbers.
func BuildSource(p Problem, lang, user string) (string, int) {
	var pre, main string
	switch lang {
	case "cpp":
		pre, main = cppPrelude, cppMain(p)
	case "java":
		pre, main = javaPrelude, javaMain(p)
	default:
		pre, main = pyPrelude, pyMain(p)
	}
	return pre + user + "\n" + main, strings.Count(pre, "\n")
}

func names(p Problem) string {
	n := make([]string, len(p.Params))
	for i, pr := range p.Params {
		n[i] = pr.Name
	}
	return strings.Join(n, ", ")
}

func cppMain(p Problem) string {
	var b strings.Builder
	b.WriteString("\nint main() {\n    ios::sync_with_stdio(false);\n    cin.tie(nullptr);\n")
	for _, pr := range p.Params {
		n := pr.Name
		switch pr.Type {
		case "int":
			fmt.Fprintf(&b, "    int %s; cin >> %s;\n", n, n)
		case "string":
			fmt.Fprintf(&b, "    string %s; cin >> %s;\n", n, n)
		case "int[]":
			fmt.Fprintf(&b, "    int n_%[1]s; cin >> n_%[1]s; vector<int> %[1]s(n_%[1]s); for (auto& v_ : %[1]s) cin >> v_;\n", n)
		}
	}
	fmt.Fprintf(&b, "    Solution sol_;\n    auto res_ = sol_.%s(%s);\n", p.FuncName, names(p))
	switch p.Ret {
	case "int":
		b.WriteString("    cout << res_ << \"\\n\";\n")
	case "bool":
		b.WriteString("    cout << (res_ ? \"true\" : \"false\") << \"\\n\";\n")
	case "int[]":
		b.WriteString("    for (size_t i_ = 0; i_ < res_.size(); i_++) { if (i_) cout << ' '; cout << res_[i_]; }\n    cout << \"\\n\";\n")
	}
	b.WriteString("    return 0;\n}\n")
	return b.String()
}

func javaMain(p Problem) string {
	var b strings.Builder
	b.WriteString("\npublic class Main {\n    public static void main(String[] argv_) throws Exception {\n")
	b.WriteString(`        String[] tk_ = new String(System.in.readAllBytes()).trim().split("\\s+");` + "\n        int p_ = 0;\n")
	for _, pr := range p.Params {
		n := pr.Name
		switch pr.Type {
		case "int":
			fmt.Fprintf(&b, "        int %s = Integer.parseInt(tk_[p_++]);\n", n)
		case "string":
			fmt.Fprintf(&b, "        String %s = tk_[p_++];\n", n)
		case "int[]":
			fmt.Fprintf(&b, "        int n_%[1]s = Integer.parseInt(tk_[p_++]);\n        int[] %[1]s = new int[n_%[1]s];\n"+
				"        for (int i_ = 0; i_ < n_%[1]s; i_++) %[1]s[i_] = Integer.parseInt(tk_[p_++]);\n", n)
		}
	}
	jt := map[string]string{"int": "int", "bool": "boolean", "int[]": "int[]"}[p.Ret]
	fmt.Fprintf(&b, "        %s res_ = new Solution().%s(%s);\n", jt, p.FuncName, names(p))
	switch p.Ret {
	case "int":
		b.WriteString("        System.out.println(res_);\n")
	case "bool":
		b.WriteString("        System.out.println(res_ ? \"true\" : \"false\");\n")
	case "int[]":
		b.WriteString("        StringBuilder sb_ = new StringBuilder();\n" +
			"        for (int i_ = 0; i_ < res_.length; i_++) { if (i_ > 0) sb_.append(' '); sb_.append(res_[i_]); }\n" +
			"        System.out.println(sb_);\n")
	}
	b.WriteString("    }\n}\n")
	return b.String()
}

func pyMain(p Problem) string {
	var b strings.Builder
	b.WriteString("\ndef run_():\n    try:\n        tk_ = sys.stdin.read().split()\n        p_ = 0\n")
	for _, pr := range p.Params {
		n := pr.Name
		switch pr.Type {
		case "int":
			fmt.Fprintf(&b, "        %s = int(tk_[p_])\n        p_ += 1\n", n)
		case "string":
			fmt.Fprintf(&b, "        %s = tk_[p_]\n        p_ += 1\n", n)
		case "int[]":
			fmt.Fprintf(&b, "        n_%[1]s = int(tk_[p_])\n        p_ += 1\n"+
				"        %[1]s = [int(v_) for v_ in tk_[p_:p_ + n_%[1]s]]\n        p_ += n_%[1]s\n", n)
		}
	}
	fmt.Fprintf(&b, "        res_ = Solution().%s(%s)\n", p.FuncName, names(p))
	switch p.Ret {
	case "int":
		b.WriteString("        print(res_)\n")
	case "bool":
		b.WriteString("        print(\"true\" if res_ else \"false\")\n")
	case "int[]":
		b.WriteString("        print(\" \".join(str(v_) for v_ in res_))\n")
	}
	// A big thread stack lets deep (memoised) recursion work; os._exit makes crashes visible as a failed exit code.
	b.WriteString("    except BaseException:\n        traceback.print_exc()\n        sys.stdout.flush()\n        os._exit(1)\n")
	b.WriteString("\nsys.setrecursionlimit(1000000)\nthreading.stack_size(256 * 1024 * 1024)\n" +
		"th_ = threading.Thread(target=run_)\nth_.start()\nth_.join()\n")
	return b.String()
}