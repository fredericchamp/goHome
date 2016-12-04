// gpio.go
package main

import (
	"fmt"
	"github.com/stianeikeland/go-rpio"
	"os"
	"time"
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

	if err = rpio.Open(); err != nil {
		glog.Errorf("rpio.Open failed : %s", err)
		return
	}
	defer rpio.Close()

	pin = rpio.Pin(23)

	// Set pin to output mode
	pin.Output()

	// Toggle pin 20 times
	for x := 0; x < 20; x++ {
		pin.Toggle()
		time.Sleep(time.Second / 5)
	}

	result = "Dome"
	return
}
