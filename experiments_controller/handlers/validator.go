package handlers

import (
	"fmt"
	"strconv"

	"github.com/go-playground/validator/v10"
)

func convertValidationErrors(err error, fieldMap map[string]string) string {
	var validationErrors validator.ValidationErrors
	var ok bool
	if validationErrors, ok = err.(validator.ValidationErrors); !ok {
		// 处理非 validator.ValidationErrors 类型的错误
		if numErr, ok := err.(*strconv.NumError); ok {
			return fmt.Sprintf("The input %s is invalid", numErr.Num)
		}
		return "Invalid request parameters"
	}

	error := validationErrors[0]
	field := fieldMap[error.Field()]
	tag := error.Tag()

	var message string
	switch tag {
	case "required":
		message = fmt.Sprintf("The field %s is required", field)
	case "min":
		message = fmt.Sprintf("The field %s must be greater than or equal to %s", field, error.Param())
	default:
		message = "Invalid request parameters"
	}

	return message
}
