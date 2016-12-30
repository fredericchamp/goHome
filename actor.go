// actor.go
package main

import (
	//"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/golang/glog"
)

func init() {
	RegisterInternalFunc(ActorFunc, "SerialATSMS", SerialATSMS)
}

// triggerActorById : trigger actor function using ActCmd, restirered parameter 'ActParam' and dynamic param 'param'
func triggerActorById(actorId int, userId int, param string) (result string, err error) {

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
		result, err = ExecExternalCmd(actCmd, actParam, param)
	}

	go recordActorResult(actor, userId, param, result)

	return
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
