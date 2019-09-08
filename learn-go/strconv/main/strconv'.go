package main

import (
	"fmt"
	"strconv"
)

func main()  {
	i, err := strconv.Atoi("-42")   // string to int
	fmt.Println(i)
	fmt.Println(err)

	s:= strconv.Itoa(48)  // int to string
	fmt.Println(s)
	//parser
	b, _ := strconv.ParseBool("true")
	f, _ := strconv.ParseFloat("5.66", 64)
	in, _ := strconv.ParseInt("-3", 10, 64)
	ui, _ := strconv.ParseUint("21", 10, 64)
	fmt.Println(b,f,in,ui)

	//format eg
	strBool := strconv.FormatBool(true)
	fmt.Println(strBool)

	q := strconv.Quote("Hello, 世界")
	fmt.Println(q)
	q = strconv.QuoteToASCII("Hello, 世界")
	fmt.Println(q)
}