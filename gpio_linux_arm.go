// gpio_linux_arm.go
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/stianeikeland/go-rpio"
)

// Prevent concurrent access to gpio
var gpioLock sync.Mutex

type GPIOParam struct {
	Pin      int    // Pin number (BCM numbering)
	Do       string // read | write
	Value    string // high | low | toggle
	Duration int    // in ms
	Repeat   int
	Interval int    // in ms
	Op       string // min | max | avg
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

	if glog.V(2) {
		glog.Infof("CallGPIO : %v ", gpioParam)
	}

	gpioLock.Lock()
	defer gpioLock.Unlock()

	if err = rpio.Open(); err != nil {
		glog.Errorf("rpio.Open failed : %s", err)
		return
	}
	defer rpio.Close()

	pin := rpio.Pin(gpioParam.Pin)

	if gpioParam.Do == "write" {
		pin.Output()
	} else {
		pin.Input()
	}

	if gpioParam.Repeat <= 0 {
		gpioParam.Repeat = 1
	}
	vals := make([]rpio.State, gpioParam.Repeat)

	for i := 0; i < gpioParam.Repeat; i++ {
		if gpioParam.Do == "write" {
			setPinVal(pin, gpioParam.Value, false)
		} else {
			vals[i] = pin.Read()
		}
		if gpioParam.Interval > 0 {
			time.Sleep(time.Millisecond * time.Duration(gpioParam.Interval))
		}
	}

	if gpioParam.Duration > 0 {
		time.Sleep(time.Millisecond * time.Duration(gpioParam.Duration))
		if gpioParam.Do == "write" {
			setPinVal(pin, gpioParam.Value, true)
		}
	}

	if gpioParam.Do == "write" {
		result = "Done"
	} else {
		result = calcResult(vals, gpioParam.Op)
	}

	return
}

func setPinVal(pin rpio.Pin, value string, reverse bool) {
	switch value {
	case "toggle":
		pin.Toggle()
		break
	case "high":
		if !reverse {
			pin.Write(rpio.High)
		} else {
			pin.Write(rpio.Low)
		}
		break
	default:
		if !reverse {
			pin.Write(rpio.Low)
		} else {
			pin.Write(rpio.High)
		}
	}
}

func calcResult(vals []rpio.State, op string) (result string) {
	val := rpio.State(0)
	switch op {
	case "min":
		val = vals[0]
		for i := 0; i < len(vals); i++ {
			if val > vals[i] {
				val = vals[i]
			}
		}
		break
	case "max":
		val = vals[0]
		for i := 0; i < len(vals); i++ {
			if val < vals[i] {
				val = vals[i]
			}
		}
		break
	case "avg":
		val = 0
		for i := 0; i < len(vals); i++ {
			val += vals[i]
		}
		val /= rpio.State(len(vals))
		break
	default:
		val = 0
		return fmt.Sprintf("Error unknown op : %s", op)
	}

	result = fmt.Sprintf("%d", val)
	return
}
