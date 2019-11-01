package main

// 在go语言中, 一个类只需要实现了接口要求的所有函数
// 我们就说这个类实现了该接口
//
// 非入侵式接口
type File struct {
	//...
}

func (f* File) Read(buf []byte) (n int, err error)  {
	return
}

func (f* File) Write(buf []byte) (n int, err error)  {
	return
}

func (f* File) Seek(off int64, whence int) (pos int64, err error) {
	return
}

func (f* File) Close() (err error) {
	return
}

// 以上定义了File类, 并且实现了 Read(), Write(), Seek(), Close()等方法

type IFile interface {
	Read(buf []byte) (n int, err error)
	Write(buf []byte) (n int, err error)
	Seek(off int64, whence int) (pos int64, err error)
	Close() (err error)
}

type IReader interface {
	Read(buf []byte) (n int, err error)
}

type IWriter interface {
	Write(buf []byte) (n int, err error)
}

type ICloser interface {
	Close() (err error)
}


// 尽管File类并没有从这些接口继承, 但是File类实现了这些接口,并可以进行赋值

var file1 IFile = new (File)
var file2 IReader = new (File)