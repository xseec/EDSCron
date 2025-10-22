package copierx

import (
	"reflect"
	"time"

	"github.com/jinzhu/copier"
	"seeccloud.com/edscron/pkg/vars"
)

// Copy 在copier.Copy基础上，支持time.Time和int64/string的转换
func Copy(toValue any, fromValue any) {
	// 先使用原始copier进行复制
	copier.Copy(toValue, fromValue)

	fromVal := reflect.ValueOf(fromValue)
	toVal := reflect.ValueOf(toValue)

	// 检查目标是否为指针且不为nil
	if toVal.Kind() != reflect.Ptr || toVal.IsNil() {
		return
	}
	toVal = toVal.Elem() // 解引用指针

	// 处理源指针
	if fromVal.Kind() == reflect.Ptr {
		if fromVal.IsNil() {
			return // 源指针为nil，直接返回
		}
		fromVal = fromVal.Elem() // 解引用源指针
	}

	// 分别处理切片和结构体类型
	switch toVal.Kind() {
	case reflect.Slice:
		copySlice(toVal, fromVal)
	case reflect.Struct:
		copyStruct(toVal, fromVal)
	}
}

// 处理切片类型的复制
func copySlice(toVal, fromVal reflect.Value) {
	// 检查源是否也是切片
	if fromVal.Kind() != reflect.Slice {
		return
	}

	// 确保目标切片有足够长度
	if toVal.Len() < fromVal.Len() {
		return
	}

	// 遍历切片元素，逐个复制
	for i := 0; i < fromVal.Len(); i++ {
		fromElem := fromVal.Index(i)
		toElem := toVal.Index(i)

		// 如果元素是结构体，处理字段转换
		if fromElem.Kind() == reflect.Struct && toElem.Kind() == reflect.Struct {
			copyStruct(toElem, fromElem)
		} else if fromElem.Kind() == reflect.Ptr && !fromElem.IsNil() &&
			toElem.Kind() == reflect.Ptr && !toElem.IsNil() {
			// 如果元素是指针且指向结构体
			copyStruct(toElem.Elem(), fromElem.Elem())
		}
	}
}

// 处理结构体类型的复制
func copyStruct(toElem, fromElem reflect.Value) {
	toType := toElem.Type()

	// 遍历目标结构体的字段
	for i := 0; i < toElem.NumField(); i++ {
		toField := toElem.Field(i)
		toFieldName := toType.Field(i).Name

		// 在源结构体中查找同名字段
		fromField := fromElem.FieldByName(toFieldName)
		if !fromField.IsValid() || !toField.CanSet() {
			continue // 字段不存在或不可设置，跳过
		}

		switch {
		// time.Time -> string
		case fromField.Type() == reflect.TypeOf(time.Time{}) && toField.Type().Kind() == reflect.String:
			timeVal := fromField.Interface().(time.Time)
			toField.SetString(timeVal.Format(vars.DatetimeFormat))

		// string -> time.Time
		case fromField.Type().Kind() == reflect.String && toField.Type() == reflect.TypeOf(time.Time{}):
			strVal := fromField.String()
			t, err := time.ParseInLocation(vars.DatetimeFormat, strVal, time.Local)
			if err == nil {
				toField.Set(reflect.ValueOf(t))
			}

		// time.Time -> int64
		case fromField.Type() == reflect.TypeOf(time.Time{}) && toField.Type().Kind() == reflect.Int64:
			timeVal := fromField.Interface().(time.Time)
			toField.SetInt(timeVal.Unix())

		// int64 -> time.Time
		case fromField.Type().Kind() == reflect.Int64 && toField.Type() == reflect.TypeOf(time.Time{}):
			intVal := fromField.Int()
			t := time.Unix(intVal, 0)
			toField.Set(reflect.ValueOf(t))

		default:
		}
	}
}
