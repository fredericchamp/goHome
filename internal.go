// internal.go
package main

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/golang/glog"
)

var internalFuncsLock sync.Mutex
var internalFuncs = map[string]func(string, string) (string, error){
	"DummyFunc": DummyFunc,
}

// RegisterInternalFunc : Add a fucntion so it can be call as a sensor or actor
func RegisterInternalFunc(funcName string, function func(string, string) (string, error)) error {
	internalFuncsLock.Lock()
	defer internalFuncsLock.Unlock()
	_, funcExist := internalFuncs[funcName]
	if funcExist {
		glog.Errorf("Can't register '%s' : Already have a function with this name", funcName)
		return errors.New("Already have a function with this name")
	}
	internalFuncs[funcName] = function
	return nil
}

// CallInternalFunc : Call an existing registered func
func CallInternalFunc(funcName string, param1 string, param2 string) (string, error) {
	internalFuncsLock.Lock()
	function, funcExist := internalFuncs[funcName]
	internalFuncsLock.Unlock()
	if !funcExist {
		err := errors.New(fmt.Sprintf("Function '%s' unknown", funcName))
		glog.Error(err)
		return "", err
	}
	return function(param1, param2)
}

// -----------------------------------------------
// -----------------------------------------------

func DummyFunc(param1 string, param2 string) (string, error) {
	glog.Info("DummyFunc : does nothing ", param1, " + ", param2)
	return time.Now().String(), nil
}
