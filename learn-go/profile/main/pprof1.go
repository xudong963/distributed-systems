package main

import (
	"fmt"
	"log"
	"net/http"
	_"net/http/pprof"
)

func main() {
	go func() {
		for {
			fmt.Println("hello world")
		}
	}()
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}
