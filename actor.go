// actor.go
package main

import (
	"database/sql"
	"errors"
	"fmt"
	//	"sync"
	"time"

	"github.com/golang/glog"
)

//var actorsMapLock sync.Mutex
//var actorsMap = map[string]HomeObject{}

func init() {
	RegisterInternalFunc(ActorFunc, "SerialATSMS", SerialATSMS)
}

// actorSetup : read defined actors from DB then TODO
func actorSetup(db *sql.DB) (err error) {

	//	actorObjs, err := getHomeObjects(db, ItemActor, -1)
	//	if err != nil {
	//		return
	//	}
	//	if glog.V(2) {
	//		glog.Info("\nActor\n", actorObjs)
	//	}

	//	actorsMapLock.Lock()
	//	defer actorsMapLock.Unlock()

	//	for _, actor := range actorObjs {

	//		isActive, err := actor.getIntVal("IsActive")
	//		if err != nil {
	//			return err
	//		}
	//		if isActive == 0 {
	//			continue
	//		}

	//		actorName, err := actor.getStrVal("Name")
	//		if err != nil {
	//			return err
	//		}

	//		actorsMap[actorName] = actor
	//	}

	//	if glog.V(1) {
	//		glog.Infof("actorSetup Done (%d)", len(actorsMap))
	//	}

	return
}

// triggerActorById : trigger actor function using ActCmd, restirered parameter 'ActParam' and dynamic param 'param'
func triggerActorById(actorId int, userId int, param string) (result string, err error) {
	//	actorsMapLock.Lock()
	//	defer actorsMapLock.Unlock()
	//	for _, actor := range actorsMap {
	//		if actor.getId() == actorId {
	//			result, err = triggerObjActor(actor, userId, param)
	//			return
	//		}
	//	}
	//	err = errors.New(fmt.Sprintf("No known actor with id = %d", actorId))
	//	glog.Error(err)

	objs, err := getHomeObjects(nil, ItemIdNone, actorId)
	if err != nil {
		return
	}
	if len(objs) <= 0 {
		err = errors.New(fmt.Sprintf("No actor with id = %d", actorId))
		glog.Error(err)
	}
	actor := objs[0]

	result, err = triggerObjActor(actor, userId, param)
	return
}

// triggerActorByName : trigger actor function using ActCmd, registered parameter 'ActParam' and dynamic param 'param'
//func triggerActorByName(actorName string, userId int, param string) (result string, err error) {
//	actorsMapLock.Lock()
//	actor, found := actorsMap[actorName]
//	actorsMapLock.Unlock()
//	if !found {
//		err = errors.New(fmt.Sprintf("No known actor '%s'", actorName))
//		glog.Error(err)
//		return
//	}
//	result, err = triggerObjActor(actor, userId, param)
//	return
//}

// triggerObjActor : trigger actor function using ActCmd, registered parameter 'ActParam' and dynamic param 'param'
func triggerObjActor(actor HomeObject, userId int, param string) (result string, err error) {
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

	go recordActorResult(actor, userId, param, result)

	return
}

// actorCleanup : stop and remove all actor ticker
func actorCleanup() {
	//	actorsMapLock.Lock()
	//	actorsMap = make(map[string]HomeObject)
	//	actorsMapLock.Unlock()
	//	if glog.V(1) {
	//		glog.Info("actorCleanup Done")
	//	}
}

// recordActorResult : store in DB param and result for an actor
func recordActorResult(actor HomeObject, userId int, param string, result string) {
	db, err := openDB()
	if err != nil {
		return
	}
	defer db.Close()

	actorId := actor.getId()

	_, err = db.Exec("insert into HistoActor values ( ?, ?, ?, ?, ?);", time.Now().Unix(), actorId, userId, param, result)
	if err != nil {
		glog.Errorf("Fail to store result (%s) for actor %d : %s ", result, actorId, err)
	}
	if glog.V(1) {
		glog.Infof("recordActorResult : %d - %s - %s", time.Now().Unix(), param, result)
	}
}

// -----------------------------------------------
// -----------------------------------------------

func SerialATSMS(param1 string, param2 string) (string, error) {
	glog.Info("SerialATSMS Not Implemented") // TODO
	return time.Now().String(), nil
}
