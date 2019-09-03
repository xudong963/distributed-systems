package iterator

import "testing"

func TestRepeat(t *testing.T) {

	actual := Repeat("a")
	expected := "aaaaa"

	if actual != expected {
		t.Errorf("expected: '%s' but got '%s'", expected, actual)
	}
}

//基准测试
//测试时，代码会运行b.N次，并测量需要多长时间
func BenchmarkRepeat(b *testing.B) {
	for i:=0; i<b.N; i++ {
		Repeat("a")
	}
}