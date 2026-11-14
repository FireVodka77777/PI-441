package main

import "fmt"

type Limiter interface {
	CanAccept(n int) bool
	Accept(n int)
}

type CountLimiter struct {
	maxCount int
	current  int
}

func NewCountLimiter(max int) *CountLimiter {
	return &CountLimiter{maxCount: max, current: 0}
}

func (l *CountLimiter) CanAccept(n int) bool {
	return l.current < l.maxCount
}

func (l *CountLimiter) Accept(n int) {
	l.current++
}

type SumLimiter struct {
	maxSum  int
	current int
}

func NewSumLimiter(max int) *SumLimiter {
	return &SumLimiter{maxSum: max, current: 0}
}

func (l *SumLimiter) CanAccept(n int) bool {
	return l.current+n <= l.maxSum
}

func (l *SumLimiter) Accept(n int) {
	l.current += n
}

func acceptAll(nums []int, l Limiter) []int {
	result := make([]int, 0)

	for _, n := range nums {
		if l.CanAccept(n) {
			result = append(result, n)
			l.Accept(n)
		}
	}

	return result
}


func main() {
	numbers := []int{5, 10, 15, 20, 25}

	fmt.Println("Исходный массив:", numbers)
	fmt.Println("---------------------------------")

	countLimiter := NewCountLimiter(2)
	byCount := acceptAll(numbers, countLimiter)
	fmt.Println("Не больше 2 чисел:", byCount)

	sumLimiter := NewSumLimiter(30)
	bySum := acceptAll(numbers, sumLimiter)
	fmt.Println("Сумма не больше 30:", bySum)
}
