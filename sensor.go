// internal.go
package main

import (
	"database/sql"
	//	"errors"
	"go/constant"
	"go/token"
	"go/types"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
)

// -----------------------------------------------

/*
const ( TODO : remove
	DurationMS = "ms"
	DurationS  = "s"
	DurationM  = "m"
	DurationH  = "h"
)
*/

// -----------------------------------------------

const (
	TagSensorName = "@sensorName@"
	TagPrevVal    = "@prevVal@"
	TagLastVal    = "@lastVal@"
	TagCondition  = "@condition@"
)

// -----------------------------------------------

var sensorTickersLock sync.Mutex
var sensorTickers = map[string]*time.Ticker{}

var sensorPrevValLock sync.Mutex
var sensorPrevVal = map[string]string{}

// -----------------------------------------------

func init() {
	RegisterInternalFunc(SensorFunc, "CpuUsage", CpuUsage)
	RegisterInternalFunc(SensorFunc, "MemoryUsage", MemoryUsage)
	RegisterInternalFunc(SensorFunc, "GPIO", SensorGPIO)
}

// sensorSetup : read defined sensors from DB then create a ticker and start reading goroutine for each sensor
func sensorSetup(db *sql.DB) (err error) {

	sensorObjs, err := getHomeObjects(db, 1, -1, ItemSensor)
	if err != nil {
		return
	}
	if glog.V(3) {
		glog.Info("\nSensor Objs\n", sensorObjs)
	}

	sensorTickersLock.Lock()
	defer sensorTickersLock.Unlock()

	for _, sensor := range sensorObjs {

		if glog.V(3) {
			glog.Info("Sensor Values", sensor.Values)
		}
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
		/* TODO : remove
		var duration time.Duration
		var i int
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
		default:
			err = errors.New("Unknown duration format")
		}
		*/
		duration, err := time.ParseDuration(durationStr)
		if err != nil {
			glog.Errorf("Falied to parse duration (%s) : %s", durationStr, err)
			return err
		}

		if glog.V(2) {
			glog.Infof("Sensor %s (#Act=%d) / %s => New Ticker (%d)", sensorName, len(sensor.linkedObjs), durationStr, duration)
		}
		sensorTickers[sensorName] = time.NewTicker(duration)

		go readSensor(sensor)
	}

	if glog.V(1) {
		glog.Infof("sensorSetup Done (%d)", len(sensorTickers))
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

		handleSensorValue(t, sensor, result)
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
	if glog.V(1) {
		glog.Info("sensorCleanup Done")
	}
}

// handleSensorValue : trigger actor and store value in DB
func handleSensorValue(t time.Time, sensor HomeObject, value string) {
	sensorName, err := sensor.getStrVal("Name")
	if err != nil {
		return
	}
	record, err := sensor.getIntVal("Record")
	if err != nil {
		return
	}

	// Previous value
	sensorPrevValLock.Lock()
	prevVal, found := sensorPrevVal[sensorName]
	if !found {
		prevVal = value
	}
	sensorPrevVal[sensorName] = value
	sensorPrevValLock.Unlock()

	// Record value if required
	if record != 0 {
		go recordSensorValue(t, sensor, value)
	}
	// Trigger linked sensorAct if any
	for _, sensorAct := range sensor.linkedObjs {
		go triggerSensorAct(sensorAct, sensorName, prevVal, value)
	}
}

// recordSensorValue : store in DB a value for a given sensor reading
// TODO : allow to store value outsite of main DB
func recordSensorValue(t time.Time, sensor HomeObject, value string) {
	db, err := openDB()
	if err != nil {
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
			glog.Errorf("Fail to get int(%s) for sensor %d : %s", value, sensorId, err)
			return
		}
		_, err = db.Exec("insert into HistoSensor values ( ?, ?, ?, ?, ?);", t.Unix(), sensorId, intVal, 0, "")
	case DBTypeText:
		_, err = db.Exec("insert into HistoSensor values ( ?, ?, ?, ?, ?);", t.Unix(), sensorId, 0, 0, value)
	case DBTypeFloat:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			glog.Errorf("Fail to get float64(%s) for sensor %d : %s", value, sensorId, err)
			return
		}
		_, err = db.Exec("insert into HistoSensor values ( ?, ?, ?, ?, ?);", t.Unix(), sensorId, 0, floatVal, "")
	default:
		glog.Errorf("Unknown data type %d for sensor %d", dataType, sensorId)
	}
	if err != nil {
		glog.Errorf("Fail to store %d value (%s) for sensor %d : %s ", dataType, value, sensorId, err)
	}
	if glog.V(2) {
		sensorName, _ := sensor.getStrVal("Name")
		glog.Infof("recordSensorValue for %s (%s)", sensorName, value)
	}
}

// triggerSensorAct
func triggerSensorAct(sensorAct HomeObject, sensorName string, prevVal string, lastVal string) {
	sensorActId := sensorAct.getId()

	// Prepare condition
	condition, err := sensorAct.getStrVal("Condition")
	if err != nil {
		return
	}
	condition = strings.TrimSpace(condition)
	condition = strings.Replace(condition, TagSensorName, sensorName, -1)
	condition = strings.Replace(condition, TagPrevVal, prevVal, -1)
	condition = strings.Replace(condition, TagLastVal, lastVal, -1)
	if glog.V(2) {
		glog.Infof("Condition for sensorAct #%d (%s) = '%s'", sensorActId, sensorName, condition)
	}

	launchAct := false
	if len(condition) <= 0 {
		launchAct = true
	} else {
		// Eval condition
		tv, err := types.Eval(token.NewFileSet(), nil, token.NoPos, condition)
		if err != nil {
			glog.Errorf("Fail to eval condition for sensorAct #%d (%s) '%s' : %s", sensorActId, sensorName, condition, err)
		} else {
			launchAct = constant.BoolVal(tv.Value)
		}
	}
	if launchAct {
		// Prepare actorParam
		actorParam, err := sensorAct.getStrVal("ActorParam")
		if err != nil {
			actorParam = lastVal
		} else {
			actorParam = strings.TrimSpace(actorParam)
			actorParam = strings.Replace(actorParam, TagSensorName, sensorName, -1)
			actorParam = strings.Replace(actorParam, TagPrevVal, prevVal, -1)
			actorParam = strings.Replace(actorParam, TagLastVal, lastVal, -1)
			actorParam = strings.Replace(actorParam, TagCondition, condition, -1)
		}

		actorId, err := sensorAct.getIntVal("idActor")
		if err != nil {
			return
		}

		triggerActorById(actorId, actorParam)
	}

}

// -----------------------------------------------
// -----------------------------------------------

func CpuUsage(param1 string, param2 string) (string, error) {
	// TODO CpuUsage
	glog.V(2).Info("CpuUsage Not Implemented")
	return "99", nil
}

func MemoryUsage(param1 string, param2 string) (string, error) {
	// TODO MemoryUsage
	glog.V(2).Info("MemoryUsage Not Implemented")
	return "99", nil
}

func SensorGPIO(param1 string, param2 string) (string, error) {
	// TODO SensorGPIO
	glog.V(2).Info("GPIO Not Implemented")
	return "1", nil
}