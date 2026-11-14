package main

import "fmt"

type Node struct {
	Value int
	Left  *Node
	Right *Node
}

type Visitor interface {
	Visit(n int)
}

type PrintVisitor struct{}

func (v PrintVisitor) Visit(n int) {
	fmt.Println(n)
}

type SumVisitor struct {
	Sum int
}

func (v *SumVisitor) Visit(n int) {
	v.Sum += n
}

func walk(root *Node, v Visitor) {
	if root == nil {
		return
	}

	walk(root.Left, v)
	v.Visit(root.Value)
	walk(root.Right, v)
}

func main() {
	root := &Node{
		Value: 10,
		Left: &Node{
			Value: 5,
			Left:  &Node{Value: 3},
			Right: &Node{Value: 7},
		},
		Right: &Node{
			Value: 15,
			Right: &Node{Value: 20},
		},
	}

	fmt.Println("Обход in-order (левый — узел — правый):")
	printVisitor := PrintVisitor{}
	walk(root, printVisitor)

	fmt.Println("---------------------------------")
	sumVisitor := &SumVisitor{}
	walk(root, sumVisitor)
	fmt.Println("Сумма всех значений:", sumVisitor.Sum)
}