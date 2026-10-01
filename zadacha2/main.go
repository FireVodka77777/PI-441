package main

import (
	"fmt"
	"math"
	"sync"
)

type Shape interface {
	Area() float64
	Perimeter() float64
}

type Circle struct {
	Radius float64
}

func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

func (c Circle) Perimeter() float64 {
	return 2 * math.Pi * c.Radius
}

type Rectangle struct {
	Width  float64
	Height float64
}

func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

func (r Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

type Triangle struct {
	A, B, C float64 
}

func (t Triangle) Perimeter() float64 {
	return t.A + t.B + t.C
}

func (t Triangle) Area() float64 {
	p := t.Perimeter() / 2
	return math.Sqrt(p * (p - t.A) * (p - t.B) * (p - t.C))
}
func TotalAreaAndPerimeter(shapes []Shape) (area float64, perimeter float64) {
	for _, s := range shapes {
		area += s.Area()
		perimeter += s.Perimeter()
	}
	return
}
func TotalAreaAndPerimeterParallel(shapes []Shape, n int) (float64, float64) {
	if len(shapes) == 0 {
		return 0, 0
	}
	if n <= 0 {
		n = 1
	}
	if n > len(shapes) {
		n = len(shapes)
	}

	chunkSize := (len(shapes) + n - 1) / n

	var (
		wg         sync.WaitGroup
		mu         sync.Mutex
		totalArea  float64
		totalPerim float64
	)

	for i := 0; i < len(shapes); i += chunkSize {
		end := i + chunkSize
		if end > len(shapes) {
			end = len(shapes)
		}

		wg.Add(1)
		go func(part []Shape) {
			defer wg.Done()

			localArea, localPerim := 0.0, 0.0
			for _, s := range part {
				localArea += s.Area()
				localPerim += s.Perimeter()
			}

			mu.Lock()
			totalArea += localArea
			totalPerim += localPerim
			mu.Unlock()
		}(shapes[i:end])
	}

	wg.Wait()
	return totalArea, totalPerim
}
func main() {
	shapes := []Shape{
		Circle{Radius: 1},
		Rectangle{Width: 3, Height: 4},
		Triangle{A: 3, B: 4, C: 5},
	}

	const parts = 3

	seqArea, seqPerim := TotalAreaAndPerimeter(shapes)
	parArea, parPerim := TotalAreaAndPerimeterParallel(shapes, parts)

	fmt.Printf("Последовательно: площадь=%.4f, периметр=%.4f\n", seqArea, seqPerim)
	fmt.Printf("Параллельно (%d): площадь=%.4f, периметр=%.4f\n", parts, parArea, parPerim)
	fmt.Printf("Совпадает:       %v\n",
		math.Abs(seqArea-parArea) < 1e-9 && math.Abs(seqPerim-parPerim) < 1e-9)
}