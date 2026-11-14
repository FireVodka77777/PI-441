package main

import (
	"fmt"
	"strconv"
)

type Source interface {
	Next() (data []string, hasMore bool)
}

type SliceSource struct {
	items []string
	pos   int
}

func NewSliceSource(items []string) *SliceSource {
	return &SliceSource{items: items, pos: 0}
}

func (s *SliceSource) Next() ([]string, bool) {
	if s.pos >= len(s.items) {
		return nil, false
	}

	current := s.items[s.pos]
	s.pos++

	return []string{current}, true
}

type GeneratedSource struct {
	limit int
	curr  int
}

func NewGeneratedSource(limit int) *GeneratedSource {
	return &GeneratedSource{limit: limit, curr: 1}
}

func (g *GeneratedSource) Next() ([]string, bool) {
	if g.curr > g.limit {
		return nil, false
	}

	batch := make([]string, 0, 2)
	for i := 0; i < 2 && g.curr <= g.limit; i++ {
		batch = append(batch, "item-"+strconv.Itoa(g.curr))
		g.curr++
	}

	return batch, true
}


func collectAll(s Source) []string {
	result := make([]string, 0)

	for {
		data, hasMore := s.Next()
		if !hasMore {
			break
		}

		result = append(result, data...)
	}

	return result
}

func main() {
	sliceSource := NewSliceSource([]string{"яблоко", "банан", "вишня"})
	fromSlice := collectAll(sliceSource)
	fmt.Println("Источник из среза:", fromSlice)

	genSource := NewGeneratedSource(5)
	fromGen := collectAll(genSource)
	fmt.Println("Генератор (limit=5):", fromGen)

	emptySource := NewSliceSource([]string{})
	fromEmpty := collectAll(emptySource)
	fmt.Println("Пустой источник:", fromEmpty)
}