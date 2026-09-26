package main

import (
	"fmt"
	"sync"
)

func ParallelSum(nums []int, n int) int {
	if len(nums) == 0 {
		return 0
	}
	if n <= 0 {
		n = 1
	}
	if n > len(nums) {
		n = len(nums)
	}

	chunkSize := (len(nums) + n - 1) / n

	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		totalSum int
	)

	for i := 0; i < len(nums); i += chunkSize {
		end := i + chunkSize
		if end > len(nums) {
			end = len(nums)
		}

		wg.Add(1)
		go func(part []int) {
			defer wg.Done()

			local := 0
			for _, v := range part {
				local += v
			}

			mu.Lock()
			totalSum += local
			mu.Unlock()
		}(nums[i:end])
	}

	wg.Wait()
	return totalSum
}

func main() {
	nums := make([]int, 1_000_000)
	for i := range nums {
		nums[i] = i + 1
	}

	const parts = 8
	sum := ParallelSum(nums, parts)
	expected := len(nums) * (len(nums) + 1) / 2

	fmt.Printf("Частей:    %d\n", parts)
	fmt.Printf("Сумма:     %d\n", sum)
	fmt.Printf("Ожидалось: %d\n", expected)
	fmt.Printf("Совпадает: %v\n", sum == expected)
}