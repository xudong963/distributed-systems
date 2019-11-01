package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime/pprof"
	"sort"
)

var num = 100000000
var findNum = 10000000

var t = flag.String("t", "map", "use map")

func main() {
	cpuf, err := os.Create("cpu_profile")
	if err != nil {
		log.Fatal(err)
	}

	pprof.StartCPUProfile(cpuf)
	defer pprof.StopCPUProfile()

	flag.Parse()
	if *t == "map" {
		fmt.Println("map")
		findMapMax()
	} else {
		fmt.Println("slice")
		findSliceMax()
	}
}

func findMapMax() {
	m := map[int]int{}
	for i := 0; i < num; i++ {
		m[i] = i
	}

	for i := 0; i < findNum; i++ {
		toFind := rand.Int31n(int32(num))
		_ = m[int(toFind)]
	}
}

func findSliceMax() {
	m := make([]int, num)
	for i := 0; i < num; i++ {
		m[i] = i
	}

	for i := 0; i < findNum; i++ {
		toFind := rand.Int31n(int32(num))
		v := sort.SearchInts(m, int(toFind))
		fmt.Println(v)
	}
}

