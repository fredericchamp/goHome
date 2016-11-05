// user.go
package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/golang/glog"
)

// -----------------------------------------------

// Accepted Format = {"command":"api...", "itemtypeid":id, "itemid":id, "objectid":id, "startts":ts, "endts":ts, "jsonparam":{...}}

type apiCommand string

const (
	apiReadItemType  apiCommand = "ReadItemTypes"
	apiReadItem                 = "ReadItems"
	apiReadObject               = "ReadObject"
	apiReadSensor               = "ReadSensor"
	apiReadHistoVal             = "ReadHistoVal"
	apiReadActorRes             = "ReadActorRes"
	apiSaveItem                 = "SaveItems"
	apiSaveObject               = "SaveObject"
	apiDeleteItem               = "DeleteItems"
	apiDeleteObject             = "DeleteObject"
	apiSendSensorVal            = "SendSensorVal"
	apiTriggerActor             = "TriggerActor"
)

type apiCommandSruct struct {
	Command    apiCommand
	Itemtypeid itemType
	Itemid     int
	Objectid   int
	Startts    int64
	Endts      int64
	Jsonparam  string
}

// -----------------------------------------------

func apiError(errMsg string) (apiResp []byte) {
	apiResp, err := json.Marshal(struct{ Error string }{errMsg})
	if err != nil {
		glog.Errorf("json.Marshal Failed for error message '%s'", errMsg)
		apiResp = []byte(`{"error":"Error (json.Marshal Failed to error message)"}`)
	}
	return
}

func fctApiReadItem(profil userProfil, jsonCmde apiCommandSruct) (apiResp []byte) {
	items, err := getManageItems(nil, jsonCmde.Itemtypeid, jsonCmde.Itemid)
	if err != nil {
		apiResp = apiError(fmt.Sprintf("%s failed for (type=%d, item=%d) : %s", jsonCmde.Command, jsonCmde.Itemtypeid, jsonCmde.Itemid, err))
		return
	}

	apiResp, err = json.Marshal(profilFilteredItems(profil, items))
	if err != nil {
		apiResp = apiError(fmt.Sprintf("%s failed for (type=%d, item=%d) : %s", jsonCmde.Command, jsonCmde.Itemtypeid, jsonCmde.Itemid, err))
		return
	}
	return
}

func fctApiReadObject(profil userProfil, jsonCmde apiCommandSruct) (apiResp []byte) {
	objs, err := getHomeObjects(nil, jsonCmde.Itemtypeid, jsonCmde.Itemid, jsonCmde.Objectid)
	if err != nil {
		apiResp = apiError(fmt.Sprintf("%s failed for (type=%d, item=%d, obj=%d) : %s", jsonCmde.Command, jsonCmde.Itemtypeid, jsonCmde.Itemid, jsonCmde.Objectid, err))
		return
	}

	apiResp, err = json.Marshal(profilFilteredObjects(profil, objs))
	if err != nil {
		apiResp = apiError(fmt.Sprintf("%s failed for (type=%d, item=%d, obj=%d) : %s", jsonCmde.Command, jsonCmde.Itemtypeid, jsonCmde.Itemid, jsonCmde.Objectid, err))
		return
	}
	return
}

func fctApiReadSensor(profil userProfil, jsonCmde apiCommandSruct) (apiResp []byte) {
	objs, err := getHomeObjects(nil, ItemNone, -1, jsonCmde.Objectid)
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

	value, err := readSensoValue(sensor)
	if err != nil {
		apiResp = apiError(fmt.Sprintf("%s failed for (obj=%d) : %s", jsonCmde.Command, jsonCmde.Objectid, err))
		return
	}

	apiResp, err = json.Marshal(value)
	if err != nil {
		apiResp = apiError(fmt.Sprintf("%s failed for (obj=%d) : %s", jsonCmde.Command, jsonCmde.Objectid, err))
		return
	}
	return
}

func fctApiReadHistoVal(profil userProfil, jsonCmde apiCommandSruct) (apiResp []byte) {
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

func fctApiReadActorRes(profil userProfil, jsonCmde apiCommandSruct) (apiResp []byte) {
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

func fctApiSaveObject(profil userProfil, jsonCmde apiCommandSruct) (apiResp []byte) {
	var objIn HomeObject
	if err := json.Unmarshal([]byte(jsonCmde.Jsonparam), &objIn); err != nil {
		apiResp = apiError(fmt.Sprintf("%s fail to unmarshal jsonparam (%s) : %s", jsonCmde.Command, jsonCmde.Jsonparam, err))
		return
	}
	if glog.V(3) {
		glog.Infof("Object to save \n%+v", objIn)
	}

	if glog.V(2) {
		glog.Info("fctApiSaveObject : check object validity")
	}
	if err := objIn.ValidateValues(objIn.Values); err != nil {
		apiResp = apiError(fmt.Sprintf("%s : %s", jsonCmde.Command, err.Error()))
		return
	}
	// check profil access rights on new object
	if err := checkAccessToObject(profil, objIn); err != nil {
		apiResp = apiError(err.Error())
		return
	}

	// Must not use jsonCmde.Objectid, just in case jsonCmde.Objectid != objIn.Values[0].
	objectid := objIn.Values[0].IdObject

	var objFields []ItemField

	if objectid > 0 {
		// if objectid > 0 it's an UPDATE => fetch existing object
		objs, err := getHomeObjects(nil, jsonCmde.Itemtypeid, jsonCmde.Itemid, objectid)
		if err != nil {
			apiResp = apiError(fmt.Sprintf("%s fail to load matching object (%s) : %s", jsonCmde.Command, objectid, err))
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
		objFields = objs[0].Fields
	} else {
		// else it's an INSERT => fetch fields definition
		fields, err := getItemFields(nil, jsonCmde.Itemid, objectid)
		if err != nil {
			apiResp = apiError(err.Error())
			return
		}
		objFields = fields
	}

	// TODO check profil access rights on item holding object to save ?
	//	if err = checkAccessToObjectId(profil, objDb.Fields[0].IdItem); err != nil {
	//		apiResp = apiError(err.Error())
	//		return
	//	}

	// Check objIn fields match objDb fields
	if !reflect.DeepEqual(objIn.Fields, objFields) {
		apiResp = apiError(fmt.Sprintf("%s received []Fields does not match []Fields in DB for itemid=%d", jsonCmde.Command, jsonCmde.Itemid))
		return
	}

	// write object to DB
	objectid, err := writeObject(objIn)
	if err != nil {
		apiResp = apiError(err.Error())
		return
	}

	// return saved object
	jsonCmde.Itemtypeid = ItemNone
	jsonCmde.Itemid = -1
	jsonCmde.Objectid = objectid
	if glog.V(2) {
		glog.Infof("fctApiReadObject for id=%d", jsonCmde.Objectid)
	}
	apiResp = fctApiReadObject(profil, jsonCmde)

	return
}
