// ser.go
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

var userListLock sync.Mutex
var userList []HomeObject

// -----------------------------------------------

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
	userList, err = getHomeObjects(db, -1, userItemId, ItemNone)

	nbUser = len(userList)

	return
}

// getUserFromCert : return user HomeObject or error if not found
func getUserFromCert(peerCrt []*x509.Certificate) (userObj HomeObject, err error) {

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

	for _, obj := range userList {
		if glog.V(3) {
			glog.Info("User : ", obj)
		}
		val, err1 := obj.getStrVal("Email")
		if err1 != nil || strings.ToUpper(val) != strings.ToUpper(email) {
			continue
		}
		iProfil, err1 := obj.getIntVal("IdProfil")
		if err1 != nil {
			continue // ignore if non IdProfile field
		}
		iActive, err1 := obj.getIntVal("IsActive")
		if err1 != nil {
			continue // ignore if non IsActive field
		}
		if glog.V(2) {
			glog.Infof("Found active(%d) user for '%s' : id=%d profil=%d)", iActive, email, userObj.getId(), iProfil)
		}
		userObj = obj
		return
	}

	err = errors.New(fmt.Sprintf("No user found for '%s'", email))
	if glog.V(1) {
		glog.Error("Error getUserFromCert : ", err)
	}

	return
}

func checkApiUser(userObj HomeObject) (profil userProfil, err error) {
	i, err := userObj.getIntVal("IdProfil")
	if err != nil {
		return
	}
	profil = userProfil(i)
	if profil <= ProfilNone {
		err = errors.New(`{"error":"insufficient privileges"}`)
		return
	}

	iActive, err := userObj.getIntVal("IsActive")
	if err != nil {
		return
	}
	if iActive <= 0 {
		err = errors.New(`{"error":"Not an active user"}`)
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
		// HomeObject without IdProfil must not be filtered out
		if obj.hasField("IdProfil") {
			objProfil, err := obj.getIntVal("IdProfil")
			if err != nil {
				continue
			}
			if objProfil < int(profil) {
				continue
			}
		}
		filteredObjs = append(filteredObjs, obj)
	}
	return
}
