package slicex

import (
	"golang.org/x/exp/constraints"
	"math/rand"
)

// All 检查切片中所有元素是否都满足条件 f
// 如果切片为空，返回 true
func All[T any](elems []T, f func(T) bool) bool {
	for _, e := range elems {
		if !f(e) {
			return false
		}
	}
	return true
}

// Any 检查切片中是否存在满足条件 f 的元素
func Any[T any](elems []T, f func(T) bool) bool {
	for _, e := range elems {
		if f(e) {
			return true
		}
	}
	return false
}

// Contains 检查切片中是否包含指定元素
func Contains[T comparable](elems []T, a T) bool {
	for _, e := range elems {
		if e == a {
			return true
		}
	}
	return false
}

// LenFunc 返回切片中满足条件 f 的元素数量
func LenFunc[T any](elems []T, f func(T) bool) int {
	return len(FilterFunc(elems, f))
}

// EachFunc 对切片中的每个元素执行函数 f
func EachFunc[T any](elems []T, f func(T)) {
	for _, elem := range elems {
		f(elem)
	}
}

// FirstFunc 返回切片中第一个满足条件 f 的元素和是否找到的布尔值
func FirstFunc[T any](elems []T, f func(T) bool) (T, bool) {
	for _, v := range elems {
		if f(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}

// FirstOrDefFunc 返回切片中第一个满足条件 f 的元素
// 如果未找到，返回默认值 def
func FirstOrDefFunc[T any](elems []T, def T, f func(T) bool) T {
	for _, v := range elems {
		if f(v) {
			return v
		}
	}
	return def
}

// FirstIndexFunc 返回切片中第一个满足条件 f 的元素的索引
// 如果未找到，返回 -1
func FirstIndexFunc[T any](elems []T, f func(T) bool) int {
	for k, v := range elems {
		if f(v) {
			return k
		}
	}
	return -1
}

// FilterFunc 返回切片中所有满足条件 f 的元素组成的新切片
func FilterFunc[T any](elems []T, f func(T) bool) []T {
	results := make([]T, 0, len(elems))
	for _, s := range elems {
		if f(s) {
			results = append(results, s)
		}
	}
	return results
}

// LastFunc 返回切片中最后一个满足条件 f 的元素和是否找到的布尔值
func LastFunc[T any](elems []T, f func(T) bool) (T, bool) {
	for i := len(elems) - 1; i >= 0; i-- {
		if f(elems[i]) {
			return elems[i], true
		}
	}
	var zero T
	return zero, false
}

// LastIndexFunc 返回切片中最后一个满足条件 f 的元素的索引
// 如果未找到，返回 -1
func LastIndexFunc[T any](elems []T, f func(T) bool) int {
	for i := len(elems) - 1; i >= 0; i-- {
		if f(elems[i]) {
			return i
		}
	}
	return -1
}

// RandomFunc 返回切片中随机一个满足条件 f 的元素和是否找到的布尔值
// 如果未找到满足条件的元素，返回 zero 值和 false
func RandomFunc[T any](elems []T, f func(T) bool) (T, bool) {

	var matches []T
	for _, v := range elems {
		if f(v) {
			matches = append(matches, v)
		}
	}

	if len(matches) == 0 {
		var zero T
		return zero, false
	}

	return matches[rand.Intn(len(matches))], true
}

// MapFunc 将切片中的每个元素通过函数 f 转换，返回新切片
func MapFunc[S, T any](elems []S, f func(S) T) []T {
	results := make([]T, len(elems))
	for i, s := range elems {
		results[i] = f(s)
	}
	return results
}

// Max 返回切片中的最大值和是否找到的布尔值
// 如果切片为空，返回 false
func Max[T constraints.Ordered](elems []T) (T, bool) {
	if len(elems) == 0 {
		var zero T
		return zero, false
	}

	max := elems[0]
	for _, v := range elems[1:] {
		if v > max {
			max = v
		}
	}
	return max, true
}

// MaxFunc 根据转换函数 f 的结果返回切片中的"最大"元素和是否找到的布尔值
// 如果切片为空，返回 false
func MaxFunc[T any, O constraints.Ordered](elems []T, f func(T) O) (T, bool) {
	if len(elems) == 0 {
		var zero T
		return zero, false
	}

	maxElem := elems[0]
	maxVal := f(maxElem)
	for _, v := range elems[1:] {
		current := f(v)
		if current > maxVal {
			maxVal = current
			maxElem = v
		}
	}
	return maxElem, true
}

// NewFunc 创建一个新切片，每个元素由函数 f 根据索引生成
func NewFunc[T any](length int, f func(int) T) []T {
	elems := make([]T, length)
	for i := range elems {
		elems[i] = f(i)
	}
	return elems
}
