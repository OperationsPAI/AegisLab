package utils

import (
	"reflect"
	"strconv"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func InitValidator() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation("min_ptr", ValidateMinPtr)
		_ = v.RegisterValidation("non_zero_int_slice", ValidateNonZeroIntSlice)
	}
}

func ValidateMinPtr(fl validator.FieldLevel) bool {
	field := fl.Field()

	if field.IsNil() {
		return true
	}

	value := field.Elem()
	param := fl.Param()

	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		minVal, err := strconv.ParseInt(param, 0, 64)
		if err != nil {
			return false
		}
		return value.Int() >= minVal

	default:
		return true
	}
}

func ValidateNonZeroIntSlice(fl validator.FieldLevel) bool {
	field := fl.Field()
	if field.Kind() != reflect.Slice {
		return true
	}

	for i := 0; i < field.Len(); i++ {
		element := field.Index(i)
		if element.CanInt() && element.Int() <= 0 {
			return false
		}
	}

	return true
}
