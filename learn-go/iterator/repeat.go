package iterator

func Repeat(charactor string) string {
	var repeated string
	for i:=0; i<5; i++ {
		repeated += charactor
	}
	return repeated
}

