// internal.go
package main

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/golang/glog"
)

type internalFunctType int

const (
	SensorFunc internalFunctType = 1 + iota
	ActorFunc
)

var internalSensorFuncsLock sync.Mutex
var internalSensorFuncs = map[string]func(string, string) (string, error){
	"DummyFunc": DummyFunc,
}

var internalActorFuncsLock sync.Mutex
var internalActorFuncs = map[string]func(string, string) (string, error){
	"DummyFunc": DummyFunc,
}

// RegisterInternalFunc : Add a fucntion so it can be call as a sensor or actor
func RegisterInternalFunc(funcType internalFunctType, funcName string, function func(string, string) (string, error)) (err error) {
	switch funcType {
	case SensorFunc:
		err = regIntFunc(internalSensorFuncsLock, internalSensorFuncs, funcName, function)
	case ActorFunc:
		err = regIntFunc(internalActorFuncsLock, internalActorFuncs, funcName, function)
	default:
		err = errors.New("Unknown internalFunctType")
	}
	return
}

func regIntFunc(lock sync.Mutex, fctmap map[string]func(string, string) (string, error), funcName string, function func(string, string) (string, error)) error {
	lock.Lock()
	defer lock.Unlock()
	_, funcExist := fctmap[funcName]
	if funcExist {
		glog.Errorf("Can't register '%s' : Already have a function with this name", funcName)
		return errors.New("Already have a function with this name")
	}
	fctmap[funcName] = function
	return nil
}

// CallInternalFunc : Call an existing registered func
func CallInternalFunc(funcType internalFunctType, funcName string, param1 string, param2 string) (string, error) {
	var function func(string, string) (string, error)
	var funcExist bool
	switch funcType {
	case SensorFunc:
		internalSensorFuncsLock.Lock()
		function, funcExist = internalSensorFuncs[funcName]
		internalSensorFuncsLock.Unlock()
	case ActorFunc:
		internalActorFuncsLock.Lock()
		function, funcExist = internalActorFuncs[funcName]
		internalActorFuncsLock.Unlock()
	default:
		err := errors.New("Unknown internalFunctType")
		return "", err
	}
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
