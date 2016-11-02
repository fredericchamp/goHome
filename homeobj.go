// homeobj.go
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/golang/glog"
	_ "github.com/mattn/go-sqlite3"
)

// -----------------------------------------------

type itemType int

const (
	ItemNone itemType = iota
	ItemEntity
	ItemSensor
	ItemActor
	ItemSensorAct
	ItemVideoSensor // TODO
)

var itemTypeNames = map[itemType]string{
	ItemNone:        "None",
	ItemEntity:      "Entity",
	ItemSensor:      "Sensor",
	ItemActor:       "Actor",
	ItemSensorAct:   "Actor trigger by sensor",
	ItemVideoSensor: "VideoSensor",
}

// -----------------------------------------------

const (
	DBTypeBool = 1 + iota
	DBTypeInt
	DBTypeFloat
	DBTypeText
	DBTypeDateTime
	DBTypeBlob
)

var dbTypeNames = map[int]string{
	DBTypeBool:     "Boolean",
	DBTypeInt:      "Integer",
	DBTypeFloat:    "Float",
	DBTypeText:     "Text",
	DBTypeDateTime: "DateTime",
	DBTypeBlob:     "Bytes",
}

// -----------------------------------------------

type HomeObject struct {
	Fields     []ItemField
	Values     []ItemFieldVal
	linkedObjs []HomeObject
}

// getId : return objectId if exist or -1
func (obj HomeObject) getId() int {
	if len(obj.Values) <= 0 {
		glog.Error("Error trying to get Id from an HomeObject with empty []Values")
		return -1
	}
	return obj.Values[0].IdObject
}

// hasField : return true if fieldName is present
func (obj HomeObject) hasField(fieldName string) bool {
	for _, v := range obj.Fields {
		if v.Name == fieldName {
			return true
		}
	}
	return false
}

// getFieldIndex : return field index of fieldName if present else -1
func (obj HomeObject) getFieldIndex(fieldName string) (int, error) {
	for i, v := range obj.Fields {
		if v.Name == fieldName {
			return i, nil
		}
	}
	err := errors.New(fmt.Sprintf("Field '%s' not found", fieldName))
	glog.Error(err)
	if glog.V(1) {
		glog.Info(obj)
	}
	return -1, err
}

// getIntVal : return integer value for fieldName if possible else err
func (obj HomeObject) getIntVal(fieldName string) (value int, err error) {
	idx, err := obj.getFieldIndex(fieldName)
	if err != nil {
		return
	}
	switch obj.Fields[idx].IdDataType {
	case DBTypeBool, DBTypeInt, DBTypeDateTime:
		value = obj.Values[idx].IntVal
	case DBTypeText:
		value, err = strconv.Atoi(obj.Values[idx].TextVal)
	case DBTypeFloat:
		err = errors.New(fmt.Sprintf("Not converting float to int for '%s' field", fieldName))
	case DBTypeBlob:
		err = errors.New(fmt.Sprintf("Not converting blob to int for '%s' field", fieldName))
	default:
		err = errors.New(fmt.Sprintf("Unknown data type %d for '%s' field", obj.Fields[idx].IdDataType, fieldName))
	}
	if err != nil {
		glog.Error(err)
		if glog.V(1) {
			glog.Info(obj)
		}
	}
	return
}

// getStrVal : return string value for fieldName if possible else err
func (obj HomeObject) getStrVal(fieldName string) (value string, err error) {
	idx, err := obj.getFieldIndex(fieldName)
	if err != nil {
		return
	}
	switch obj.Fields[idx].IdDataType {
	case DBTypeBool:
		if obj.Values[idx].IntVal == 0 {
			value = "No"
		} else {
			value = "Yes"
		}
	case DBTypeInt, DBTypeDateTime:
		value = fmt.Sprint(obj.Values[idx].IntVal)
	case DBTypeFloat:
		value = fmt.Sprint(obj.Values[idx].FloatVal)
	case DBTypeText:
		value = obj.Values[idx].TextVal
	case DBTypeBlob:
		err = errors.New(fmt.Sprintf("Not converting blob to string for '%s' field", fieldName))
	default:
		err = errors.New(fmt.Sprintf("Unknown data type %d for '%s' field", obj.Fields[idx].IdDataType, fieldName))
	}
	if err != nil {
		glog.Error(err)
		if glog.V(1) {
			glog.Info(obj)
		}
	}
	return
}

// getByteVal : return []byte value for fieldName if possible else err
func (obj HomeObject) getByteVal(fieldName string) (value []byte, err error) {
	idx, err := obj.getFieldIndex(fieldName)
	if err != nil {
		return
	}
	switch obj.Fields[idx].IdDataType {
	case DBTypeBool:
		if obj.Values[idx].IntVal == 0 {
			value = []byte("No")
		} else {
			value = []byte("Yes")
		}
	case DBTypeInt, DBTypeFloat, DBTypeDateTime:
		value = []byte(fmt.Sprint(obj.Values[idx].IntVal))
	case DBTypeText:
		value = []byte(obj.Values[idx].TextVal)
	case DBTypeBlob:
		value = obj.Values[idx].ByteVal
	default:
		err = errors.New(fmt.Sprintf("Unknown data type %d for '%s' field", obj.Fields[idx].IdDataType, fieldName))
	}
	if err != nil {
		glog.Error(err)
		if glog.V(1) {
			glog.Info(obj)
		}
	}
	return
}

// -----------------------------------------------

// getLinkedObjects add to each objs[] corresponding linked objects
// TODO : reorg/optimisation : currently need 1 db query for each objs[] + 1 db query for each linked obj
func getLinkedObjects(db *sql.DB, objs []HomeObject) error {
	if db == nil {
		db, err := openDB()
		if err != nil {
			return err
		}
		defer db.Close()
	}

	for i, _ := range objs {
		lstLinkedObjId, err := getLinkedObjIds(db, objs[i].getId())
		if err != nil {
			return err
		}
		for _, linkedObjId := range lstLinkedObjId {
			linkedObjs, err := getHomeObjects(db, linkedObjId, -1, ItemNone)
			if err != nil {
				return err
			}
			if glog.V(2) {
				glog.Infof("Add linkedObj %d to masterObj %d", linkedObjs[0].getId(), objs[i].getId())
			}
			objs[i].linkedObjs = append(objs[i].linkedObjs, linkedObjs[0])
		}

	}
	return nil
}

// getHomeObjects : read objects
// If idObject > 0 return object with Id = idObject
// Else if idItem > 0 return all objects for Item definition idItem
// Else Return all objects for ItemType idItemType
func getHomeObjects(db *sql.DB, idObject int, idItem int, idItemType itemType) (objs []HomeObject, err error) {
	if db == nil {
		db, err = openDB()
		if err != nil {
			return
		}
		defer db.Close()
	}

	if idItemType <= ItemNone {
		var curObj HomeObject

		// read fields
		curObj.Fields, err = getItemFields(db, idItem, idObject)
		if err != nil {
			return
		}

		// read values
		values, err1 := getItemFieldValues(db, idItem, idObject)
		if err1 != nil {
			err = err1
			return
		}

		// build objs[] and assign values[] to each HomeObject instance
		if len(values) > 0 {
			startIdx := 0
			curIdObject := values[startIdx].IdObject
			for i, v := range values {
				if v.IdObject == curIdObject {
					continue
				}
				curObj.Values = values[startIdx:i]
				objs = append(objs, curObj)
				startIdx = i
				curIdObject = values[startIdx].IdObject
			}
			curObj.Values = values[startIdx:]
			objs = append(objs, curObj)
		}

		// Read linkedObjs

		if getLinkedObjects(db, objs) != nil {
			return
		}

	} else {

		// get all items for idItemType
		items, err1 := getManageItems(db, -1, idItemType)
		if err1 != nil {
			err = err1
			return
		}

		// for each item, read objects
		for _, item := range items {
			lstObjs, err1 := getHomeObjects(db, -1, item.Id, ItemNone)
			if err1 != nil {
				err = err1
				return
			}
			for _, obj := range lstObjs {
				objs = append(objs, obj)
			}
		}

	}

	return
}