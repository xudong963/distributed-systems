package main

import "fmt"

const helloPrefix = "hello, "    //常量可以提高程序的性能

func hello(name string) string{
	if name == "" {
		name = "world"
	}
	return helloPrefix + name
}

func main(){

	fmt.Println(hello("Chris"))
}