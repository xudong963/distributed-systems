package arrayslice

//数组有一个有趣的属性，它的大小也属于类型的一部分，
//如果你尝试将 [4]int 作为 [5]int 类型的参数传入函数，是不能通过编译的。它们是不同的类型

func Sum(numbers []int) int {
	var sum int

	for _, number := range numbers{   // _忽略索引
		sum += number
	}
	return sum
}
