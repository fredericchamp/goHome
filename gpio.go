// +build !arm

// gpio.go
package main

import (
	"encoding/json"
	"errors"
	"github.com/golang/glog"
	//	"time"
	"fmt"
)

type GPIOParam struct {
	Pin      int
	Do       string
	Value    int
	Duration int // in ms
	Repeat   int
	Interval int // in ms
	Op       string
}

func init() {
	RegisterInternalFunc(ActorFunc, "GPIO", CallGPIO)
	RegisterInternalFunc(SensorFunc, "GPIO", CallGPIO)
}

func CallGPIO(param1 string, param2 string) (result string, err error) {

	var gpioParam GPIOParam
	if len(param1) <= 0 {
		result = "Missing GPIO parameters"
		err = errors.New(result)
		glog.Errorf("GPIO error : %s ", err)
		return
	}
	err = json.Unmarshal([]byte(param1), &gpioParam)
	if err != nil {
		result = fmt.Sprintf("Fail to unmarshal gpioParam '%s' : %s", gpioParam, err)
		glog.Errorf(result)
		return
	}

	glog.Infof("CallGPIO : %v ", gpioParam)

	return "Done", nil
}
