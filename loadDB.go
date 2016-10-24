// loadDB.go
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/golang/glog"
)

// getGlobalParam : fetch param value from db table goHome
// if id >=0 query db using this id (ignoring scope and name parameter) else query using scope and name
// If multiple rows received only the first is read
func getGlobalParam(db *sql.DB, id int, scope string, name string) (value string, err error) {
	if db == nil {
		db, err = openDB()
		if err != nil {
			glog.Error(err)
			return
		}
		defer db.Close()
	}

	var rows *sql.Rows

	if id >= 0 {
		rows, err = db.Query("select value from goHome where id = ?", id)
	} else {
		rows, err = db.Query("select value from goHome where scope = ? and name = ?", scope, name)
	}
	if err != nil {
		glog.Error(err)
		return
	}
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&value)
	if err != nil {
		glog.Error(err)
		return
	}

	if err = rows.Err(); err != nil {
		glog.Error(err)
		return
	}
	return
}

// -----------------------------------------------
// -----------------------------------------------

type Item struct {
	Id           int
	Name         string
	IdProfil     int
	IdItemType   int
	ItemTable    string
	IdMasterItem int
	Icone        []byte
}

// getManageItems select manage Items from DB
// If idItem >=0 return Item with given Id else return all Items
func getManageItems(db *sql.DB, idItem int, idItemType int) (items []Item, err error) {
	if db == nil {
		db, err = openDB()
		if err != nil {
			glog.Error(err)
			return
		}
		defer db.Close()
	}

	var curItem Item
	var rows *sql.Rows

	switch {
	case idItem >= 0:
		rows, err = db.Query("select id, Name, idProfil, idItemType, idMasterItem, icone from Item where id = ? ", idItem)
	case idItemType >= 0:
		rows, err = db.Query("select id, Name, idProfil, idItemType, idMasterItem, icone from Item where idItemType = ? ", idItemType)
	default:
		rows, err = db.Query("select id, Name, idProfil, idItemType, idMasterItem, icone from Item order by id")
	}
	if err != nil {
		glog.Error(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&curItem.Id, &curItem.Name, &curItem.IdProfil, &curItem.IdItemType, &curItem.IdMasterItem, &curItem.Icone)
		if err != nil {
			glog.Error(err)
			return
		}
		items = append(items, curItem)
	}
	if err = rows.Err(); err != nil {
		glog.Error(err)
		return
	}

	return
}

// -----------------------------------------------
// -----------------------------------------------

type ItemField struct {
	IdField    int
	IdItem     int
	NOrder     int
	Name       string
	IdDataType int
	Helper     string
	Regexp     string
}

// getItemFields select item fields description.
// If idItem >=0 return item fields for the given item
// Else If idObject >=0 return item fields for the item of the given object
// Else return fields for all item
func getItemFields(db *sql.DB, idItem int, idObject int) (fields []ItemField, err error) {
	if db == nil {
		db, err = openDB()
		if err != nil {
			glog.Error(err)
			return
		}
		defer db.Close()
	}

	var curField ItemField
	var rows *sql.Rows

	switch {
	case idItem >= 0:
		rows, err = db.Query("select idField, idItem, nOrder, Name, idDataType, Helper, Regexp from ItemField where idItem = ? order by nOrder", idItem)
	case idObject >= 0:
		rows, err = db.Query("select f.idField, f.idItem, f.nOrder, f.Name, f.idDataType, f.Helper, f.Regexp from ItemField f, ItemFieldVal v where f.idField = v.idField and v.idObject = ? order by nOrder", idObject)
	default:
		rows, err = db.Query("select idField, idItem, nOrder, Name, idDataType, Helper, Regexp from ItemField order by idItem, nOrder")
	}
	if err != nil {
		glog.Error(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&curField.IdField, &curField.IdItem, &curField.NOrder, &curField.Name, &curField.IdDataType, &curField.Helper, &curField.Regexp)
		if err != nil {
			glog.Error(err)
			return
		}
		fields = append(fields, curField)
	}
	if err = rows.Err(); err != nil {
		glog.Error(err)
		return
	}
	return
}

type ItemFieldVal struct {
	IdObject int
	IdField  int
	IntVal   int
	FloatVal float32
	TextVal  string
	ByteVal  []byte
}

type HomeObject struct {
	Fields []ItemField
	Values []ItemFieldVal
}

func (obj HomeObject) getIntVal(fieldName string) (value int, err error) {
	var idx int = -1
	for i, v := range obj.Fields {
		if v.Name == fieldName {
			idx = i
			break
		}
	}
	if idx < 0 {
		err = errors.New(fmt.Sprintf("Field '%s' not found", fieldName))
		glog.Error(err)
		if glog.V(2) {
			glog.Info(obj)
		}
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
		if glog.V(2) {
			glog.Info(obj)
		}
	}
	return
}

func (obj HomeObject) getStrVal(fieldName string) (value string, err error) {
	var idx int = -1
	for i, v := range obj.Fields {
		if v.Name == fieldName {
			idx = i
			break
		}
	}
	if idx < 0 {
		err = errors.New(fmt.Sprintf("Field '%s' not found", fieldName))
		glog.Error(err)
		if glog.V(2) {
			glog.Info(obj)
		}
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
		if glog.V(2) {
			glog.Info(obj)
		}
	}
	return
}

func (obj HomeObject) getByteVal(fieldName string) (value []byte, err error) {
	var idx int = -1
	for i, v := range obj.Fields {
		if v.Name == fieldName {
			idx = i
			break
		}
	}
	if idx < 0 {
		err = errors.New(fmt.Sprintf("Field '%s' not found", fieldName))
		glog.Error(err)
		if glog.V(2) {
			glog.Info(obj)
		}
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
		if glog.V(2) {
			glog.Info(obj)
		}
	}
	return
}

// getDBObject select object from db
// If idObject >=0 return Object with Id = idObject else return all object for Item definition idItem
func getDBObjects(db *sql.DB, idObject int, idItem int) (objs []HomeObject, err error) {
	if db == nil {
		db, err = openDB()
		if err != nil {
			glog.Error(err)
			return
		}
		defer db.Close()
	}

	var curObj HomeObject
	var curVal ItemFieldVal
	var curIdObject int
	var rows *sql.Rows

	curObj.Fields, err = getItemFields(db, idItem, idObject)
	if err != nil {
		return
	}

	if idObject >= 0 {
		rows, err = db.Query("select v.idObject, v.idField, v.intVal, v.floatVal, v.textVal, v.byteVal from ItemFieldVal v, ItemField f where v.idField = f.idField and v.idObject = ? order by v.idObject, f.NOrder", idObject)
	} else {
		rows, err = db.Query("select v.idObject, v.idField, v.intVal, v.floatVal, v.textVal, v.byteVal from ItemFieldVal v, ItemField f where v.idField = f.idField and f.idItem = ? order by v.idObject, f.NOrder", idItem)
	}
	if err != nil {
		glog.Error(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&curVal.IdObject, &curVal.IdField, &curVal.IntVal, &curVal.FloatVal, &curVal.TextVal, &curVal.ByteVal)
		if curIdObject == 0 {
			curIdObject = curVal.IdObject
		}
		if err != nil {
			glog.Error(err)
			return
		}
		if curIdObject != curVal.IdObject {
			objs = append(objs, curObj)
			curObj.Values = make([]ItemFieldVal, 0, len(curObj.Fields))
			curIdObject = curVal.IdObject
		}
		curObj.Values = append(curObj.Values, curVal)
	}
	objs = append(objs, curObj)

	if err = rows.Err(); err != nil {
		glog.Error(err)
		return
	}

	return
}
