package method_interface_

import "testing"

//表格驱动测试：创建一系列相同测试方式的测试用例时很有用


func Test_Area(t *testing.T) {
	//匿名结构体数组
	areaTest := []struct{
		shape Shape
		want float64
	}{
		{Rectangle{12, 6}, 72},
		{Circle{10}, 314.1592653589793},
	}

	for _, tt := range areaTest {
		got := tt.shape.Area()
		if got != tt.want {
			t.Errorf("got %.2f want %.2f", got, tt.want)
		}
	}
}