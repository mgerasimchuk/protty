package util

import (
	"reflect"
	"regexp"
	"runtime"
	"strings"
)

// GetCurrentFuncName returns caller name func in format: <package>.<type name>.<func name>
// example: "main.StructType.FuncName"
func GetCurrentFuncName() string {
	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)

	// convert string from "github.com/project/repo/main.(*StructType).FuncName" to string: "main.StructType.FuncName"
	return regexp.MustCompile(`.*/|\\(\\*|\\)`).ReplaceAllString(runtime.FuncForPC(pc[0]).Name(), "")
}

// GetTypeNameByObject returns type of the object in format: <package>.<type name>
// example: "main.StructType"
func GetTypeNameByObject(obj interface{}) string {
	return strings.Trim(reflect.TypeOf(obj).String(), "*")
}

// GetFuncName returns name of function in format: <package>.<type name>.<func name>
// example: "main.StructType.FuncName"
func GetFuncName(f interface{}) string {
	// convert string from "github.com/project/repo/main.(*StructType).FuncName" to string: "main.StructType.FuncName"
	return regexp.MustCompile(`.*/|\(\*|\)|-fm`).ReplaceAllString(runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name(), "")
}
