package main

import (
	"fmt"
	"math"
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

func main() {
	shapes := []Shape{
		Circle{Radius: 1},
		Rectangle{Width: 3, Height: 4},
		Triangle{A: 3, B: 4, C: 5},
	}

	area, perimeter := TotalAreaAndPerimeter(shapes)

	fmt.Printf("Фигур:    %d\n", len(shapes))
	fmt.Printf("Площадь:  %.4f\n", area)
	fmt.Printf("Периметр: %.4f\n", perimeter)

	expectedArea := math.Pi + 12 + 6
	expectedPerimeter := 2*math.Pi + 14 + 12

	fmt.Printf("Ожидалось площадь:  %.4f\n", expectedArea)
	fmt.Printf("Ожидалось периметр: %.4f\n", expectedPerimeter)
	fmt.Printf("Совпадает: %v\n",
		math.Abs(area-expectedArea) < 1e-9 &&
			math.Abs(perimeter-expectedPerimeter) < 1e-9)
}