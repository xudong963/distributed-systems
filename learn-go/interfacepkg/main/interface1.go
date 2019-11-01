package main

// 接口赋值两种情况
// 1.将对象实例赋值给接口
// 2.将接口赋值给接口

// 1.
type Integer int

func (a Integer) Less(b Integer) bool {
	return a < b
}

func (a* Integer) Add(b Integer)   {
	*a += b
}

// 定义接口LessAdder
type LessAdder interface {
	Less (b Integer) bool
	Add (b Integer)
}

// 定义接口Lesser
type Lesser interface {
	Less (b Integer) bool
}

// 定义接口Adder
type Adder interface {
	Add (b Integer)
}

var a Integer = 1  // 定义对象实例

var b LessAdder = &a   // 必须是&a, 而不是a
var c Lesser = a
var d Lesser = &a       // 可以转化为上一句的含义
var e Adder = &a
// var f Adder = a     // error


// 2.
// 在go中, 只要两个接口拥有相同的方法列表, 它们就是等同的, 可以相互赋值
// 接口赋值并不要求两个接口必须等价, 只要接口A的方法列表是接口B的方法列表的子集,那么接口B就可以赋值给接口A