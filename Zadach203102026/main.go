package main

import "fmt"

func generate(n int, out chan<- int) {
	for i := 1; i <= n; i++ {
		out <- i
	}
	close(out)
}

func square(in <-chan int, out chan<- int) {
	for num := range in { 
		out <- num * num
	}
	close(out) 
}

func main() {
	const N = 10

	numbers := make(chan int)
	squares := make(chan int)


	go generate(N, numbers)
	go square(numbers, squares)

	result := make([]int, 0, N)
	for sq := range squares {
		result = append(result, sq)
	}

	fmt.Printf("Числа от 1 до %d\n", N)
	fmt.Println("Квадраты:", result)
}