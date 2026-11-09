package coding

import "strings"

func problemsA() []Problem {
	return []Problem{
		{
			Slug: "two-sum", Title: "Two Sum", Difficulty: "EASY", FuncName: "twoSum",
			Description: "Given an array of integers nums and an integer target, return the indices of the two numbers that add up to target.\n\n" +
				"Exactly one valid answer exists, and you may not use the same element twice. You can return the indices in any order.\n\n" +
				"Constraints:\n- 2 <= nums.length <= 200000\n- -10^9 <= nums[i], target <= 10^9",
			Params: []Param{{"nums", "int[]"}, {"target", "int"}}, Ret: "int[]", SortedOutput: true,
			TimeOpt: "O(n)", SpaceOpt: "O(n)", Approach: "One pass with a hash map (value to index), looking up target - nums[i] before storing nums[i].",
			Hints: [5]string{
				"Checking every pair works but is O(n²). Can you avoid re-scanning the array for each element?",
				"For each number x you already know exactly which value you need: target - x.",
				"Which data structure answers 'have I seen this value before?' in O(1) on average?",
				"Walk the array once. Before storing the current number, check whether its complement is already stored.",
				"Use a hash map from value to index. At index i look up target - nums[i]; if present, return both indices, otherwise store nums[i] -> i.",
			},
			Tests: []Test{
				T("sample", []int{2, 7, 11, 15}, 9),
				T("sample", []int{3, 2, 4}, 6),
				T("duplicate values", []int{3, 3}, 6),
				T("negative numbers", []int{-1, -2, -3, -4, -5}, -8),
				T("zeros", []int{0, 4, 3, 0}, 0),
			},
			Stress: func() []any {
				n := 200000
				a := make([]int, n)
				for i := range a {
					a[i] = 4 * i // multiples of 4, so only the two special values below can sum to 2 mod 4
				}
				a[n-2], a[n-1] = 4*(n-2)+1, 4*(n-1)+1
				return []any{a, a[n-2] + a[n-1]}
			},
			StressDesc: "200,000 numbers",
			Ref: func(a []any) any {
				nums, target := a[0].([]int), a[1].(int)
				seen := map[int]int{}
				for i, v := range nums {
					if j, ok := seen[target-v]; ok {
						return []int{j, i}
					}
					seen[v] = i
				}
				return []int{}
			},
		},
		{
			Slug: "valid-parentheses", Title: "Valid Parentheses", Difficulty: "EASY", FuncName: "isValid",
			Description: "Given a string s containing only the characters ( ) [ ] { }, determine whether it is valid.\n\n" +
				"A string is valid if every opening bracket is closed by the same type of bracket, and brackets are closed in the correct order.\n\n" +
				"Constraints:\n- 1 <= s.length <= 100000",
			Params: []Param{{"s", "string"}}, Ret: "bool",
			TimeOpt: "O(n)", SpaceOpt: "O(n)", Approach: "A stack: push openers, and on each closer check the top is its matching opener.",
			Hints: [5]string{
				"Think about which bracket must be closed first when you see an opening bracket.",
				"The most recently opened bracket is always the first one to be closed (last in, first out).",
				"Which data structure gives you last-in-first-out access?",
				"Push opening brackets. On a closing bracket, check that the top of the stack is its matching opener, then pop.",
				"A closing bracket with an empty stack is invalid, and a non-empty stack at the end means something was never closed.",
			},
			Tests: []Test{
				T("sample", "()"),
				T("sample", "()[]{}"),
				T("mismatched types", "(]"),
				T("wrong closing order", "([)]"),
				T("unclosed brackets", "(("),
				T("closer first", "}{"),
			},
			Stress: func() []any {
				return []any{strings.Repeat("(", 50000) + strings.Repeat(")", 50000)}
			},
			StressDesc: "a 100,000-character nested string",
			Ref: func(a []any) any {
				s := a[0].(string)
				pair := map[byte]byte{')': '(', ']': '[', '}': '{'}
				var st []byte
				for i := 0; i < len(s); i++ {
					if o, ok := pair[s[i]]; ok {
						if len(st) == 0 || st[len(st)-1] != o {
							return false
						}
						st = st[:len(st)-1]
					} else {
						st = append(st, s[i])
					}
				}
				return len(st) == 0
			},
		},
		{
			Slug: "best-time-to-buy-and-sell-stock", Title: "Best Time to Buy and Sell Stock", Difficulty: "EASY", FuncName: "maxProfit",
			Description: "prices[i] is the price of a stock on day i. Choose one day to buy and a LATER day to sell, and return the maximum profit you can achieve. If no profit is possible, return 0.\n\n" +
				"Constraints:\n- 1 <= prices.length <= 200000\n- 0 <= prices[i] <= 10000",
			Params: []Param{{"prices", "int[]"}}, Ret: "int",
			TimeOpt: "O(n)", SpaceOpt: "O(1)", Approach: "Scan once, tracking the lowest price so far and the best profit if you sold today.",
			Hints: [5]string{
				"You must buy before you sell. For each day, what would the best buy day before it be?",
				"Checking all pairs works but is O(n²). Can you summarise 'everything before today' in a single number?",
				"Keep track of the lowest price seen so far as you scan left to right.",
				"At each day, the best profit if you sell today is price - lowestSoFar.",
				"Maintain two variables, minPrice and maxProfit, and update both in one pass. Return maxProfit (0 if never profitable).",
			},
			Tests: []Test{
				T("sample", []int{7, 1, 5, 3, 6, 4}),
				T("sample", []int{7, 6, 4, 3, 1}),
				T("single day", []int{5}),
				T("best buy comes before a later dip", []int{2, 4, 1}),
				T("always rising", []int{1, 2, 3, 4, 5}),
			},
			Stress: func() []any {
				r := rng(3)
				a := make([]int, 200000)
				for i := range a {
					a[i] = 1 + r.Intn(10000)
				}
				return []any{a}
			},
			StressDesc: "200,000 prices",
			Ref: func(a []any) any {
				p := a[0].([]int)
				lo, best := p[0], 0
				for _, v := range p {
					if v < lo {
						lo = v
					}
					if v-lo > best {
						best = v - lo
					}
				}
				return best
			},
		},
		{
			Slug: "contains-duplicate", Title: "Contains Duplicate", Difficulty: "EASY", FuncName: "containsDuplicate",
			Description: "Given an integer array nums, return true if any value appears at least twice, and false if every element is distinct.\n\n" +
				"Constraints:\n- 1 <= nums.length <= 200000\n- -10^9 <= nums[i] <= 10^9",
			Params: []Param{{"nums", "int[]"}}, Ret: "bool",
			TimeOpt: "O(n)", SpaceOpt: "O(n)", Approach: "Insert into a hash set while scanning; a repeat means a duplicate. (Sorting and comparing neighbours is O(n log n).)",
			Hints: [5]string{
				"You need to know whether you have seen a value before.",
				"Comparing every pair is O(n²). What lets you remember earlier values with fast lookup?",
				"Think about a set: insertion and membership checks are both O(1) on average.",
				"Scan once. If the current value is already in the set, return true; otherwise add it.",
				"Alternative: sort the array and compare neighbours, which is O(n log n) time with no extra set.",
			},
			Tests: []Test{
				T("sample", []int{1, 2, 3, 1}),
				T("sample", []int{1, 2, 3, 4}),
				T("single element", []int{1}),
				T("duplicate far apart", []int{1, 1, 1, 3, 3, 4, 3, 2, 4, 2}),
				T("negative duplicates", []int{-1, 5, -1}),
			},
			Stress: func() []any {
				a := make([]int, 200000)
				for i := range a {
					a[i] = 3*i + 1
				}
				return []any{a}
			},
			StressDesc: "200,000 distinct numbers",
			Ref: func(a []any) any {
				seen := map[int]bool{}
				for _, v := range a[0].([]int) {
					if seen[v] {
						return true
					}
					seen[v] = true
				}
				return false
			},
		},
		{
			Slug: "maximum-subarray", Title: "Maximum Subarray", Difficulty: "MEDIUM", FuncName: "maxSubArray",
			Description: "Given an integer array nums, find the contiguous subarray (containing at least one number) with the largest sum, and return that sum.\n\n" +
				"Constraints:\n- 1 <= nums.length <= 200000\n- -10000 <= nums[i] <= 10000",
			Params: []Param{{"nums", "int[]"}}, Ret: "int",
			TimeOpt: "O(n)", SpaceOpt: "O(1)", Approach: "Kadane's algorithm: best sum ending here = max(nums[i], previous + nums[i]).",
			Hints: [5]string{
				"A subarray must be contiguous. At each position, decide: extend the previous subarray or start fresh here.",
				"If the best subarray ending at the previous index is negative, it can only hurt you.",
				"Track currentSum, the best sum of a subarray ending at the current index.",
				"currentSum = max(nums[i], currentSum + nums[i]). Keep a separate global maximum.",
				"Initialise both variables with nums[0], not 0, so all-negative arrays work. This is Kadane's algorithm.",
			},
			Tests: []Test{
				T("sample", []int{-2, 1, -3, 4, -1, 2, 1, -5, 4}),
				T("sample", []int{1}),
				T("mostly positive", []int{5, 4, -1, 7, 8}),
				T("all negative numbers", []int{-3, -2, -5}),
				T("single negative", []int{-1}),
			},
			Stress: func() []any {
				r := rng(5)
				a := make([]int, 200000)
				for i := range a {
					a[i] = r.Intn(201) - 100
				}
				return []any{a}
			},
			StressDesc: "200,000 numbers",
			Ref: func(a []any) any {
				n := a[0].([]int)
				cur, best := n[0], n[0]
				for _, v := range n[1:] {
					cur = max(cur+v, v)
					best = max(best, cur)
				}
				return best
			},
		},
		{
			Slug: "product-of-array-except-self", Title: "Product of Array Except Self", Difficulty: "MEDIUM", FuncName: "productExceptSelf",
			Description: "Given an integer array nums, return an array answer where answer[i] is the product of every element of nums EXCEPT nums[i].\n\n" +
				"Do not use division. Every product fits in a 32-bit integer.\n\n" +
				"Constraints:\n- 2 <= nums.length <= 200000",
			Params: []Param{{"nums", "int[]"}}, Ret: "int[]",
			TimeOpt: "O(n)", SpaceOpt: "O(1) extra (output not counted)", Approach: "Prefix products left to right, then multiply by suffix products right to left.",
			Hints: [5]string{
				"Division is tempting, but arrays can contain zeros. Try without it.",
				"The answer at i is (product of everything left of i) × (product of everything right of i).",
				"Compute the left products in one pass and store them in the output array.",
				"Then sweep from the right keeping a running suffix product and multiply it into the output.",
				"Two passes give O(n) time. The output array itself can hold the prefix products, so extra space is O(1).",
			},
			Tests: []Test{
				T("sample", []int{1, 2, 3, 4}),
				T("sample", []int{-1, 1, 0, -3, 3}),
				T("two elements", []int{2, 3}),
				T("multiple zeros", []int{0, 0, 1}),
				T("all ones", []int{1, 1, 1, 1}),
			},
			Stress: func() []any {
				r := rng(6)
				a := make([]int, 200000)
				for i := range a {
					a[i] = r.Intn(2)*2 - 1
				}
				a[len(a)/2] = 0
				return []any{a}
			},
			StressDesc: "200,000 numbers",
			Ref: func(a []any) any {
				n := a[0].([]int)
				out := make([]int, len(n))
				p := 1
				for i, v := range n {
					out[i] = p
					p *= v
				}
				p = 1
				for i := len(n) - 1; i >= 0; i-- {
					out[i] *= p
					p *= n[i]
				}
				return out
			},
		},
		{
			Slug: "valid-anagram", Title: "Valid Anagram", Difficulty: "EASY", FuncName: "isAnagram",
			Description: "Given two lowercase strings s and t, return true if t is an anagram of s (the same letters with the same counts), otherwise false.\n\n" +
				"Constraints:\n- 1 <= s.length, t.length <= 100000\n- s and t contain only lowercase letters a-z",
			Params: []Param{{"s", "string"}, {"t", "string"}}, Ret: "bool",
			TimeOpt: "O(n)", SpaceOpt: "O(1)", Approach: "Count letter frequencies in an array of 26 counters. (Sorting both strings is O(n log n).)",
			Hints: [5]string{
				"Two strings are anagrams if they contain the same letters the same number of times.",
				"If the lengths differ you can answer immediately.",
				"Count how many times each letter appears. There are only 26 lowercase letters.",
				"Increment counts for the characters of s and decrement for the characters of t.",
				"If every count ends at zero the strings are anagrams. (Sorting both strings and comparing also works.)",
			},
			Tests: []Test{
				T("sample", "anagram", "nagaram"),
				T("sample", "rat", "car"),
				T("different lengths", "a", "ab"),
				T("same letters, different counts", "aacc", "ccac"),
				T("longer anagram", "listen", "silent"),
			},
			Ref: func(a []any) any {
				s, t := a[0].(string), a[1].(string)
				if len(s) != len(t) {
					return false
				}
				var c [26]int
				for i := range s {
					c[s[i]-'a']++
					c[t[i]-'a']--
				}
				for _, v := range c {
					if v != 0 {
						return false
					}
				}
				return true
			},
		},
		{
			Slug: "binary-search", Title: "Binary Search", Difficulty: "EASY", FuncName: "search",
			Description: "Given an array nums sorted in ascending order (all values distinct) and an integer target, return the index of target, or -1 if it is not present.\n\n" +
				"Your algorithm must run in O(log n) time.\n\nConstraints:\n- 1 <= nums.length <= 200000",
			Params: []Param{{"nums", "int[]"}, {"target", "int"}}, Ret: "int",
			TimeOpt: "O(log n)", SpaceOpt: "O(1)", Approach: "Halve the search range each step by comparing the middle element with the target.",
			Hints: [5]string{
				"The array is sorted. Can you discard half of the remaining candidates with each comparison?",
				"Look at the middle element and compare it with the target.",
				"If the middle is smaller than the target, the answer can only be to the right; otherwise to the left.",
				"Keep two boundaries (lo and hi) and shrink them until they cross.",
				"Compute mid = lo + (hi - lo) / 2 to avoid overflow, return mid on equality, and return -1 once lo passes hi.",
			},
			Tests: []Test{
				T("sample", []int{-1, 0, 3, 5, 9, 12}, 9),
				T("sample", []int{-1, 0, 3, 5, 9, 12}, 2),
				T("single element, found", []int{5}, 5),
				T("single element, missing", []int{5}, -5),
				T("last element", []int{1, 3}, 3),
			},
			Ref: func(a []any) any {
				n, t := a[0].([]int), a[1].(int)
				lo, hi := 0, len(n)-1
				for lo <= hi {
					m := (lo + hi) / 2
					switch {
					case n[m] == t:
						return m
					case n[m] < t:
						lo = m + 1
					default:
						hi = m - 1
					}
				}
				return -1
			},
		},
		{
			Slug: "climbing-stairs", Title: "Climbing Stairs", Difficulty: "EASY", FuncName: "climbStairs",
			Description: "You are climbing a staircase with n steps. Each time you can climb 1 or 2 steps. Return the number of distinct ways to reach the top.\n\n" +
				"Constraints:\n- 1 <= n <= 45",
			Params: []Param{{"n", "int"}}, Ret: "int",
			TimeOpt: "O(n)", SpaceOpt: "O(1)", Approach: "ways(n) = ways(n-1) + ways(n-2), computed iteratively with two variables.",
			Hints: [5]string{
				"To reach step n, what was your last move? How many ways are there to reach the earlier steps?",
				"You arrive at step n from step n-1 (one step) or from step n-2 (two steps).",
				"So ways(n) = ways(n-1) + ways(n-2). Does that sequence look familiar?",
				"A plain recursion repeats the same work exponentially. Remember results you already computed.",
				"Iterate from 1 up to n keeping only the last two values: O(n) time and O(1) space.",
			},
			Tests: []Test{
				T("sample", 2),
				T("sample", 3),
				T("smallest input", 1),
				T("medium n", 10),
				T("larger n", 30),
			},
			Stress:     func() []any { return []any{45} },
			StressDesc: "n = 45",
			Ref: func(a []any) any {
				n := a[0].(int)
				x, y := 1, 1
				for i := 2; i <= n; i++ {
					x, y = y, x+y
				}
				return y
			},
		},
		{
			Slug: "missing-number", Title: "Missing Number", Difficulty: "EASY", FuncName: "missingNumber",
			Description: "Given an array nums containing n distinct numbers taken from the range 0..n, return the one number in that range that is missing from the array.\n\n" +
				"Constraints:\n- 1 <= nums.length <= 200000",
			Params: []Param{{"nums", "int[]"}}, Ret: "int",
			TimeOpt: "O(n)", SpaceOpt: "O(1)", Approach: "Expected sum n(n+1)/2 minus the actual sum (or XOR all indices and values).",
			Hints: [5]string{
				"The array holds n distinct numbers from 0..n. Exactly one is absent.",
				"What do you know about the sum of 0..n without needing the array?",
				"The expected sum is n(n+1)/2. Compare it with the actual sum.",
				"The difference between the expected sum and the actual sum is the missing number.",
				"Alternative: XOR all indices and all values together; every present number cancels out, leaving the missing one.",
			},
			Tests: []Test{
				T("sample", []int{3, 0, 1}),
				T("sample", []int{0, 1}),
				T("single element, missing 1", []int{0}),
				T("single element, missing 0", []int{1}),
				T("unsorted larger array", []int{9, 6, 4, 2, 3, 5, 7, 0, 1}),
			},
			Stress: func() []any {
				r := rng(10)
				n := 200000
				miss := r.Intn(n + 1)
				a := make([]int, 0, n)
				for v := 0; v <= n; v++ {
					if v != miss {
						a = append(a, v)
					}
				}
				r.Shuffle(len(a), func(i, j int) { a[i], a[j] = a[j], a[i] })
				return []any{a}
			},
			StressDesc: "200,000 numbers",
			Ref: func(a []any) any {
				n := a[0].([]int)
				total := len(n) * (len(n) + 1) / 2
				for _, v := range n {
					total -= v
				}
				return total
			},
		},
	}
}