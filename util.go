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

// getHomeObjects read objects
// If idObject > 0 return Object with Id = idObject
// Else return all object for Item definition idItem
// TODO : add linkedObjs loading ?
func getHomeObjects(db *sql.DB, idObject int, idItem int) (objs []HomeObject, err error) {
	if db == nil {
		db, err = openDB()
		if err != nil {
			return
		}
		defer db.Close()
	}

	var curObj HomeObject

	// read fields
	curObj.Fields, err = getItemFields(db, idItem, idObject)
	if err != nil {
		return
	}

	// read values
	values, err := getItemFieldValues(db, idItem, idObject)
	if err != nil {
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
	}
	objs = append(objs, curObj)

	return
}

// getHomeObjectsForType read objects
// Return all object ItemType idItemType
func getHomeObjectsForType(db *sql.DB, idItemType itemType) (objs []HomeObject, err error) {
	if db == nil {
		db, err = openDB()
		if err != nil {
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
		lst, err = getHomeObjects(db, -1, item.Id)
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
func loadUsers(db *sql.DB, force bool) (nbUser int, err error) {
	if db == nil {
		db, err = openDB()
		if err != nil {
			return
		}
		defer db.Close()
	}

	userListLock.Lock()
	defer userListLock.Unlock()

	nbUser = len(userList)
	if nbUser > 0 && !force {
		return
	}

	value, err := getGlobalParam(db, -1, "goHome", "UserItemId")
	if err != nil {
		return
	}
	userItemId, err := strconv.Atoi(value)
	if err != nil {
		glog.Errorf("Error converting userItemId (%s) : %s", value, err)
		return
	}

	// read all users
	userList, err = getHomeObjects(db, -1, userItemId)

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

	_, err = loadUsers(nil, false)
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

// profilFilteredItems : return an []Item with only Item matching user profil
func profilFilteredItems(profil userProfil, items []Item) (filteredItems []Item) {
	for _, item := range items {
		if item.IdProfil < int(profil) {
			continue
		}
		filteredItems = append(filteredItems, item)
	}
	return
}

// profilFilteredObjects : return an []HomeObject with only Item matching user profil
func profilFilteredObjects(profil userProfil, objs []HomeObject) (filteredObjs []HomeObject) {
	for _, obj := range objs {
		objProfil, err := obj.getIntVal("IdProfil")
		if err != nil {
			continue
		}
		if objProfil < int(profil) {
			continue
		}
		filteredObjs = append(filteredObjs, obj)
	}
	return
}
