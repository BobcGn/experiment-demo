package main

import (
	"fmt"
	"math"
)

// 形状接口
type Shape interface {
	Area() float64
	Perimeter() float64
}

// 圆形结构体
type Circle struct {
	radius float64
}

// 计算面积的函数
func (c Circle) Area() float64 {
	return math.Pi * c.radius * c.radius
}

// 计算周长的函数
func (c Circle) Perimeter() float64 {
	return 2 * math.Pi * c.radius
}

// main
func main() {
	c := Circle{radius: 5}
	var s Shape = c // 接口变量存储实现了接口的类型
	fmt.Println("面积：", s.Area())
	fmt.Println("周长：", s.Perimeter())
}
