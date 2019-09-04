package Map

import "errors"

type Dictionary map[string]string


/*func (d Dictionary) Search(word string) string {
	return  d[word]
}*/

// 重构： 假设搜索的元素不在map中呢？ （注意: key不在map中 或者key未定义是两种情况）

var ErrNotFound = errors.New("could not find the word you were looking for")

func (d Dictionary) Search(word string) (string, error) {
	ans, ok := d[word]   //返回两个值, ok是布尔类型：表示是否成功找到key
	if !ok {
		return "", ErrNotFound
	}


	return ans, nil
}

// 关于map : map是引用类型,拥有对底层数据结构的引用，就像指针一样
// 永远不应该初始化一个空的map : var m map[string]string x
// var m = map[string]string {}   or var m = make(m[string]string)

