package arrayslice

import "testing"

func TestSum(t *testing.T) {

	numbers :=  []int{1, 2, 3, 4, 5}    //切片类型,与数组的区别:声明时不指定长度
	actual := Sum(numbers)
	expected := 15

	if actual != expected {
		t.Errorf("expected: '%d', but got: '%d'", expected, actual)
	}
}

// go内置有计算测试 覆盖率的工具， 它能够帮助发现没有被测试过的区域
// go test -cover