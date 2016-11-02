// internal.go
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/golang/glog"
)

var actorsMapLock sync.Mutex
var actorsMap = map[string]HomeObject{}

func init() {
	RegisterInternalFunc(ActorFunc, "GPIO", ActorGPIO)
	RegisterInternalFunc(ActorFunc, "SerialATSMS", SerialATSMS)
}

// actorSetup : read defined actors from DB then create a ticker and start reading goroutine for each actor
func actorSetup(db *sql.DB) (err error) {

	actorObjs, err := getHomeObjects(db, -1, -1, ItemActor)
	if err != nil {
		return
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

	if glog.V(1) {
		glog.Infof("actorSetup Done (%d)", len(actorsMap))
	}

	return
}

// triggerActorById : trigger actor function using ActCmd, restirered parameter 'ActParam' and dynamic param 'param'
func triggerActorById(actorId int, param string) (result string, err error) {
	actorsMapLock.Lock()
	defer actorsMapLock.Unlock()
	for _, actor := range actorsMap {
		if actor.getId() == actorId {
			result, err = triggerObjActor(actor, param)
			return
		}
	}
	err = errors.New(fmt.Sprintf("No known actor with id = %d", actorId))
	glog.Error(err)

	return
}

// triggerActorByName : trigger actor function using ActCmd, restirered parameter 'ActParam' and dynamic param 'param'
func triggerActorByName(actorName string, param string) (result string, err error) {
	actorsMapLock.Lock()
	actor, found := actorsMap[actorName]
	actorsMapLock.Unlock()
	if !found {
		err = errors.New(fmt.Sprintf("No known actor '%s'", actorName))
		glog.Error(err)
		return
	}
	result, err = triggerObjActor(actor, param)
	return
}

// triggerObjActor : trigger actor function using ActCmd, restirered parameter 'ActParam' and dynamic param 'param'
func triggerObjActor(actor HomeObject, param string) (result string, err error) {
	result = "Failed"
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
	if glog.V(1) {
		glog.Info("actorCleanup Done")
	}
}

// recordActorResult : store in DB param and result for an actor
// TODO : allow to store value outsite of main DB ?
func recordActorResult(actor HomeObject, param string, result string) {
	db, err := openDB()
	if err != nil {
		return
	}
	defer db.Close()

	actorId := actor.getId()

	_, err = db.Exec("insert into HistoActor values ( ?, ?, ?, ?);", time.Now().Unix(), actorId, param, result)
	if err != nil {
		glog.Errorf("Fail to store result (%s) for actor %d from : %s ", result, actorId, err)
	}
	if glog.V(1) {
		glog.Infof("recordActorResult : %d - %s - %s", time.Now().Unix(), param, result)
	}
}

// -----------------------------------------------
// -----------------------------------------------

func ActorGPIO(param1 string, param2 string) (string, error) {
	// TODO ActorGPIO
	glog.Info("GPIO Not Implemented")
	return time.Now().String(), nil
}

func SerialATSMS(param1 string, param2 string) (string, error) {
	// TODO SerialATSMS
	glog.Info("SerialATSMS Not Implemented")
	return time.Now().String(), nil
}