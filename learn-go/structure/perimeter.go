package structure

type Rectangle struct {
	length float64
	width float64
}


func Perimeter(rectangle Rectangle) float64 {
	return (rectangle.length + rectangle.width) * 2
}



