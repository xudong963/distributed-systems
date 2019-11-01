package main

type IStream interface {
	Write(buf []byte) (n int, err error)
	Read(buf []byte) (n int, err error)
}

type Writer interface {
	Write(buf []byte) (n int, err error)
}

// 接口查询
/* var file1 Writer = ...
   if file5, ok := file1.(IStream); ok {
		...
    }
 */
