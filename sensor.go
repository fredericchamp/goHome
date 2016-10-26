// internal.go
package main

import (
	"database/sql"
	"fmt"
	"math/rand" // TODO : remove (for testing purpose only)
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
)

var sensorTickersLock sync.Mutex
var sensorTickers = map[string]*time.Ticker{}

func init() {
	RegisterInternalFunc(SensorFunc, "CpuUsage", CpuUsage)
	RegisterInternalFunc(SensorFunc, "MemoryUsage", MemoryUsage)
	RegisterInternalFunc(SensorFunc, "GPIO", SensorGPIO)
	rand.Seed(2948536) // TODO : remove (for testing purpose only)
}

// sensorSetup : read defined sensors from DB then create a ticker and start reading goroutine for each sensor
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

	sensorTickersLock.Lock()
	defer sensorTickersLock.Unlock()

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

// readSensor : perform sensor readings using ReadCmd according to initialised corresponding ticker
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

	sensorTickersLock.Lock()
	localTicker := sensorTickers[sensorName]
	sensorTickersLock.Unlock()

	for t := range localTicker.C {
		var result string
		var err error
		if isInternal != 0 {
			result, err = CallInternalFunc(SensorFunc, readCmd, readParam, "")
		} else {
			result, err = LaunchExternalCmd(SensorFunc, readCmd, readParam, "")
		}
		if err != nil {
			continue
		}
		if record != 0 {
			go recordSensorValue(t, sensor, result)
		}
	}
}

// sensorCleanup : stop and remove all sensor ticker
func sensorCleanup() {
	sensorTickersLock.Lock()
	for key, ticker := range sensorTickers {
		ticker.Stop()
		delete(sensorTickers, key)
	}
	sensorTickersLock.Unlock()
	if glog.V(2) {
		glog.Info("sensorCleanup Done")
	}
}

// recordSensorValue : store in DB a value for a given sensor reading
// TODO : allow to store value outsite of main DB
func recordSensorValue(t time.Time, sensor HomeObject, value string) {
	db, err := openDB()
	if err != nil {
		glog.Error(err)
		return
	}
	defer db.Close()

	sensorId := sensor.getId()
	dataType, err := sensor.getIntVal("IdDataType")
	if err != nil {
		return
	}

	switch dataType {
	case DBTypeBool, DBTypeInt, DBTypeDateTime:
		intVal, err := strconv.Atoi(value)
		if err != nil {
			glog.Errorf("Fail to get int(%s) for sensor %d from : %s", value, sensorId, err)
			return
		}
		_, err = db.Exec("insert into HistoSensor values ( ?, ?, ?, ?, ?);", t.Unix(), sensorId, intVal, 0, "")
	case DBTypeText:
		_, err = db.Exec("insert into HistoSensor values ( ?, ?, ?, ?, ?);", t.Unix(), sensorId, 0, 0, value)
	case DBTypeFloat:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			glog.Errorf("Fail to get float64(%s) for sensor %d from : %s", value, sensorId, err)
			return
		}
		_, err = db.Exec("insert into HistoSensor values ( ?, ?, ?, ?, ?);", t.Unix(), sensorId, 0, floatVal, "")
	default:
		glog.Errorf("Unknown data type %d for sensor %d", dataType, sensorId)
	}
	if err != nil {
		glog.Errorf("Fail to store %d value (%s) for sensor %d from : %s ", dataType, value, sensorId, err)
	}
	if glog.V(2) {
		glog.Infof("recordSensorValue : %d - %s", t.Unix(), value)
	}
}

// -----------------------------------------------
// -----------------------------------------------

func CpuUsage(param1 string, param2 string) (string, error) {
	// TODO
	glog.Info("CpuUsage Not Implemented")
	return fmt.Sprint(rand.Intn(100)), nil
}

func MemoryUsage(param1 string, param2 string) (string, error) {
	// TODO
	glog.Info("MemoryUsage Not Implemented")
	return fmt.Sprint(rand.Intn(100)), nil
}

func SensorGPIO(param1 string, param2 string) (string, error) {
	// TODO
	glog.Info("GPIO Not Implemented")
	return time.Now().String(), nil
}
