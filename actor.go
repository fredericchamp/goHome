// internal.go
package main

import (
	"database/sql"
	"sync"
	"time"

	"github.com/golang/glog"
)

var actorsMapLock sync.Mutex
var actorsMap = map[string]HomeObject{}

func init() {
	RegisterInternalFunc(ActorFunc, "GPIO", ActorGPIO)
	RegisterInternalFunc(ActorFunc, "SerialATSMS", SerialATSMS)
	//	actorsMap = make(map[string]HomeObject)
}

// actorSetup : read defined actors from DB then create a ticker and start reading goroutine for each actor
func actorSetup(db *sql.DB) (err error) {
	actorItems, err := getManageItems(db, -1, ItemActor)
	if err != nil {
		return
	}
	if glog.V(2) {
		glog.Info("\nActor Items\n", actorItems)
	}

	var actorObjs []HomeObject
	for _, v := range actorItems {
		lst, err := getDBObjects(db, -1, v.Id)
		if err != nil {
			return err
		}
		for _, v := range lst {
			actorObjs = append(actorObjs, v)
		}
	}
	if glog.V(2) {
		glog.Info("\nActor\n", actorObjs)
	}

	actorsMapLock.Lock()
	defer actorsMapLock.Unlock()

	for _, actor := range actorObjs {

		isActive, err := actor.getIntVal("IsActive")
		if err != nil {
			return err
		}
		if isActive == 0 {
			continue
		}

		actorName, err := actor.getStrVal("Name")
		if err != nil {
			return err
		}

		actorsMap[actorName] = actor
	}

	return
}

// triggerActor : trigger actor function using ActCmd, restirered parameter 'ActParam' and dynamic param 'param'
func triggerActor(actorName string, param string) (result string, err error) {
	result = "Failed"
	actorsMapLock.Lock()
	actor, found := actorsMap[actorName]
	actorsMapLock.Unlock()
	if !found {
		glog.Errorf("No known actor '%s'", actorName)
		return
	}

	actCmd, err := actor.getStrVal("ActCmd")
	if err != nil {
		return
	}
	actParam, err := actor.getStrVal("ActParam")
	if err != nil {
		return
	}
	isInternal, err := actor.getIntVal("IsInternal")
	if err != nil {
		return
	}

	if isInternal != 0 {
		result, err = CallInternalFunc(ActorFunc, actCmd, actParam, param)
	} else {
		result, err = LaunchExternalCmd(ActorFunc, actCmd, actParam, param)
	}

	go recordActorResult(actor, param, result)

	return
}

// actorCleanup : stop and remove all actor ticker
func actorCleanup() {
	actorsMapLock.Lock()
	actorsMap = make(map[string]HomeObject)
	actorsMapLock.Unlock()
	if glog.V(2) {
		glog.Info("actorCleanup Done")
	}
}

// recordActorValue : store in DB param and result for an actor
// TODO : allow to store value outsite of main DB ?
func recordActorResult(actor HomeObject, param string, result string) {
	db, err := openDB()
	if err != nil {
		glog.Error(err)
		return
	}
	defer db.Close()

	actorId := actor.getId()

	_, err = db.Exec("insert into HistoActor values ( ?, ?, ?, ?);", time.Now().Unix(), actorId, param, result)
	if err != nil {
		glog.Errorf("Fail to store result (%s) for actor %d from : %s ", result, actorId, err)
	}
	if glog.V(2) {
		glog.Infof("recordActorValue : %d - %s - %s", time.Now().Unix(), param, result)
	}
}

// -----------------------------------------------
// -----------------------------------------------

func ActorGPIO(param1 string, param2 string) (string, error) {
	// TODO
	glog.Info("GPIO Not Implemented")
	return time.Now().String(), nil
}

func SerialATSMS(param1 string, param2 string) (string, error) {
	// TODO
	glog.Info("SerialATSMS Not Implemented")
	return time.Now().String(), nil
}
