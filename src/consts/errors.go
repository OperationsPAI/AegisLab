package consts

import "errors"

var ErrAuthenticationFailed = errors.New("Authentication failed")
var ErrPermissionDenied = errors.New("Permission denied")
var ErrNotFound = errors.New("Record not found")
var ErrAlreadyExists = errors.New("Record already exists")
var ErrInternal = errors.New("Internal server error")
