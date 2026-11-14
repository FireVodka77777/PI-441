package main

import (
	"fmt"
	"math"
)

type Mapper interface {
	Map(n int) int
}

type DoubleMapper struct{}

func (d DoubleMapper) Map(n int) int {
	return n * 2
}

type AbsMapper struct{}

func (a AbsMapper) Map(n int) int {
	return int(math.Abs(float64(n)))
}

func Transform(nums []int, m Mapper) []int {
	result := make([]int, len(nums))

	for i, n := range nums {
		result[i] = m.Map(n)
	}
	return result
}

func main() {
	numbers := []int{-5, -2, 0, 3, 7}

	fmt.Println("Исходный массив:", numbers)
	fmt.Println("---------------------------------")

	doubleMapper := DoubleMapper{}
	doubled := Transform(numbers, doubleMapper)
	fmt.Println("Умножить на 2:", doubled)

	absMapper := AbsMapper{}
	abs := Transform(numbers, absMapper)
	fmt.Println("Модуль числа:", abs)
}