package common

import (
	"strings"
	"runtime"
)

// GetFunctionName() returns name of the function within which it executed.
func GetFunctionName() string {
	pc, _, _, _ := runtime.Caller(1)
	fullName := runtime.FuncForPC(pc).Name()
	parts := strings.Split(fullName, ".")
	return parts[len(parts)-1]
}
