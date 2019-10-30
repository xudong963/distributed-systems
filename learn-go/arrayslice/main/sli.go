package main

func main() {
	a := []int {1, 2, 3}
	b := make([]int, 1)
	b = append(b, a[1:]...)
	for _, v := range b {
		println(v)
	}
}

