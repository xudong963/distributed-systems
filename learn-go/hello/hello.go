package main

import "fmt"

const helloPrefix = "hello, "    //常量可以提高程序的性能
const spanish = "Spanish"
const french = "French"
const spanishHelloPrefix = "Hola, "
const frenchHelloPrefix = "Bonjour, "

func hello(name string, language string) string{
	if name == "" {
		name = "world"
	}

	prefix := helloPrefix
	switch language {
	case french:
		prefix = frenchHelloPrefix
	case spanish:
		prefix = spanishHelloPrefix
	}
	if language == spanish {
		return prefix + name
	}

	if language == french {
		return prefix + name
	}
	return helloPrefix + name
}

func main(){

	fmt.Println(hello("Chris", ""))
}