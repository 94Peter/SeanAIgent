package util

import "reflect"

func GetTypeName(v any) string {
	t := reflect.TypeOf(v)
	if t == nil {
		return "<nil>"
	}
	// 如果是指標，取實際型別
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.String()
}
