package expx

import (
	"fmt"
	"reflect"
)

// If 是一个泛型三元条件表达式函数
// 参数:
//   - cond: 布尔条件，决定返回哪个值
//   - trueVal: 当cond为true时返回的值
//   - falseVal: 当cond为false时返回的值
//
// 返回:
//   - 根据cond返回trueVal或falseVal
func If[T any](cond bool, trueVal, falseVal T) T {
	if cond {
		return trueVal
	}
	return falseVal
}

// HasZeroError 检查结构体中的指定字段是否为零值
// 参数:
//   - obj: 要检查的结构体实例或指针
//   - noZeros: 不允许为零值的字段名列表
//
// 返回:
//   - 如果发现指定字段为零值，返回包含字段信息的错误
//   - 如果obj不是结构体或没有零值字段，返回nil
func HasZeroError(obj any, noZeros ...string) error {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil
	}

	typ := val.Type()
	for _, name := range noZeros {
		field := val.FieldByName(name)
		if !field.IsValid() {
			continue // 忽略不存在的字段
		}
		if field.IsZero() {
			return fmt.Errorf("%s.%s is zero value", typ.Name(), name)
		}
	}

	return nil
}
