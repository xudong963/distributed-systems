package main

import "fmt"
// 不定参数
func myfunc(args ...int)  {
	for _, arg := range args {
		fmt.Print(arg)
	}
}

func main()  {
	myfunc(1, 2, 3)
	myfunc(1, 2, 3, 4)
}