// api.go
package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang/glog"
)

// -----------------------------------------------

// Accepted Format = {"command":"api...", "itemid":id, "objectid":id, "startts":ts, "endts":ts, "jsonparam":{...}}

type apiCommand string

const (
	apiReadRefList      apiCommand = "ReadRefList"
	apiReadCurrentUser             = "ReadCurrentUser"
	apiReadItem                    = "ReadItem"
	apiReadObject                  = "ReadObject"
	apiReadSensor                  = "ReadSensor"
	apiGetSensorLastVal            = "GetSensorLastVal"
	apiReadHistoVal                = "ReadHistoVal"
	apiReadActorRes                = "ReadActorRes"
	apiSaveItem                    = "SaveItems"
	apiSaveObject                  = "SaveObject"
	apiDeleteItem                  = "DeleteItems"
	apiDeleteObject                = "DeleteObject"
	apiSendSensorVal               = "SendSensorVal"
	apiTriggerActor                = "TriggerActor"
)

type apiCommandSruct struct {
	Command   apiCommand
	Itemid    TItemId
	Objectid  int
	Startts   int64
	Endts     int64
	Jsonparam string
	UserCode  string
}

// -----------------------------------------------

func apiResponse(msgName string, msgText string) (apiResp []byte) {
	jsonMsg, err := json.Marshal(msgText)
	if err != nil {
		glog.Errorf("json.Marshal Failed for response message '%s'", msgText)
		apiResp = apiError("Error (json.Marshal Failed for response message)")
	}
	apiResp = []byte(fmt.Sprintf(`{"%s":%s}`, msgName, jsonMsg))
	return
}

func apiError(errMsg string) (apiResp []byte) {
	apiResp = apiResponse("error", errMsg)
	glog.Error(errMsg)
	return
}

func apiObjectResponse(profil TUserProfil, obj HomeObject) (apiResp []byte) {
	if err := checkAccessToObject(profil, obj); err != nil {
		apiResp = apiError(fmt.Sprintf("apiObjectResponse failed : %s", err))
		return
	}
	apiResp, err := json.Marshal(obj)
	if err != nil {
		apiResp = apiError(fmt.Sprintf("apiObjectResponse failed : %s", err))
		return
	}
	return
}

func fctApiRefList(jsonCmde apiCommandSruct) (apiResp []byte) {
	list, err := getRefList(nil, jsonCmde.Jsonparam)
	if err != nil {
		apiResp = apiError(fmt.Sprintf("%s failed for (%s) : %s", jsonCmde.Command, jsonCmde.Jsonparam, err))
		return
	}

	apiResp, err = json.Marshal(list)
	if err != nil {
		apiResp = apiError(fmt.Sprintf("%s failed for (item=%d) : %s", jsonCmde.Command, jsonCmde.Itemid, err))
		return
	}
	return
}

func fctApiReadItem(profil TUserProfil, jsonCmde apiCommandSruct) (apiResp []byte) {
	items, err := getManageItems(nil, jsonCmde.Itemid)
	if err != nil {
		apiResp = apiError(fmt.Sprintf("%s failed for (item=%d) : %s", jsonCmde.Command, jsonCmde.Itemid, err))
		return
	}

	apiResp, err = json.Marshal(profilFilteredItems(profil, items))
	if err != nil {
		apiResp = apiError(fmt.Sprintf("%s failed for (item=%d) : %s", jsonCmde.Command, jsonCmde.Itemid, err))
		return
	}
	return
}

func fctApiReadObject(profil TUserProfil, jsonCmde apiCommandSruct) (apiResp []byte) {
	objs, err := getHomeObjects(nil, jsonCmde.Itemid, jsonCmde.Objectid)
	if err != nil {
		apiResp = apiError(fmt.Sprintf("%s failed for (item=%d, obj=%d) : %s", jsonCmde.Command, jsonCmde.Itemid, jsonCmde.Objectid, err))
		return
	}

	apiResp, err = json.Marshal(profilFilteredObjects(profil, objs))
	if err != nil {
		apiResp = apiError(fmt.Sprintf("%s failed for (item=%d, obj=%d) : %s", jsonCmde.Command, jsonCmde.Itemid, jsonCmde.Objectid, err))
		return
	}
	return
}

func fctApiGetSensorVal(profil TUserProfil, jsonCmde apiCommandSruct, read bool) (apiResp []byte) {
	objs, err := getHomeObjects(nil, ItemIdNone, jsonCmde.Objectid)
	if err != nil {
		apiResp = apiError(fmt.Sprintf("%s failed for (obj=%d) : %s", jsonCmde.Command, jsonCmde.Objectid, err))
		return
	}
	if len(objs) <= 0 {
		apiResp = apiError(fmt.Sprintf("%s failed for (obj=%d) : object not found", jsonCmde.Command, jsonCmde.Objectid))
		return
	}
	sensor := objs[0]

	err = checkAccessToObject(profil, sensor)
	if err != nil {
		apiResp = apiError(err.Error())
		return
	}

	var value string
	if read {
		value, err = readSensorValue(sensor)
	} else {
		value, err = getSensorLastValue(sensor)
	}
	if err != nil {
		apiResp = apiError(fmt.Sprintf("%s failed for (obj=%d) : %s", jsonCmde.Command, jsonCmde.Objectid, err))
		return
	}

	apiResp, err = json.Marshal(HistoSensor{time.Now(), jsonCmde.Objectid, value})
	if err != nil {
		apiResp = apiError(fmt.Sprintf("%s failed for (obj=%d) : %s", jsonCmde.Command, jsonCmde.Objectid, err))
		return
	}
	return
}

func fctApiReadHistoVal(profil TUserProfil, jsonCmde apiCommandSruct) (apiResp []byte) {
	last := false
	if (jsonCmde.Startts <= 0 && jsonCmde.Endts <= 0) || jsonCmde.Startts > time.Now().Unix() {
		last = true
	}

	err := checkAccessToObjectId(profil, jsonCmde.Objectid)
	if err != nil {
		apiResp = apiError(err.Error())
		return
	}

	sVals, err := getHistoSensor(nil, jsonCmde.Objectid, last, time.Unix(jsonCmde.Startts, 0), time.Unix(jsonCmde.Endts, 0))
	if err != nil {
		apiResp = apiError(fmt.Sprintf("%s failed for (obj=%d, start=%d, end=%d) : %s", jsonCmde.Command, jsonCmde.Objectid, jsonCmde.Startts, jsonCmde.Endts, err))
		return
	}

	apiResp, err = json.Marshal(sVals)
	if err != nil {
		apiResp = apiError(fmt.Sprintf("%s failed for (obj=%d, start=%d, end=%d) : %s", jsonCmde.Command, jsonCmde.Objectid, jsonCmde.Startts, jsonCmde.Endts, err))
		return
	}

	return
}

func fctApiReadActorRes(profil TUserProfil, jsonCmde apiCommandSruct) (apiResp []byte) {
	last := false
	if (jsonCmde.Startts <= 0 && jsonCmde.Endts <= 0) || jsonCmde.Startts > time.Now().Unix() {
		last = true
	}

	err := checkAccessToObjectId(profil, jsonCmde.Objectid)
	if err != nil {
		apiResp = apiError(err.Error())
		return
	}

	aVals, err := getHistActor(nil, jsonCmde.Objectid, last, time.Unix(jsonCmde.Startts, 0), time.Unix(jsonCmde.Endts, 0))
	if err != nil {
		apiResp = apiError(fmt.Sprintf("%s failed for (obj=%d, start=%d, end=%d) : %s", jsonCmde.Command, jsonCmde.Objectid, jsonCmde.Startts, jsonCmde.Endts, err))
		return
	}

	apiResp, err = json.Marshal(aVals)
	if err != nil {
		apiResp = apiError(fmt.Sprintf("%s failed for (obj=%d, start=%d, end=%d) : %s", jsonCmde.Command, jsonCmde.Objectid, jsonCmde.Startts, jsonCmde.Endts, err))
		return
	}

	return
}

func fctApiSaveObject(profil TUserProfil, jsonCmde apiCommandSruct) (apiResp []byte) {
	var objIn, objPrev HomeObject
	if err := json.Unmarshal([]byte(jsonCmde.Jsonparam), &objIn); err != nil {
		apiResp = apiError(fmt.Sprintf("%s fail to unmarshal jsonparam (%s) : %s", jsonCmde.Command, jsonCmde.Jsonparam, err))
		return
	}
	if glog.V(3) {
		glog.Infof("Object to save \n%+v", objIn)
	}

	// check profil access rights on new object
	if err := checkAccessToObject(profil, objIn); err != nil {
		apiResp = apiError(err.Error())
		return
	}

	// Must not use jsonCmde.Objectid, just in case jsonCmde.Objectid != objIn.Values[0].
	objectid := objIn.Values[0].IdObject

	// fetch Fields definition from DB, and ignore received fields definition if any
	if objectid > 0 {
		// if objectid > 0 it's an UPDATE => fetch existing object
		objs, err := getHomeObjects(nil, jsonCmde.Itemid, objectid)
		if err != nil {
			apiResp = apiError(fmt.Sprintf("%s fail to load matching object (%d) : %s", jsonCmde.Command, objectid, err))
			return
		}
		if len(objs) != 1 {
			apiResp = apiError(fmt.Sprintf("%s should have only 1 matching object, not %d", jsonCmde.Command, len(objs)))
			return
		}
		if glog.V(2) {
			if glog.V(3) {
				glog.Infof("Object to update \n%+v", objs[0])
			}
			glog.Infof("checkAccessToObject(profil, objs[0]) id=%d", objectid)
		}
		// check profil access rights on existing object
		if err = checkAccessToObject(profil, objs[0]); err != nil {
			apiResp = apiError(err.Error())
			return
		}
		objPrev = objs[0]
		objIn.Fields = objs[0].Fields
	} else {
		// else it's an INSERT => fetch fields definition
		fields, err := getItemFields(nil, jsonCmde.Itemid, objectid)
		if err != nil {
			apiResp = apiError(err.Error())
			return
		}
		objIn.Fields = fields
	}

	if glog.V(2) {
		glog.Info("fctApiSaveObject : check object validity")
	}
	if err := objIn.ValidateValues(objIn.Values); err != nil {
		apiResp = apiError(fmt.Sprintf("%s : %s", jsonCmde.Command, err.Error()))
		return
	}

	// write object to DB
	objectid, err := writeObject(objIn)
	if err != nil {
		apiResp = apiError(err.Error())
		return
	}

	// Update in-memory data with new object / new object values
	switch objIn.Fields[0].IdItem {
	case ItemUser:
		go loadUsers(nil, true)
		break
	case ItemSensor:
		err = sensorUpdateTicker(objIn)
		if err != nil {
			glog.Errorf("fctApiSaveObject : sensor #%d update failed : %s", objIn.Values[0].IdObject, err)
		}
		break
	case ItemSensorAct:
		masterid, err := objIn.getIntVal("idMasterObj")
		if err != nil {
			glog.Errorf("fctApiSaveObject : idMasterObj fail : %s", err)
		}
		if objectid > 0 {
			// It's an update, check if master have change
			masteridprev, err := objPrev.getIntVal("idMasterObj")
			if err != nil {
				glog.Errorf("fctApiSaveObject : idMasterObj prev fail : %s", err)
			}
			if masterid != masteridprev {
				// Prev master need update so current snesorAct get remove from it list
				sensors, err := getHomeObjects(nil, ItemIdNone, masteridprev)
				if err != nil {
					glog.Errorf("fctApiSaveObject : sensorAct, read sensor %d fail : %s", masteridprev, err)
				}
				err = sensorUpdateTicker(sensors[0])
				if err != nil {
					glog.Errorf("fctApiSaveObject : sensor %d update failed : %s", masteridprev, err)
				}
			}
		}
		sensors, err := getHomeObjects(nil, ItemIdNone, masterid)
		if err != nil {
			glog.Errorf("fctApiSaveObject : sensorAct, read sensor %d fail : %s", masterid, err)
		}
		err = sensorUpdateTicker(sensors[0])
		if err != nil {
			glog.Errorf("fctApiSaveObject : sensor %d update failed : %s", masterid, err)
		}
		break
	}

	// return saved object
	jsonCmde.Itemid = ItemIdNone
	jsonCmde.Objectid = objectid
	if glog.V(2) {
		glog.Infof("fctApiReadObject for id=%d", jsonCmde.Objectid)
	}
	apiResp = fctApiReadObject(profil, jsonCmde)

	return
}

func fctApiTriggerActor(profil TUserProfil, userId int, jsonCmde apiCommandSruct) (apiResp []byte) {
	if err := checkAccessToObjectId(profil, jsonCmde.Objectid); err != nil {
		apiResp = apiError(err.Error())
		return
	}

	result, err := triggerActorById(jsonCmde.Objectid, userId, jsonCmde.Jsonparam)
	if err != nil {
		apiResp = apiError(err.Error())
		return
	}

	apiResp = apiResponse("response", result)

	return

}
