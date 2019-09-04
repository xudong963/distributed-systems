package method_interface_

import "math"

type Rectangle struct {
	length float64
	width float64
}

type Circle struct {
	radius float64
}

//接口
type Shape interface {
	Area() float64
}

// 方法： func(receiverName ReceiverType) MethodName(args)
func (r Rectangle) Area() float64 {
	return r.length * r.width
}

func (c Circle) Area() float64 {
	return math.Pi * c.radius * c.radius
}