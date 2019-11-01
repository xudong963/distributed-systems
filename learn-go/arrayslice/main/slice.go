package main

import "fmt"

func main()  {
	slice1 := []int {1,2,3,4,5}
	slice2 := []int {4,5,6}

	copy(slice1, slice2)
	fmt.Println(slice1)
	copy(slice2, slice1)
	fmt.Println(slice2)
	defer func() {

	}()
}
