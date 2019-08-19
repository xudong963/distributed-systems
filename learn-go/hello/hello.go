package main

import "fmt"

func hello(name string) string{
	return "hello, " + name
}

func main(){
	fmt.Println(hello("Chris"))
}