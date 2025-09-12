package utils

import (
	"path/filepath"
	"runtime"
)

func GetCallerInfo(skip int) (file string, line int, funcName string) {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown", 0, "unknown"
	}

	filename := filepath.Base(file)
	fn := runtime.FuncForPC(pc)
	funcName = "unknown"
	if fn != nil {
		funcName = fn.Name()
	}

	return filename, line, funcName
}
