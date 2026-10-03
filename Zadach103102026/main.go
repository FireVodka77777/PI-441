package main
import (
	"fmt"
	"math/rand"
	"time"
)

func sumSlise(numbers []int, ch chan int){
	sum := 0
	for _, n := range numbers{
		sum += n
	}
	ch <- sum
}

func main(){
	rand.Seed(time.Now().UnixNano())
	numbers := make([]int, 10)
	for i := range numbers{
		numbers[i] = rand.Intn(100)
	}
	fmt.Println("Исходный срез:", numbers)
	mid := len(numbers) / 2
	ch := make(chan int)
	go sumSlise(numbers[:mid], ch)
	go sumSlise(numbers[mid:], ch)
	sum1 := <-ch
	sum2 := <-ch
	total := sum1 + sum2

	fmt.Printf("Сумма первой половины: %d\n", sum1)
	fmt.Printf("Сумма второй половины: %d\n", sum2)
	fmt.Printf("Общий сумма:           %d\n", total)
}