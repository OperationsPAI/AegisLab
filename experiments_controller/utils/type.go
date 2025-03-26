package utils

import "reflect"

func GetTypeName(obj any) string {
	objType := reflect.TypeOf(obj)
	if objType != nil && objType.Kind() == reflect.Ptr {
		objType = objType.Elem()
	}

	// 处理指针类型
	if objType.Kind() == reflect.Ptr {
		objType = objType.Elem()
	}

	// 处理匿名类型或无效类型
	objName := "item"
	if objType != nil {
		objName = objType.Name()
	}

	return objName
}
func Must(err error) {
	if err != nil {
		panic(err)
	}
}
