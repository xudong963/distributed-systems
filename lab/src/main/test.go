package main

import "fmt"

func main()  {

	a := make([]int, 5)

		a = append(a, 5)
		a = append(a, 6)

	fmt.Println(a[5])
	fmt.Println(a[6])
	fmt.Println(len(a))
}
