package coding

import (
	"sort"
	"strconv"
)

func problemsB() []Problem {
	return []Problem{
		{
			Slug: "single-number", Title: "Single Number", Difficulty: "EASY", FuncName: "singleNumber",
			Description: "In an integer array nums, every element appears exactly twice except for one element that appears once. Return that single element.\n\n" +
				"Aim for linear time and constant extra space.\n\nConstraints:\n- 1 <= nums.length <= 200001 (odd length)",
			Params: []Param{{"nums", "int[]"}}, Ret: "int",
			TimeOpt: "O(n)", SpaceOpt: "O(1)", Approach: "XOR every element: pairs cancel to 0 and the single number remains.",
			Hints: [5]string{
				"Every number appears twice except one. What operation cancels out equal values?",
				"A hash map of counts works but needs O(n) extra memory. Can you do it in O(1) extra space?",
				"Think about bitwise XOR: what is a ^ a, and what is a ^ 0?",
				"XOR is commutative and associative, so the order of the numbers does not matter.",
				"XOR every element into one accumulator starting at 0. The pairs vanish and the single number remains.",
			},
			Tests: []Test{
				T("sample", []int{2, 2, 1}),
				T("sample", []int{4, 1, 2, 1, 2}),
				T("single element", []int{1}),
				T("negative single", []int{-1, -1, -2}),
				T("single is zero", []int{1, 0, 1}),
			},
			Stress: func() []any {
				r := rng(11)
				a := make([]int, 0, 200001)
				for v := 1; v <= 100000; v++ {
					a = append(a, v, v)
				}
				a = append(a, 5000000)
				r.Shuffle(len(a), func(i, j int) { a[i], a[j] = a[j], a[i] })
				return []any{a}
			},
			StressDesc: "200,001 numbers",
			Ref: func(a []any) any {
				x := 0
				for _, v := range a[0].([]int) {
					x ^= v
				}
				return x
			},
		},
		{
			Slug: "longest-substring-without-repeating-characters", Title: "Longest Substring Without Repeating Characters", Difficulty: "MEDIUM", FuncName: "lengthOfLongestSubstring",
			Description: "Given a string s, return the length of the longest substring that contains no repeated characters.\n\n" +
				"Constraints:\n- 1 <= s.length <= 200000\n- s contains letters and digits only",
			Params: []Param{{"s", "string"}}, Ret: "int",
			TimeOpt: "O(n)", SpaceOpt: "O(k) for k distinct characters", Approach: "Sliding window with a map of each character's last index; move the left edge past a repeat, never backwards.",
			Hints: [5]string{
				"Think of a window over the string that must never contain a repeated character.",
				"Checking every substring is O(n²) or worse. A window that only moves forward is faster.",
				"Keep the left edge of the window and extend the right edge one character at a time.",
				"When the new character is already inside the window, move the left edge just past its previous occurrence.",
				"Store each character's last index in a map. Careful: the left edge must never move backwards (see \"abba\").",
			},
			Tests: []Test{
				T("sample", "abcabcbb"),
				T("sample", "bbbbb"),
				T("repeat inside the window", "pwwkew"),
				T("window must not move backwards", "abba"),
				T("single character", "a"),
				T("overlapping repeats", "dvdf"),
			},
			Ref: func(a []any) any {
				s := a[0].(string)
				last := map[byte]int{}
				best, l := 0, 0
				for r := 0; r < len(s); r++ {
					if i, ok := last[s[r]]; ok && i >= l {
						l = i + 1
					}
					last[s[r]] = r
					best = max(best, r-l+1)
				}
				return best
			},
		},
		{
			Slug: "container-with-most-water", Title: "Container With Most Water", Difficulty: "MEDIUM", FuncName: "maxArea",
			Description: "height[i] is the height of a vertical line at position i. Choose two lines that, together with the x-axis, form a container, and return the maximum amount of water it can hold.\n\n" +
				"The area is the distance between the lines times the shorter height.\n\nConstraints:\n- 2 <= height.length <= 200000\n- 0 <= height[i] <= 1000",
			Params: []Param{{"height", "int[]"}}, Ret: "int",
			TimeOpt: "O(n)", SpaceOpt: "O(1)", Approach: "Two pointers from both ends, always moving the pointer at the shorter line inward.",
			Hints: [5]string{
				"The area is limited by the shorter of the two lines times their distance.",
				"Trying every pair is O(n²). Start with the widest container and think about how to improve it.",
				"Use two pointers at both ends. Which pointer should move inward?",
				"Moving the taller line can never help, because the shorter line still limits the height while the width shrinks.",
				"Always move the pointer at the shorter line inward, tracking the maximum area. O(n) time, O(1) space.",
			},
			Tests: []Test{
				T("sample", []int{1, 8, 6, 2, 5, 4, 8, 3, 7}),
				T("sample", []int{1, 1}),
				T("tall lines at the ends", []int{4, 3, 2, 1, 4}),
				T("tall line in the middle", []int{1, 2, 1}),
				T("best pair is not at the ends", []int{2, 3, 4, 5, 18, 17, 6}),
			},
			Stress: func() []any {
				r := rng(13)
				a := make([]int, 200000)
				for i := range a {
					a[i] = 1 + r.Intn(1000)
				}
				return []any{a}
			},
			StressDesc: "200,000 lines",
			Ref: func(a []any) any {
				h := a[0].([]int)
				l, r, best := 0, len(h)-1, 0
				for l < r {
					best = max(best, min(h[l], h[r])*(r-l))
					if h[l] < h[r] {
						l++
					} else {
						r--
					}
				}
				return best
			},
		},
		{
			Slug: "coin-change", Title: "Coin Change", Difficulty: "MEDIUM", FuncName: "coinChange",
			Description: "Given an array coins of coin denominations and an integer amount, return the FEWEST number of coins needed to make that amount. If it cannot be made, return -1. You may use each coin any number of times.\n\n" +
				"Constraints:\n- 1 <= coins.length <= 12\n- 1 <= coins[i] <= 10000\n- 0 <= amount <= 10000",
			Params: []Param{{"coins", "int[]"}, {"amount", "int"}}, Ret: "int",
			TimeOpt: "O(amount × coins)", SpaceOpt: "O(amount)", Approach: "Bottom-up DP: dp[a] = 1 + min(dp[a - coin]) over all coins that fit.",
			Hints: [5]string{
				"A greedy 'largest coin first' strategy fails for some coin sets (try coins [1,3,4] and amount 6).",
				"Define the answer for every smaller amount, building up toward the target.",
				"Let dp[a] be the fewest coins needed to make amount a. What is dp[0]?",
				"dp[a] = 1 + min(dp[a - coin]) over all coins that fit, ignoring unreachable amounts.",
				"Initialise dp with a large value (infinity) except dp[0] = 0. Return -1 if dp[amount] is still infinity.",
			},
			Tests: []Test{
				T("sample", []int{1, 2, 5}, 11),
				T("sample", []int{2}, 3),
				T("amount is zero", []int{1}, 0),
				T("greedy fails", []int{1, 3, 4}, 6),
				T("larger amount", []int{186, 419, 83, 408}, 6249),
			},
			Stress: func() []any {
				return []any{[]int{3, 7, 13, 29, 31, 47, 61, 83, 101, 199}, 10000}
			},
			StressDesc: "amount = 10,000",
			Ref: func(a []any) any {
				c, amt := a[0].([]int), a[1].(int)
				inf := amt + 1
				dp := make([]int, amt+1)
				for i := 1; i <= amt; i++ {
					dp[i] = inf
					for _, k := range c {
						if k <= i && dp[i-k]+1 < dp[i] {
							dp[i] = dp[i-k] + 1
						}
					}
				}
				if dp[amt] >= inf {
					return -1
				}
				return dp[amt]
			},
		},
		{
			Slug: "house-robber", Title: "House Robber", Difficulty: "MEDIUM", FuncName: "rob",
			Description: "nums[i] is the money in house i along a street. You cannot rob two adjacent houses. Return the maximum amount you can rob.\n\n" +
				"Constraints:\n- 1 <= nums.length <= 100000\n- 0 <= nums[i] <= 400",
			Params: []Param{{"nums", "int[]"}}, Ret: "int",
			TimeOpt: "O(n)", SpaceOpt: "O(1)", Approach: "DP: best(i) = max(best(i-1), best(i-2) + nums[i]), keeping only two variables.",
			Hints: [5]string{
				"At each house you either rob it or skip it. What does robbing it forbid?",
				"Let best(i) be the most money from the first i houses.",
				"best(i) = max(best(i-1), best(i-2) + nums[i]).",
				"Plain recursion recomputes the same subproblems. Store results or iterate.",
				"You only need the last two values, so one loop with two variables gives O(n) time and O(1) space.",
			},
			Tests: []Test{
				T("sample", []int{1, 2, 3, 1}),
				T("sample", []int{2, 7, 9, 3, 1}),
				T("single house", []int{5}),
				T("two houses", []int{2, 1}),
				T("greedy fails", []int{2, 1, 1, 2}),
			},
			Stress: func() []any {
				r := rng(15)
				a := make([]int, 100000)
				for i := range a {
					a[i] = r.Intn(401)
				}
				return []any{a}
			},
			StressDesc: "100,000 houses",
			Ref: func(a []any) any {
				p, q := 0, 0
				for _, v := range a[0].([]int) {
					p, q = q, max(q, p+v)
				}
				return q
			},
		},
		{
			Slug: "palindrome-number", Title: "Palindrome Number", Difficulty: "EASY", FuncName: "isPalindrome",
			Description: "Given an integer x, return true if x reads the same forwards and backwards, otherwise false.\n\n" +
				"Constraints:\n- -2^31 <= x <= 2^31 - 1",
			Params: []Param{{"x", "int"}}, Ret: "bool",
			TimeOpt: "O(log10 x)", SpaceOpt: "O(1)", Approach: "Reverse half of the digits arithmetically and compare with the remaining half.",
			Hints: [5]string{
				"Negative numbers can never be palindromes, since the minus sign is only on one side.",
				"Converting to a string works. Can you do it using only arithmetic?",
				"Use % 10 to take the last digit and / 10 to drop it.",
				"Build the reversed number digit by digit (watch for overflow on large inputs) and compare it with the original.",
				"Or reverse only half the digits: stop when the reversed half is >= the remaining half. Numbers ending in 0 (except 0 itself) are not palindromes.",
			},
			Tests: []Test{
				T("sample", 121),
				T("sample", -121),
				T("ends with zero", 10),
				T("zero", 0),
				T("odd number of digits", 12321),
				T("almost a palindrome", 1000021),
				T("near the 32-bit limit", 2147447412),
			},
			Ref: func(a []any) any {
				x := a[0].(int)
				if x < 0 {
					return false
				}
				s := strconv.Itoa(x)
				for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
					if s[i] != s[j] {
						return false
					}
				}
				return true
			},
		},
		{
			Slug: "search-in-rotated-sorted-array", Title: "Search in Rotated Sorted Array", Difficulty: "MEDIUM", FuncName: "search",
			Description: "An ascending array of distinct integers was rotated at an unknown pivot (for example [0,1,2,4,5,6,7] became [4,5,6,7,0,1,2]). Given the rotated array nums and an integer target, return the index of target, or -1 if absent.\n\n" +
				"Your algorithm must run in O(log n) time.\n\nConstraints:\n- 1 <= nums.length <= 200000",
			Params: []Param{{"nums", "int[]"}, {"target", "int"}}, Ret: "int",
			TimeOpt: "O(log n)", SpaceOpt: "O(1)", Approach: "Modified binary search: at each step one half is sorted; check whether the target lies in that half.",
			Hints: [5]string{
				"A rotated sorted array is still made of two sorted pieces.",
				"At any midpoint, at least one half of the current range is properly sorted.",
				"Work out which half is sorted by comparing nums[lo] with nums[mid].",
				"If the target lies within the sorted half's range, search that half; otherwise search the other half.",
				"This is binary search with one extra check per step: O(log n). Mind the inclusive/exclusive boundaries.",
			},
			Tests: []Test{
				T("sample", []int{4, 5, 6, 7, 0, 1, 2}, 0),
				T("sample", []int{4, 5, 6, 7, 0, 1, 2}, 3),
				T("single element, missing", []int{1}, 0),
				T("two elements, rotated", []int{3, 1}, 1),
				T("target at the front", []int{5, 1, 3}, 5),
			},
			Ref: func(a []any) any {
				n, t := a[0].([]int), a[1].(int)
				for i, v := range n {
					if v == t {
						return i
					}
				}
				return -1
			},
		},
		{
			Slug: "find-minimum-in-rotated-sorted-array", Title: "Find Minimum in Rotated Sorted Array", Difficulty: "MEDIUM", FuncName: "findMin",
			Description: "A sorted array of distinct integers may have been rotated. Return the minimum element.\n\n" +
				"Your algorithm must run in O(log n) time.\n\nConstraints:\n- 1 <= nums.length <= 200000",
			Params: []Param{{"nums", "int[]"}}, Ret: "int",
			TimeOpt: "O(log n)", SpaceOpt: "O(1)", Approach: "Binary search comparing nums[mid] with nums[hi] to find the rotation point.",
			Hints: [5]string{
				"Scanning everything is O(n). The array is sorted but rotated; can you still halve the search?",
				"The minimum is the break point where the order drops.",
				"Compare nums[mid] with nums[hi] to decide which side holds the break point.",
				"If nums[mid] > nums[hi], the minimum is to the right of mid; otherwise it is at mid or to its left.",
				"Loop while lo < hi, setting hi = mid or lo = mid + 1; the answer is nums[lo]. A non-rotated array must also work.",
			},
			Tests: []Test{
				T("sample", []int{3, 4, 5, 1, 2}),
				T("sample", []int{4, 5, 6, 7, 0, 1, 2}),
				T("not rotated", []int{11, 13, 15, 17}),
				T("two elements", []int{2, 1}),
				T("single element", []int{1}),
			},
			Ref: func(a []any) any {
				n := a[0].([]int)
				m := n[0]
				for _, v := range n {
					m = min(m, v)
				}
				return m
			},
		},
		{
			Slug: "trapping-rain-water", Title: "Trapping Rain Water", Difficulty: "HARD", FuncName: "trap",
			Description: "height[i] is the height of an elevation bar of width 1. Compute how much rain water is trapped between the bars after it rains.\n\n" +
				"Constraints:\n- 1 <= height.length <= 200000\n- 0 <= height[i] <= 1000",
			Params: []Param{{"height", "int[]"}}, Ret: "int",
			TimeOpt: "O(n)", SpaceOpt: "O(1)", Approach: "Two pointers with running left/right maxima, always advancing the side with the smaller maximum.",
			Hints: [5]string{
				"Water above a bar depends on the tallest bars on both its left and its right.",
				"For bar i: water = min(maxLeft, maxRight) - height[i] (never negative).",
				"Computing maxLeft and maxRight for every bar from scratch is O(n²). Can you precompute them?",
				"Two arrays (prefix max and suffix max) give O(n) time with O(n) space.",
				"Two pointers from both ends with running leftMax/rightMax give O(n) time and O(1) space: process the side with the smaller max.",
			},
			Tests: []Test{
				T("sample", []int{0, 1, 0, 2, 1, 0, 1, 3, 2, 1, 2, 1}),
				T("sample", []int{4, 2, 0, 3, 2, 5}),
				T("single bar", []int{1}),
				T("a single valley", []int{3, 0, 3}),
				T("descending bars", []int{5, 4, 3, 2, 1}),
			},
			Stress: func() []any {
				r := rng(19)
				a := make([]int, 200000)
				for i := range a {
					a[i] = r.Intn(1001)
				}
				return []any{a}
			},
			StressDesc: "200,000 bars",
			Ref: func(a []any) any {
				h := a[0].([]int)
				l, r, lm, rm, w := 0, len(h)-1, 0, 0, 0
				for l < r {
					if h[l] < h[r] {
						lm = max(lm, h[l])
						w += lm - h[l]
						l++
					} else {
						rm = max(rm, h[r])
						w += rm - h[r]
						r--
					}
				}
				return w
			},
		},
		{
			Slug: "longest-increasing-subsequence", Title: "Longest Increasing Subsequence", Difficulty: "MEDIUM", FuncName: "lengthOfLIS",
			Description: "Given an integer array nums, return the length of the longest STRICTLY increasing subsequence. A subsequence keeps the original order but elements need not be adjacent.\n\n" +
				"Constraints:\n- 1 <= nums.length <= 200000\n- 0 <= nums[i] <= 10^9",
			Params: []Param{{"nums", "int[]"}}, Ret: "int",
			TimeOpt: "O(n log n)", SpaceOpt: "O(n)", Approach: "Maintain a 'tails' array and binary-search the first tail >= each number, replacing it or appending.",
			Hints: [5]string{
				"Subsequence elements keep their order but need not be adjacent, and the sequence must be strictly increasing.",
				"A DP of 'longest increasing subsequence ending at i' works in O(n²). It can be improved.",
				"Keep an array tails where tails[k] is the smallest possible ending value of an increasing subsequence of length k+1.",
				"For each number, find the first tail that is >= it with binary search and replace it; if none exists, append it.",
				"The final length of tails is the answer, in O(n log n). Use a lower-bound (>=) search so equal values do not extend the sequence.",
			},
			Tests: []Test{
				T("sample", []int{10, 9, 2, 5, 3, 7, 101, 18}),
				T("sample", []int{0, 1, 0, 3, 2, 3}),
				T("strictly increasing needed", []int{7, 7, 7, 7}),
				T("single element", []int{1}),
				T("answer skips elements", []int{4, 10, 4, 3, 8, 9}),
			},
			Stress: func() []any {
				r := rng(20)
				a := make([]int, 200000)
				for i := range a {
					a[i] = r.Intn(1000000000)
				}
				return []any{a}
			},
			StressDesc: "200,000 numbers",
			Ref: func(a []any) any {
				var tails []int
				for _, v := range a[0].([]int) {
					i := sort.SearchInts(tails, v)
					if i == len(tails) {
						tails = append(tails, v)
					} else {
						tails[i] = v
					}
				}
				return len(tails)
			},
		},
	}
}