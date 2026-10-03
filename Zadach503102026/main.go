package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func process(ctx context.Context, num int) (int, error) {
	select {
	case <-time.After(300 * time.Millisecond): 
		return num * 5, nil
	case <-ctx.Done(): 
		return 0, ctx.Err()
	}
}
func worker(
	ctx context.Context,
	cancel context.CancelFunc,
	id int,
	jobs <-chan int,
	results chan<- int,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("[worker %d] получил отмену, выхожу\n", id)
			return
		case num, ok := <-jobs:
			if !ok {
				fmt.Printf("[worker %d] канал закрыт, выхожу\n", id)
				return
			}

			fmt.Printf("[worker %d] взял в работу: %d\n", id, num)
			res, err := process(ctx, num)
			if err != nil {
				fmt.Printf("[worker %d] обработка %d прервана: %v\n", id, num, err)
				return
			}

			fmt.Printf("[worker %d] результат для %d = %d\n", id, num, res)

			if res > 50 {
				fmt.Printf("[worker %d] результат %d > 50 → ОТМЕНА ВСЕЙ РАБОТЫ!\n", id, res)
				cancel()
				return
			}

			select {
			case results <- res:
			case <-ctx.Done():
				return
			}
		}
	}
}

func main() {
	numbers := []int{3, 7, 2, 8, 1, 9, 4, 11, 5, 6}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() 

	jobs := make(chan int)
	results := make(chan int, len(numbers))

	const numWorkers = 2
	var wg sync.WaitGroup
	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go worker(ctx, cancel, i, jobs, results, &wg)
	}

	go func() {
		defer close(jobs)
		for _, n := range numbers {
			select {
			case jobs <- n:
			case <-ctx.Done():
				fmt.Println("[producer] контекст отменён, перестаю класть задачи")
				return
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var collected []int
	for r := range results {
		collected = append(collected, r)
	}

	fmt.Println("----------------------------------------")
	if ctx.Err() != nil {
		fmt.Println("Работа была отменена:", ctx.Err())
	}
	fmt.Println("Собрано результатов:", collected)
}