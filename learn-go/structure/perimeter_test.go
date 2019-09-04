package structure

import "testing"

func TestPerimeter(t *testing.T){
	rectangle := Rectangle{10.0, 20.0}
	got := Perimeter(rectangle)
	want := 60.0

	if got != want {
		t.Errorf("want: '%f', but got: '%f'", want, got)
	}
}

