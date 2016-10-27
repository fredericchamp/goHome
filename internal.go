// internal.go
package main

import (
	"errors"
	"fmt"
	"sync"

	"github.com/golang/glog"
)

type internalFuncType int

const (
	SensorFunc internalFuncType = 1 + iota
	ActorFunc
)

var internalFuncsLock sync.Mutex
var internalFuncs = map[string]func(string, string) (string, error){}

func getFuncKey(funcType internalFuncType, funcName string) string {
	return fmt.Sprintf("%d %s", funcType, funcName)
}

// RegisterInternalFunc : Add a fucntion so it can be call as a sensor or actor
func RegisterInternalFunc(funcType internalFuncType, funcName string, function func(string, string) (string, error)) error {
	fctKey := getFuncKey(funcType, funcName)

	internalFuncsLock.Lock()
	defer internalFuncsLock.Unlock()

	_, funcExist := internalFuncs[fctKey]
	if funcExist {
		err := errors.New(fmt.Sprintf("Can't register '%s' : Already have a function with this name", fctKey))
		glog.Error(err)
		return err
	}

	internalFuncs[fctKey] = function

	return nil
}

// CallInternalFunc : Call an existing registered func
func CallInternalFunc(funcType internalFuncType, funcName string, param1 string, param2 string) (string, error) {
	fctKey := getFuncKey(funcType, funcName)

	internalFuncsLock.Lock()
	function, funcExist := internalFuncs[fctKey]
	internalFuncsLock.Unlock()

	if !funcExist {
		err := errors.New(fmt.Sprintf("Function '%s' unknown", fctKey))
		glog.Error(err)
		return "", err
	}
	return function(param1, param2)
}

// -----------------------------------------------
// -----------------------------------------------
