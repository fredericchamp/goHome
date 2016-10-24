// external.go
package main

import (
	"time"

	"github.com/golang/glog"
)

func LaunchExternalCmd(cmd string, param1 string, param2 string) (string, error) {
	// TODO
	glog.Info("LaunchExternalCmd Not Implemented ", param1, " + ", param2)
	return time.Now().String(), nil
}
