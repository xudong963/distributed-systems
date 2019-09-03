package arrayslice

import (
	"reflect"
	"testing"
)

func TestSum2(t *testing.T) {

	got := SumAll([]int{1, 2}, []int{3, 4})
	want := []int{3, 7}

	//切片不可以使用= ！=
	if !reflect.DeepEqual(got, want) {
		t.Errorf("want: '%d', but got: '%d'", want, got)
	}
}
