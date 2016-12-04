// sensor.go
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

const (
	TagSensorName = "@sensorName@"
	TagPrevVal    = "@prevVal@"
	TagLastVal    = "@lastVal@"
	TagCondition  = "@condition@"
)

// -----------------------------------------------

var sensorTickersLock sync.Mutex
var sensorTickers = map[int]*time.Ticker{}

var sensorPrevValLock sync.Mutex
var sensorPrevVal = map[int]string{}

// -----------------------------------------------

func init() {
	RegisterInternalFunc(SensorFunc, "CpuUsage", CpuUsage)
	RegisterInternalFunc(SensorFunc, "MemoryUsage", MemoryUsage)
}

// sensorSetup : read defined sensors from DB then create a ticker and start reading goroutine for each sensor
func sensorSetup(db *sql.DB) (err error) {

	sensorObjs, err := getHomeObjects(db, ItemSensor, -1)
	if err != nil {
		return
	}
	if glog.V(3) {
		glog.Info("\nSensor Objs\n", sensorObjs)
	}

	for _, sensor := range sensorObjs {
		sensorUpdateTicker(sensor)
	}

	if glog.V(1) {
		glog.Infof("sensorSetup Done (%d)", len(sensorTickers))
	}

	return
}

func sensorUpdateTicker(sensor HomeObject) (err error) {
	sensorTickersLock.Lock()
	defer sensorTickersLock.Unlock()

	if glog.V(3) {
		glog.Info("Sensor Values", sensor.Values)
	}

	// If sensor already active, stop it
	ticker, exist := sensorTickers[sensor.Values[0].IdObject]
	if exist {
		ticker.Stop()
		delete(sensorTickers, sensor.Values[0].IdObject)
	}

	isActive, err := sensor.getIntVal("IsActive")
	if err != nil || isActive == 0 {
		return
	}

	durationStr, err := sensor.getStrVal("Interval")
	if err != nil {
		return
	}

	// for empty duration, don't setup ticker
	if len(strings.Trim(durationStr, " ")) <= 0 {
		return
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		glog.Errorf("Falied to parse duration (%s) : %s", durationStr, err)
		return
	}

	if glog.V(2) {
		glog.Infof("Sensor %d (#Act=%d) / %s => New Ticker (%d)", sensor.Values[0].IdObject, len(sensor.linkedObjs), durationStr, duration)
	}

	sensorTickers[sensor.Values[0].IdObject] = time.NewTicker(duration)

	go readSensor(sensor)

	return
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

// readSensorValue : perform sensor readings
func readSensorValue(sensor HomeObject) (result string, err error) {
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

	if isInternal != 0 {
		result, err = CallInternalFunc(SensorFunc, readCmd, readParam, "")
	} else {
		result, err = LaunchExternalCmd(SensorFunc, readCmd, readParam, "")
	}

	return
}

// readSensor : call readSensorValue according to initialised corresponding ticker and handleSensorValue
func readSensor(sensor HomeObject) {
	sensorTickersLock.Lock()
	localTicker := sensorTickers[sensor.Values[0].IdObject]
	sensorTickersLock.Unlock()

	for t := range localTicker.C {
		result, err := readSensorValue(sensor)
		if err != nil {
			continue
		}

		handleSensorValue(t, sensor, result)
	}
}

// handleSensorValue : trigger actor and store sensor value in DB
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
	prevVal, found := sensorPrevVal[sensor.Values[0].IdObject]
	if !found {
		prevVal = value // todo get prev val from db ?
	}
	sensorPrevVal[sensor.Values[0].IdObject] = value
	sensorPrevValLock.Unlock()

	// Record value if required
	if record != 0 { // todo : add more options : 0=never, 1=always, 2=only if change, ...
		go recordSensorValue(t, sensor, value)
	}
	// Trigger linked sensorAct if any
	for _, sensorAct := range sensor.linkedObjs {
		go triggerSensorAct(sensorAct, sensorName, prevVal, value)
	}
}

// recordSensorValue : store in DB a value for a given sensor reading
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

	// check value regarding datatype
	switch TDataType(dataType) {
	case DBTypeBool, DBTypeInt, DBTypeDateTime:
		_, err := strconv.Atoi(value)
		if err != nil {
			glog.Errorf("Fail to get int(%s) for sensor %d : %s", value, sensorId, err)
			return
		}
	case DBTypeText:
	case DBTypeFloat:
		_, err := strconv.ParseFloat(value, 64)
		if err != nil {
			glog.Errorf("Fail to get float64(%s) for sensor %d : %s", value, sensorId, err)
			return
		}
	default:
		glog.Errorf("Unknown data type %d for sensor %d", dataType, sensorId)
		return
	}

	_, err = db.Exec("insert into HistoSensor values ( ?, ?, ?);", t.Unix(), sensorId, value)
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

		triggerActorById(actorId, 1, actorParam)
	}

}

// -----------------------------------------------
// -----------------------------------------------

func CpuUsage(param1 string, param2 string) (string, error) {
	glog.V(2).Info("CpuUsage Not Implemented") // TODO
	return "99", nil
}

func MemoryUsage(param1 string, param2 string) (string, error) {
	glog.V(2).Info("MemoryUsage Not Implemented") // TODO
	return "99", nil
}
