package main

// go语言中任何对象实例都满足空接口interface{}
// 所以空接口看起来像是可以指向任何对象的any类型

var v1 interface{} = 1
var v2 interface{} = "abc"
// ...

// 当函数可以接受任意的对象实例时, 我们会将其声明为 interface{}

func Printf(fmt string, args ...interface{})  { }
func Println(args ...interface{})  { }


