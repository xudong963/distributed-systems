package main

import "testing"

// 重构
// 将断言重构为函数，减少了重复，提高了测试的可读性。
func TestHello(t *testing.T) {
	
	assertCorrentMessage := func(t *testing.T, got, want string) {
		t.Helper()   //告诉测试套件这个方法是辅助函数, 当测试失败时报告的行号在函数调用中，而不是在辅助函数内部。
		if got != want {
			t.Errorf("got '%s' want '%s' ", got, want)
		}
	}

	t.Run("saying hello to people", func(t *testing.T) {
		got := hello("robot", "")
		want := "hello, robot"
		assertCorrentMessage(t, got, want)
	})

	t.Run("saying hello to people", func(t *testing.T) {
		got := hello("", "")
		want := "hello, world"
		assertCorrentMessage(t, got, want)
	})

	t.Run("in Spanish", func(t* testing.T) {
		got := hello("Elodie", "Spanish")
		want := "Hola, Elodie"
		assertCorrentMessage(t, got, want)
	})
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