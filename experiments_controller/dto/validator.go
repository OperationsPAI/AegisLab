package dto

import (
	"fmt"
	"strconv"

	"github.com/go-playground/validator/v10"
)

// 处理非 validator.ValidationErrors 类型的错误
func formatOtherError(err error) string {
	message := "Invalid request parameters"

	if numErr, ok := err.(*strconv.NumError); ok {
		message = fmt.Sprintf("The input %s is invalid", numErr.Num)
	}

	return message
}

func FormatErrorMessage(err error, fieldMap map[string]string) string {
	var validationErrors validator.ValidationErrors
	var ok bool
	if validationErrors, ok = err.(validator.ValidationErrors); !ok {
		return formatOtherError(err)
	}

	error := validationErrors[0]
	field := fieldMap[error.Field()]
	tag := error.Tag()

	var message string
	switch tag {
	case "required":
		message = fmt.Sprintf("The field %s is required", field)
	case "min":
		message = fmt.Sprintf("The field %s must be larger than or equal to %s", field, error.Param())
	case "max":
		message = fmt.Sprintf("The field %s must be smaller than or equal to %s", field, error.Param())
	default:
		message = "Invalid request parameters"
	}

	return message
}
