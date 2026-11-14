package main

import (
	"fmt"
	"strconv"
)

type Formatter interface {
	Format(n int) string
}

type PlainFormatter struct{}
func (f PlainFormatter) Format(n int) string {
	return strconv.Itoa(n)
}

type BracketFormatter struct{}
func (f BracketFormatter) Format(n int) string {
	return "[" + strconv.Itoa(n) + "]"
}

type EvenOddFormatter struct{}
func (f EvenOddFormatter) Format(n int) string {
	if n%2 == 0 {
		return strconv.Itoa(n) + " (чёт)"
	}
	return strconv.Itoa(n) + " (нечёт)"
}

func printAll(nums []int, f Formatter) {
	for _, n := range nums {
		fmt.Println(f.Format(n))
	}
}

func main() {
	numbers := []int{1, 2, 3, 4, 5}

	fmt.Println("--- Просто число ---")
	printAll(numbers, PlainFormatter{})
	fmt.Println("--- В квадратных скобках ---")
	printAll(numbers, BracketFormatter{})
	fmt.Println("--- Чёт/нечёт ---")
	printAll(numbers, EvenOddFormatter{})
}