// internal.go
package main

import (
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
)

var sensorTickers = map[string]*time.Ticker{}

func init() {
	RegisterInternalFunc("CpuUsage", CpuUsage)
	RegisterInternalFunc("MemoryUsage", MemoryUsage)
	RegisterInternalFunc("GPIO", GPIO)
	RegisterInternalFunc("PoloTicker", PoloTicker)
	RegisterInternalFunc("SerialATSMS", SerialATSMS)
}

func sensorSetup(db *sql.DB) (err error) {
	sensorItems, err := getManageItems(db, -1, ItemSensor)
	if err != nil {
		return
	}
	if glog.V(2) {
		glog.Info("\nSensor Items\n", sensorItems)
	}

	var sensorObjs []HomeObject
	for _, v := range sensorItems {
		lst, err := getDBObjects(db, -1, v.Id)
		if err != nil {
			return err
		}
		for _, v := range lst {
			sensorObjs = append(sensorObjs, v)
		}
	}
	if glog.V(2) {
		glog.Info("\nSensor\n", sensorObjs)
	}

	for _, sensor := range sensorObjs {
		var duration time.Duration
		var i int

		isActive, err := sensor.getIntVal("IsActive")
		if err != nil {
			return err
		}
		if isActive == 0 {
			continue
		}

		sensorName, err := sensor.getStrVal("Name")
		if err != nil {
			return err
		}

		durationStr, err := sensor.getStrVal("Interval")
		if err != nil {
			return err
		}

		switch {
		case strings.HasSuffix(durationStr, DurationMS):
			i, err = strconv.Atoi(strings.TrimSuffix(durationStr, DurationMS))
			duration = time.Millisecond * time.Duration(i)
		case strings.HasSuffix(durationStr, DurationS):
			i, err = strconv.Atoi(strings.TrimSuffix(durationStr, DurationS))
			duration = time.Second * time.Duration(i)
		case strings.HasSuffix(durationStr, DurationM):
			i, err = strconv.Atoi(strings.TrimSuffix(durationStr, DurationM))
			duration = time.Minute * time.Duration(i)
		case strings.HasSuffix(durationStr, DurationH):
			i, err = strconv.Atoi(strings.TrimSuffix(durationStr, DurationH))
			duration = time.Hour * time.Duration(i)
		}
		if err != nil {
			glog.Error("Falied to read duration (", durationStr, ") :", err)
			return err
		}

		if glog.V(2) {
			glog.Infof("Sensor %s / %s => New Ticker (%d)", sensorName, durationStr, duration)
		}
		sensorTickers[sensorName] = time.NewTicker(duration)

		go readSensor(sensor)

	}
	return
}

func readSensor(sensor HomeObject) {
	sensorName, err := sensor.getStrVal("Name")
	if err != nil {
		return
	}

	readCmd, err := sensor.getStrVal("ReadCmd")
	if err != nil {
		return
	}
	readParam, err := sensor.getStrVal("ReadParam")
	if err != nil {
		return
	}
	isInternal, err := sensor.getIntVal("IsInternal")
	if err != nil {
		return
	}

	record, err := sensor.getIntVal("Record")
	if err != nil {
		return
	}

	for t := range sensorTickers[sensorName].C {
		var result string
		var err error
		if isInternal != 0 {
			result, err = CallInternalFunc(readCmd, readParam, "")
		} else {
			result, err = LaunchExternalCmd(readCmd, readParam, "")
		}
		if err != nil {
			continue
		}
		if record != 0 {
			go RecordSensorValue(t, sensor, result)
		}
	}
}

func sensorCleanup() {

}

func RecordSensorValue(t time.Time, sensor HomeObject, result string) {
	// TODO
	sensorName, err := sensor.getStrVal("Name")
	if err != nil {
		return
	}
	glog.Infof("RecordSensorValue for %s at %s = %s ", sensorName, t, result)
}

func CpuUsage(param1 string, param2 string) (string, error) {
	// TODO
	glog.Info("CpuUsage Not Implemented ", param1, " + ", param2)
	return time.Now().String(), nil
}

func MemoryUsage(param1 string, param2 string) (string, error) {
	// TODO
	glog.Info("MemoryUsage Not Implemented ", param1, " + ", param2)
	return time.Now().String(), nil
}

func GPIO(param1 string, param2 string) (string, error) {
	// TODO
	glog.Info("GPIO Not Implemented ", param1, " + ", param2)
	return time.Now().String(), nil
}

func PoloTicker(param1 string, param2 string) (string, error) {
	// TODO
	glog.Info("PoloTicker Not Implemented ", param1, " + ", param2)
	return time.Now().String(), nil
}

func SerialATSMS(param1 string, param2 string) (string, error) {
	// TODO
	glog.Info("SerialATSMS Not Implemented ", param1, " + ", param2)
	return time.Now().String(), nil
}
