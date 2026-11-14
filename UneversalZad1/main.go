package main

import "fmt"

type Predicate interface {
	Check(n int) bool
}

type EvenFilter struct{}

func (e EvenFilter) Check(n int) bool {
	if n%2 == 0 {
		return true
	}
	return false
}

type PositiveFilter struct{}

func (p PositiveFilter) Check(n int) bool {
	if n > 0 {
		return true
	}
	return false
}

func Filter(nums []int, p Predicate) []int {
	result := make([]int, 0)

	for _, n := range nums {
		if p.Check(n) {
			result = append(result, n)
		}
	}
	return result
}

func main() {

	numbers := []int{-4, -3, -2, -1, 0, 1, 2, 3, 4, 5}

	fmt.Println("Исходный массив:", numbers)
	fmt.Println("---------------------------------")

	evenFilter := EvenFilter{}
	evens := Filter(numbers, evenFilter)
	fmt.Println("Только чётные:", evens)

	positiveFilter := PositiveFilter{}
	positives := Filter(numbers, positiveFilter)
	fmt.Println("Только положительные:", positives)
}