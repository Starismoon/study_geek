package main

import (
	"fmt"
	"sort"
)

func main() {
	//res := Delete1(5, []int{1, 2, 3, 4, 5})
	//fmt.Println("%v\n", res)
	//res := Delete2(-1, []int{1, 2, 3})
	//fmt.Println("%v\n", res)
	//res := Delete3[int](4, []int{1, 2, 3, 4, 5})
	//fmt.Println(res)
	//res1 := Delete3[string](1, []string{"1", "2", "3", "4", "5"})
	//fmt.Println(res1)
	//res2 := Delete3[bool](4, []bool{true, false, true, false, true})
	//fmt.Println(res2)
	res := Delete4[int]([]int{1, 2, 3, 4, 5, 6}, 0, 2, 1, 4)
	fmt.Println(res, len(res), cap(res))
	res1 := Delete4[string]([]string{"1", "2", "3", "4", "5", "6"}, 0, 1, 4)
	fmt.Println(res1, len(res1), cap(res1))
}

// Delete1 要求一：能够实现删除操作就可以。
func Delete1(idx int, vals []int) []int {
	vals_len := len(vals)
	if idx < 0 || idx >= vals_len {
		panic("索引不合法")
	}
	res := make([]int, vals_len-1)
	index := 0
	for i := 0; i < vals_len; i++ {
		if idx != i {
			res[index] = vals[i]
			index++
		}
	}
	return res
}

// Delete2 要求二：考虑使用比较高性能的实现。
// 要求四：支持缩容，并旦设计缩容机制。
func Delete2(idx int, vals []int) []int {
	if idx < 0 || idx >= len(vals) {
		panic("索引不合法")
	}
	vals = append(vals[:idx], vals[idx+1:]...)
	return vals
}

type SliceCommonType interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | string | bool
}

// Delete3 要求三：改造为泛型方法
func Delete3[T SliceCommonType](idx int, vals []T) []T {
	if idx < 0 || idx >= len(vals) {
		panic("索引不合法")
	}
	vals = append(vals[:idx], vals[idx+1:]...)
	return vals
}

// Delete4 要求四：支持缩容，并旦设计缩容机制。
func Delete4[T SliceCommonType](vals []T, idx ...int) []T {
	idxLen := len(idx)
	valsLen := len(vals)
	if idxLen < 0 || idxLen > valsLen {
		panic("请检查传入索引数量")
	}
	sort.Ints(idx)
	var res []T
	if valsLen-idxLen < valsLen/2 {
		res = make([]T, 0, valsLen/2)
	} else {
		res = make([]T, 0, valsLen)
	}
	res = append(res, vals[:idx[0]]...)
	for i := 1; i < idxLen; i++ {
		if idx[i-1] == idx[i] {
			panic("传入了相同索引")
		}
		res = append(res, vals[idx[i-1]+1:idx[i]]...)
	}
	res = append(res, vals[idx[idxLen-1]+1:]...)
	return res
}
