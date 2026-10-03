package main

import (
	"fmt"
	"sync"
)

func generator(nums []int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

func merge(channels ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup

	for _, ch := range channels {
		wg.Add(1)
		go func(c <-chan int) {
			defer wg.Done()
			for v := range c {
				out <- v
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func main() {
	ch1 := generator([]int{1, 2, 3, 4})
	ch2 := generator([]int{10, 20, 30})
	ch3 := generator([]int{100, 200, 300, 400, 500})

	merged := merge(ch1, ch2, ch3)

	result := make([]int, 0)
	for v := range merged {
		result = append(result, v)
	}

	fmt.Println("Собрано чисел:", len(result))
	fmt.Println("Срез:", result)
}