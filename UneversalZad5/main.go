package main

import "fmt"

type Aggregator interface {
	Add(n int)
	Finish()
	Result() string
}

type SumCountAggregator struct {
	sum   int
	count int
}

func (a *SumCountAggregator) Add(n int) {
	a.sum += n
	a.count++
}

func (a *SumCountAggregator) Finish() {
}

func (a *SumCountAggregator) Result() string {
	return fmt.Sprintf("sum=%d, count=%d", a.sum, a.count)
}

type MinMaxAggregator struct {
	min     int
	max     int
	started bool 
}

func (a *MinMaxAggregator) Add(n int) {
	if !a.started {
		a.min = n
		a.max = n
		a.started = true
		return
	}
	if n < a.min {
		a.min = n
	}
	if n > a.max {
		a.max = n
	}
}

func (a *MinMaxAggregator) Finish() {
}

func (a *MinMaxAggregator) Result() string {
	if !a.started {
		return "min=n/a, max=n/a"
	}
	return fmt.Sprintf("min=%d, max=%d", a.min, a.max)
}

func aggregate(nums []int, a Aggregator) string {
	for _, n := range nums {
		a.Add(n)
	}
	a.Finish()
	return a.Result()
}

func main() {

	numbers := []int{-5, 3, -2, 8, 0}

	fmt.Println("Массив с отрицательными числами:", numbers)
	sumCount := &SumCountAggregator{}
	fmt.Println("Сумма и количество:", aggregate(numbers, sumCount))
	minMax := &MinMaxAggregator{}
	fmt.Println("Минимум и максимум:", aggregate(numbers, minMax))

	fmt.Println("---------------------------------")

	empty := []int{}
	fmt.Println("Пустой массив:", empty)
	sumCountEmpty := &SumCountAggregator{}
	fmt.Println("Сумма и количество:", aggregate(empty, sumCountEmpty))
	minMaxEmpty := &MinMaxAggregator{}
	fmt.Println("Минимум и максимум:", aggregate(empty, minMaxEmpty))
}