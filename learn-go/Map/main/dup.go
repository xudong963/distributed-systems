package main

import "fmt"

func main()  {
	var v = []int{1,2,3,1,6}
	m := make(map[int]int)
	for i:=0; i<len(v); i++ {
		m[v[i]]=i
	}
	for key, value := range m {
		fmt.Printf("key: %v, value: %v", key, value)
	}
}
