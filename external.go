// external.go
package main

import (
	"time"

	"github.com/golang/glog"
)

func LaunchExternalCmd(funcType internalFuncType, cmd string, param1 string, param2 string) (string, error) {
	// TODO
	glog.Infof("LaunchExternalCmd Not Implemented %s(%s,%s)", cmd, param1, param2)
	return time.Now().String(), nil
}
