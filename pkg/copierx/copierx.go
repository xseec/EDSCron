package copierx

import (
	"reflect"
	"time"

	"github.com/jinzhu/copier"
	"seeccloud.com/edscron/pkg/vars"
)

// Copy 在copier.Copy基础上，支持time.Time和int64/string的转换
func Copy(toValue any, fromValue any) {
	copier.Copy(toValue, fromValue)

	fromVal := reflect.ValueOf(fromValue)
	toVal := reflect.ValueOf(toValue)
	if toVal.Kind() != reflect.Ptr || toVal.IsNil() {
		return
	}

	if fromVal.Kind() == reflect.Ptr {
		if fromVal.IsNil() {
			return // 源指针为nil，直接返回
		}
		fromVal = fromVal.Elem() // 获取指针指向的值
	}

	fromElem := fromVal
	toElem := toVal.Elem()

	// 遍历目标结构体的字段
	toType := toElem.Type()
	for i := 0; i < toElem.NumField(); i++ {
		toField := toElem.Field(i)
		toFieldName := toType.Field(i).Name

		// 在源结构体中查找同名字段
		fromField := fromElem.FieldByName(toFieldName)
		if !fromField.IsValid() {
			continue // 字段不存在，跳过
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
			if err != nil {
				return
			}
			toField.Set(reflect.ValueOf(t))

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
