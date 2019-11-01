package main

import "time"

type Context interface {
	Deadline() (deadline time.Time, ok bool)

	Done() <-chan struct{}   // Done 方法返回一个只读的 chan，类型为 struct{}

	Err() error      // Err 方法返回取消的错误原因，因为什么 Context 被取消

	Value(key interface{}) interface{}
}




