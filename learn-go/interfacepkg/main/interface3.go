package main

// 接口组合
type WWriter interface {
	Write(buf []byte) (n int, err error)
}

type RReader interface {
	Reader(buf []byte) (n int, err error)
}

type ReadWriter interface {
	RReader
	WWriter
}
