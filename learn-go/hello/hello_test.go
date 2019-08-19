package main

import "testing"

func TestHello(t *testing.T) {
	got := hello("robot")
	want := "hello, robot"

	if got != want {
		t.Errorf("got '%s' want '%s'", got, want)  //Errorf中的f表示格式化
	}
}

/*
一些规则
1. 程序需要在一个名为 xxx_test.go的文件中编写
2. 测试函数的命名必须以单词Test开始
3. 测试函数只接受一个参数 t *testing.T
*/


/*
1. 首先编写测试
2. 从测试中捕获需求,这是基本的测试驱动开发
*/