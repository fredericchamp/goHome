// database.go
package main

import (
	"crypto/x509"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/golang/glog"
	_ "github.com/mattn/go-sqlite3"
)

// -----------------------------------------------

type userProfil int

const (
	ProfilNone = iota
	ProfilAdmin
	ProfilUser
)

var userProfilNames = map[userProfil]string{
	ProfilNone:  "No privilege",
	ProfilAdmin: "Administrator",
	ProfilUser:  "User",
}

type itemType int

const (
	ItemNone itemType = iota
	ItemEntity
	ItemSensor
	ItemActor
	ItemSensorAct
	// TODO :
	ItemVideoSensor
)

var itemTypeNames = map[itemType]string{
	ItemNone:      "None",
	ItemEntity:    "Entity",
	ItemSensor:    "Sensor",
	ItemActor:     "Actor",
	ItemSensorAct: "Actor trigger by sensor",
	// TODO :
	ItemVideoSensor: "VideoSensor",
}

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

const (
	DurationMS = "ms"
	DurationS  = "s"
	DurationM  = "m"
	DurationH  = "h"
	DurationD  = "d"
)

// -----------------------------------------------

var userListLock sync.Mutex
var userList []HomeObject

// -----------------------------------------------

type HomeObject struct {
	Fields     []ItemField
	Values     []ItemFieldVal
	linkedObjs []HomeObject
}

// -----------------------------------------------

func (obj HomeObject) getId() int {
	if len(obj.Values) <= 0 {
		glog.Error("Error trying to get Id from an HomeObject with empty []Values")
		return -1
	}
	return obj.Values[0].IdObject
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
		if glog.V(1) {
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
		if glog.V(1) {
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
		if glog.V(1) {
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
		if glog.V(1) {
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
		if glog.V(1) {
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
		if glog.V(1) {
			glog.Info(obj)
		}
	}
	return
}

// getDBObject select object from db
// If idObject > 0 return Object with Id = idObject else return all object for Item definition idItem
// TODO reorg to move DB part to database.go
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

	if idObject > 0 {
		rows, err = db.Query("select v.idObject, v.idField, v.intVal, v.floatVal, v.textVal, v.byteVal from ItemFieldVal v, ItemField f where v.idField = f.id and v.idObject = ? order by v.idObject, f.NOrder", idObject)
	} else {
		rows, err = db.Query("select v.idObject, v.idField, v.intVal, v.floatVal, v.textVal, v.byteVal from ItemFieldVal v, ItemField f where v.idField = f.id and f.idItem = ? order by v.idObject, f.NOrder", idItem)
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

// getDBObject select object from db
// Return all object ItemType idItemType
func getDBObjectsForType(db *sql.DB, idItemType itemType) (objs []HomeObject, err error) {
	if db == nil {
		db, err = openDB()
		if err != nil {
			glog.Error(err)
			return
		}
		defer db.Close()
	}

	items, err := getManageItems(db, -1, idItemType)
	if err != nil {
		return
	}

	var lst []HomeObject
	for _, item := range items {
		lst, err = getDBObjects(db, -1, item.Id)
		if err != nil {
			return
		}
		for _, obj := range lst {
			objs = append(objs, obj)
		}
	}
	return
}

// getEmailFromCert : read received peer X509 certificates and return found email or err
func getEmailFromCert(peerCrt []*x509.Certificate) (email string, err error) {
	if len(peerCrt) <= 0 {
		err = errors.New("No certificat received")
		glog.Error(err)
		return
	}

	// With the certificats I use email is the 5th Attribut -  this may chage given the CA settings (TODO : check, not sure)
	if len(peerCrt[0].Subject.Names) < 5 {
		err = errors.New(fmt.Sprintf("Did not locate email in (%s)", peerCrt[0].Subject.Names))
		glog.Error(err)
		return
	}
	email = peerCrt[0].Subject.Names[4].Value.(string)

	return
}

// loadUsers : load all users from DB into global userList
// If force == false then only load if userList is empty
func loadUsers(force bool) (nbUser int, err error) {

	userListLock.Lock()
	defer userListLock.Unlock()

	nbUser = len(userList)
	if nbUser > 0 && !force {
		return
	}

	value, err := getGlobalParam(nil, -1, "goHome", "UserItemId")
	if err != nil {
		glog.Errorf("Error reading UserItemId param : %s", err)
		return
	}
	userItemId, err := strconv.Atoi(value)
	if err != nil {
		glog.Errorf("Error converting userItemId (%s) : %s", value, err)
		return
	}

	// read all users
	userList, err = getDBObjects(nil, -1, userItemId)

	nbUser = len(userList)

	return
}

// getUserEmailAndProfil : return user email and profil
func getUserEmailAndProfil(peerCrt []*x509.Certificate) (email string, profil userProfil) {

	profil = ProfilNone

	email, err := getEmailFromCert(peerCrt)
	if err != nil {
		return
	}
	if glog.V(2) {
		glog.Infof("Client email from cert = '%s'", email)
	}

	_, err = loadUsers(false)
	if err != nil {
		return
	}

	// Search a user with the email found in the certificat
	userListLock.Lock()
	defer userListLock.Unlock()

	for _, userObj := range userList {
		if glog.V(3) {
			glog.Info("User : ", userObj)
		}
		i, err := userObj.getIntVal("IsActive")
		if err != nil || i <= 0 {
			continue
		}
		val, err := userObj.getStrVal("Email")
		if err != nil || strings.ToUpper(val) != strings.ToUpper(email) {
			continue
		}
		i, err = userObj.getIntVal("IdProfil")
		if err != nil {
			return
		}
		profil = userProfil(i)
		if glog.V(2) {
			glog.Infof("Found active user for '%s' : %d (profil=%d)", email, userObj.getId(), profil)
		}
		return
	}

	if glog.V(2) {
		glog.Infof("No active user found for '%s'", email)
	}

	return
}
