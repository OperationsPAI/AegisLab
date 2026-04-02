package consts

import "errors"

var ErrAuthenticationFailed = errors.New("authentication failed")
var ErrPermissionDenied = errors.New("permission denied")
var ErrNotFound = errors.New("record not found")
var ErrAlreadyExists = errors.New("record already exists")
var ErrBadRequest = errors.New("bad request")
var ErrInternal = errors.New("internal server error")
