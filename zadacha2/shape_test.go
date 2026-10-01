package main

import (
	"math"
	"testing"
)

const eps = 1e-9

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < eps
}

func TestCircle(t *testing.T) {
	c := Circle{Radius: 2}
	if !almostEqual(c.Area(), math.Pi*4) {
		t.Errorf("Circle.Area = %v, want %v", c.Area(), math.Pi*4)
	}
	if !almostEqual(c.Perimeter(), 2*math.Pi*2) {
		t.Errorf("Circle.Perimeter = %v, want %v", c.Perimeter(), 2*math.Pi*2)
	}
}

func TestRectangle(t *testing.T) {
	r := Rectangle{Width: 3, Height: 4}
	if got := r.Area(); got != 12 {
		t.Errorf("Rectangle.Area = %v, want 12", got)
	}
	if got := r.Perimeter(); got != 14 {
		t.Errorf("Rectangle.Perimeter = %v, want 14", got)
	}
}

func TestTriangle(t *testing.T) {
	tr := Triangle{A: 3, B: 4, C: 5}
	if got := tr.Area(); !almostEqual(got, 6) {
		t.Errorf("Triangle.Area = %v, want 6", got)
	}
	if got := tr.Perimeter(); got != 12 {
		t.Errorf("Triangle.Perimeter = %v, want 12", got)
	}
}

func TestTriangleEquilateral(t *testing.T) {
	tr := Triangle{A: 2, B: 2, C: 2}
	if got := tr.Area(); !almostEqual(got, math.Sqrt(3)) {
		t.Errorf("Triangle.Area = %v, want %v", got, math.Sqrt(3))
	}
}

func TestTotalAreaAndPerimeter(t *testing.T) {
	shapes := []Shape{
		Circle{Radius: 1},
		Rectangle{Width: 3, Height: 4},
		Triangle{A: 3, B: 4, C: 5},
	}
	area, perim := TotalAreaAndPerimeter(shapes)

	wantArea := math.Pi + 12 + 6
	wantPerim := 2*math.Pi + 14 + 12

	if !almostEqual(area, wantArea) {
		t.Errorf("area = %v, want %v", area, wantArea)
	}
	if !almostEqual(perim, wantPerim) {
		t.Errorf("perimeter = %v, want %v", perim, wantPerim)
	}
}

func TestTotalEmpty(t *testing.T) {
	area, perim := TotalAreaAndPerimeter(nil)
	if area != 0 || perim != 0 {
		t.Errorf("empty: got (%v, %v), want (0, 0)", area, perim)
	}
}

func TestImplementsShape(t *testing.T) {
	var _ Shape = Circle{}
	var _ Shape = Rectangle{}
	var _ Shape = Triangle{}
}