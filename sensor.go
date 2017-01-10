// sensor.go
package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"go/constant"
	"go/token"
	"go/types"
	"os"
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
		glog.Infof("Sensor %d (nb act=%d) : Ticker(%v)", sensor.Values[0].IdObject, len(sensor.linkedObjs), duration)
	}

	sensorTickers[sensor.Values[0].IdObject] = time.NewTicker(duration)

	go readSensor(sensor, sensorTickers[sensor.Values[0].IdObject])

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
		result, err = ExecExternalCmd(readCmd, readParam, "")
	}

	if glog.V(2) {
		glog.Infof("readSensorValue %d : %s", sensor.Values[0].IdObject, result)
	}

	return
}

// readSensor : call readSensorValue according to corresponding ticker and handleSensorValue
func readSensor(sensor HomeObject, ticker *time.Ticker) {
	for t := range ticker.C {
		result, err := readSensorValue(sensor)
		if err != nil {
			glog.Errorf("readSensor fail %s ", err)
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
		prevVal = value // TODO get prev val from db ?
	}
	sensorPrevVal[sensor.Values[0].IdObject] = value
	sensorPrevValLock.Unlock()

	// Record value if required
	if record != 0 { // TODO add more options : 0=never, 1=always, 2=only if change, ...
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
	condition = strings.Replace(condition, TagPrevVal, cleanSpaces(prevVal), -1)
	condition = strings.Replace(condition, TagLastVal, cleanSpaces(lastVal), -1)
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
			actorParam = strings.Replace(actorParam, TagSensorName, sensorName, -1)
			actorParam = strings.Replace(actorParam, TagPrevVal, prevVal, -1)
			actorParam = strings.Replace(actorParam, TagLastVal, lastVal, -1)
			actorParam = strings.Replace(actorParam, TagCondition, condition, -1)
			actorParam = cleanSpaces(actorParam)
		}

		actorId, err := sensorAct.getIntVal("idActor")
		if err != nil {
			return
		}

		if glog.V(1) {
			glog.Infof("triggerSensorAct : launching Actor #%d", actorId)
		}

		triggerActorById(actorId, 1, actorParam)
	}

}

// -----------------------------------------------
// -----------------------------------------------

var cpuReadNs int64
var cpuReadVal int

func readProcStat() (nano int64, read int, nb int, err error) {
	nano = time.Now().UnixNano()
	f, err := os.Open("/proc/stat")
	if err != nil {
		glog.Errorf("CpuUsage : %s", err)
		return
	}
	defer f.Close()

	var key string
	var u, n, s int
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		_, err = fmt.Sscanf(line, "%s %d %d %d", &key, &u, &n, &s)
		if err != nil {
			glog.Errorf("Scan fail '%s' : %s", line, err)
			return
		}
		if key == "cpu" {
			read = u + n + s
		} else if strings.HasPrefix(key, "cpu") {
			nb++
		} else {
			break
		}
	}
	return
}

func CpuUsage(param1 string, param2 string) (string, error) {
	nano, read, nb, err := readProcStat()
	if err != nil {
		return "99", err
	}

	// If previous mesure was more than 60s ago
	if nano-cpuReadNs > 60000000 {
		cpuReadNs = nano
		cpuReadVal = read
		time.Sleep(time.Second)
		nano, read, nb, err = readProcStat()
		if err != nil {
			return "99", err
		}
	}

	load := float32(int64(read-cpuReadVal)*1000000000.0) / float32((nano-cpuReadNs)*int64(nb)*1.0)
	load += 0.5 // for rounding when converting with %.0f

	if glog.V(1) {
		glog.Infof("CpuUsage %3.2f = %d / ( %d * %d )", load, (read-cpuReadVal)*1000000000, nano-cpuReadNs, nb)
	}

	return fmt.Sprintf("%.0f", load), nil
}

func MemoryUsage(param1 string, param2 string) (string, error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		glog.Errorf("CpuUsage : %s", err)
		return "99", err
	}
	defer f.Close()

	var key string
	var val, total, free int
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		_, err = fmt.Sscanf(line, "%s %d", &key, &val)
		if err != nil {
			glog.Errorf("Scan fail '%s' : %s", line, err)
			return "99", err
		}
		if key == "MemTotal:" {
			total = val
		} else if key == "MemFree:" {
			free = val
		} else {
			break
		}
	}

	used := float32((total-free)*100) / float32(total)

	if glog.V(1) {
		glog.Infof("MemoryUsage %3.2f = %d / %d", used, total, free)
	}

	return fmt.Sprintf("%.0f", used), nil
}
