// homeobj.go
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/golang/glog"
)

// -----------------------------------------------
type TDataType int

const (
	DBTypeNone TDataType = iota
	DBTypeBool
	DBTypeInt
	DBTypeFloat
	DBTypeText
	DBTypeDateTime
	DBTypeURL
)

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
	if len(obj.Values) < idx {
		err = errors.New(fmt.Sprintf("No value for '%s' field", fieldName))
	} else {
		switch obj.Fields[idx].IdDataType {
		case DBTypeBool, DBTypeInt, DBTypeDateTime, DBTypeText:
			value, err = strconv.Atoi(obj.Values[idx].Val)
		case DBTypeFloat:
			err = errors.New(fmt.Sprintf("Not converting float to int for '%s' field", fieldName))
		case DBTypeURL:
			err = errors.New(fmt.Sprintf("Not converting filename to int for '%s' field", fieldName))
		default:
			err = errors.New(fmt.Sprintf("Unknown data type %d for '%s' field", obj.Fields[idx].IdDataType, fieldName))
		}
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
	if len(obj.Values) < idx {
		err = errors.New(fmt.Sprintf("No value for '%s' field", fieldName))
	} else {
		switch obj.Fields[idx].IdDataType {
		case DBTypeBool:
			if obj.Values[idx].Val == "0" {
				value = "No"
			} else {
				value = "Yes"
			}
		case DBTypeInt, DBTypeDateTime, DBTypeFloat:
			value = fmt.Sprint(obj.Values[idx].Val)
		case DBTypeText, DBTypeURL:
			value = obj.Values[idx].Val
		default:
			err = errors.New(fmt.Sprintf("Unknown data type %d for '%s' field", obj.Fields[idx].IdDataType, fieldName))
		}
	}
	if err != nil {
		glog.Error(err)
		if glog.V(1) {
			glog.Info(obj)
		}
	}
	return
}

// ValidateValues : check values are valid regarding obj.Fields
func (obj HomeObject) ValidateValues(values []ItemFieldVal) (err error) { // TODO
	return
}

// -----------------------------------------------

// getLinkedObjects add to each objs[] corresponding linked objects
// todo : reorg/optimisation : currently need 1 db query for each objs[] + 1 db query for each linked obj
func getLinkedObjects(db *sql.DB, objs []HomeObject) (err error) {
	if db == nil {
		if db, err = openDB(); err != nil {
			return
		}
		defer db.Close()
	}

	for i := range objs {
		lstLinkedObjId, err := getLinkedObjIds(db, objs[i].getId())
		if err != nil {
			return err
		}
		for _, linkedObjId := range lstLinkedObjId {
			linkedObjs, err := getHomeObjects(db, ItemTypeNone, ItemIdNone, linkedObjId)
			if err != nil {
				return err
			}
			if glog.V(2) {
				glog.Infof("Add linkedObj %d to masterObj %d", linkedObjs[0].getId(), objs[i].getId())
			}
			objs[i].linkedObjs = append(objs[i].linkedObjs, linkedObjs[0])
		}

	}
	return
}

// getHomeObjects : read objects
// If idObject > 0 return object with Id = idObject
// Else if idItem > 0 return all objects for Item definition idItem
// Else Return all objects for ItemType idItemType
func getHomeObjects(db *sql.DB, idItemType TItemType, idItem TItemId, idObject int) (objs []HomeObject, err error) {
	if db == nil {
		if db, err = openDB(); err != nil {
			return
		}
		defer db.Close()
	}

	if idItem > 0 || idObject > 0 {
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
		items, err1 := getManageItems(db, idItemType, ItemIdNone)
		if err1 != nil {
			err = err1
			return
		}

		// for each item, read objects
		for _, item := range items {
			lstObjs, err1 := getHomeObjects(db, ItemTypeNone, item.IdItem, -1)
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

// writeObject : save object to db
func writeObject(obj HomeObject) (objectid int, err error) {
	objectid, err = writeItemFieldValues(nil, obj.Values)
	return
}
