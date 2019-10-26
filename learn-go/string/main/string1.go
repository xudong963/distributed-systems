package main

import "fmt"

func main() {
	str := "hello golang"
	for i, ch := range str {
		fmt.Println(i, ch)
	}
}