package main

import (
	"fmt"
	"strings"
	"unicode"
)

func main()  {
	a := "hello hello world"
	b := "world"
	fmt.Println(strings.Compare(a, b))
	fmt.Println(strings.Contains(b, "ld"))
	fmt.Println(strings.Count(a, "l"))
	fmt.Println(strings.Fields(a))

	f := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	}
	fmt.Printf("Fields are: %q", strings.FieldsFunc("  foo1;bar2,baz3...", f))
	//output : Fields are: ["foo1" "bar2" "baz3"]
}
