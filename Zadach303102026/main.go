package main

import (
	"fmt"
	"sync"
)

func worker(id int, jobs <-chan int, results chan<- int, wg *sync.WaitGroup){
	defer wg.Done()
	for num := range jobs {
		results <- num * 2
	}
}

func main(){
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	jobs := make(chan int, len(numbers))
	results := make(chan int, len(numbers))

	const numWorkers = 3
	var wg sync.WaitGroup
	for i := 1; i <= numWorkers; i++{
		wg.Add(1)
		go worker(i, jobs, results, &wg)
	}

	for _, num := range numbers{
		jobs <- num
	}
	close(jobs)
	
	go func(){
		wg.Wait()
		close(results)
	}()
	output := make([]int, 0, len(numbers))
	for res := range results {
		output = append(output, res)
	}

	fmt.Println("Исходный срез:   ", numbers)
	fmt.Println("Результат (x2):  ", output)
}