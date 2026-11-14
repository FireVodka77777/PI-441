package main

import "fmt"

type Reader interface {
	Next() (int, bool)
}

type Writer interface {
	Write(n int)
}

type SliceReader struct {
	items []int
	pos   int
}

func NewSliceReader(items []int) *SliceReader {
	return &SliceReader{items: items, pos: 0}
}

func (r *SliceReader) Next() (int, bool) {
	if r.pos >= len(r.items) {
		return 0, false
	}

	val := r.items[r.pos]
	r.pos++

	return val, true
}

type SumWriter struct {
	sum int
}

func (w *SumWriter) Write(n int) {
	w.sum += n
}

func (w *SumWriter) Sum() int {
	return w.sum
}

func pipeline(r Reader, w Writer) {
	for {
		n, ok := r.Next()
		if !ok {
			break
		}
		w.Write(n)
	}
}

func main() {
	numbers := []int{5, 10, 15, 20, 25}
	reader := NewSliceReader(numbers)
	writer := &SumWriter{}

	pipeline(reader, writer)

	fmt.Println("Исходный массив:", numbers)
	fmt.Println("Сумма всех чисел:", writer.Sum())
}