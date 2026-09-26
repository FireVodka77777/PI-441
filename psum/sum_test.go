package main

import "testing"

func sumSeq(nums []int) int {
	s := 0
	for _, v := range nums {
		s += v
	}
	return s
}

func TestParallelSum(t *testing.T) {
	cases := []struct {
		name string
		nums []int
		n    int
		want int
	}{
		{"пустой срез", []int{}, 4, 0},
		{"один элемент", []int{42}, 4, 42},
		{"пять элементов, 2 части", []int{1, 2, 3, 4, 5}, 2, 15},
		{"n = 1", []int{1, 2, 3, 4, 5}, 1, 15},
		{"n = 0", []int{1, 2, 3, 4, 5}, 0, 15},
		{"n отрицательное", []int{1, 2, 3, 4, 5}, -3, 15},
		{"n больше длины", []int{1, 2, 3}, 100, 6},
		{"отрицательные числа", []int{-1, -2, 3, 4}, 2, 4},
		{"все нули", []int{0, 0, 0, 0}, 3, 0},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ParallelSum(c.nums, c.n)
			if got != c.want {
				t.Errorf("ParallelSum(%v, %d) = %d, want %d",
					c.nums, c.n, got, c.want)
			}
		})
	}
}

func TestParallelSumAgainstSequential(t *testing.T) {
	nums := make([]int, 100_000)
	for i := range nums {
		nums[i] = i + 1
	}
	want := sumSeq(nums)

	for _, n := range []int{1, 2, 3, 4, 7, 16, 100} {
		got := ParallelSum(nums, n)
		if got != want {
			t.Errorf("n=%d: got %d, want %d", n, got, want)
		}
	}
}