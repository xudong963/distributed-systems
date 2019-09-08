package main

import (
	"fmt"
	"hash/fnv"
)

func hash(s string) uint32 {
	h := fnv.New32a()
	a, b := h.Write([]byte(s))
	fmt.Println(a)   // a is the length of s
	fmt.Println(b)   // b is err,  nil
	return h.Sum32()
}

func main() {
	fmt.Println(hash("HelloWorld"))
	fmt.Println(hash("HelloWorld."))
}