// database.go
package main

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/golang/glog"
	_ "github.com/mattn/go-sqlite3"
)

var dbFileName = ":memory:"

func execSqlStmtsFromFile(db *sql.DB, fileName string) (err error) {
	initFile, err := os.OpenFile(fileName, os.O_RDONLY, 0444)
	if err != nil {
		glog.Errorf("Fail to open %s for reading : %s", fileName, err)
		return
	}
	defer initFile.Close()

	scanner := bufio.NewScanner(initFile)
	for scanner.Scan() {
		sqlStmt := strings.TrimSpace(scanner.Text())

		if len(sqlStmt) == 0 || strings.HasPrefix(sqlStmt, "--") {
			continue
		}

		_, err = db.Exec(sqlStmt)
		if err != nil {
			glog.Errorf("Error excuting sqlStmt (%s) : %s\n", sqlStmt, err)
			return
		}
	}
	if err = scanner.Err(); err != nil {
		glog.Errorf("Scanner error reading file '%s' : %s", fileName, err)
		return
	}
	return
}

// initDBFile : Check dbfile existance and acces right
// Create and initialize and new file if needed
func initDBFile(dbfile string) error {
	createEmptyDB := false
	dbFileName = dbfile

	// Check DB file if not ":memory:"
	if ":memory:" != dbFileName {
		file, err := os.OpenFile(dbFileName, os.O_RDWR, 0666)
		if err != nil {
			switch {
			case os.IsNotExist(err):
				createEmptyDB = true
				glog.Info("Creating new DB file")
			case os.IsPermission(err):
				glog.Errorf("Permission Error accessing '%s' : %s", dbFileName, err)
				return err
			default:
				glog.Errorf("Unknow Error accessing '%s' : %s", dbFileName, err)
				return err
			}
		}
		file.Close()
	}

	// Init DB if new
	if createEmptyDB {
		// Open DB
		db, err := openDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// initDB sql stmt must be in init.sql file in the same dir as dbFileName
		err = execSqlStmtsFromFile(db, fmt.Sprintf("%s%c%s", filepath.Dir(dbFileName), filepath.Separator, "init.sql"))
		if err != nil {
			return err
		}
		// perso.sql may not be present => ignore error
		execSqlStmtsFromFile(db, fmt.Sprintf("%s%c%s", filepath.Dir(dbFileName), filepath.Separator, "perso.sql"))
	}
	return nil
}

// openDB open a database connection and return it
func openDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbFileName)
	if err != nil {
		glog.Error(err)
	}
	return db, err
}

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
func getManageItems(db *sql.DB, idItem int, idItemType itemType) (items []Item, err error) {
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
	Id         int
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
		rows, err = db.Query("select id, idItem, nOrder, Name, idDataType, Helper, Regexp from ItemField where idItem = ? order by nOrder", idItem)
	case idObject >= 0:
		rows, err = db.Query("select f.id, f.idItem, f.nOrder, f.Name, f.idDataType, f.Helper, f.Regexp from ItemField f, ItemFieldVal v where f.idField = v.idField and v.idObject = ? order by nOrder", idObject)
	default:
		rows, err = db.Query("select id, idItem, nOrder, Name, idDataType, Helper, Regexp from ItemField order by idItem, nOrder")
	}
	if err != nil {
		glog.Error(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&curField.Id, &curField.IdItem, &curField.NOrder, &curField.Name, &curField.IdDataType, &curField.Helper, &curField.Regexp)
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
	Id       int
	IdField  int
	IntVal   int
	FloatVal float32
	TextVal  string
	ByteVal  []byte
}

type HomeObject struct {
	Fields     []ItemField
	Values     []ItemFieldVal
	linkedObjs []HomeObject
}

func (obj HomeObject) getId() int {
	return obj.Values[1].Id
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
		rows, err = db.Query("select v.id, v.idField, v.intVal, v.floatVal, v.textVal, v.byteVal from ItemFieldVal v, ItemField f where v.idField = f.id and v.idObject = ? order by v.id, f.NOrder", idObject)
	} else {
		rows, err = db.Query("select v.id, v.idField, v.intVal, v.floatVal, v.textVal, v.byteVal from ItemFieldVal v, ItemField f where v.idField = f.id and f.idItem = ? order by v.id, f.NOrder", idItem)
	}
	if err != nil {
		glog.Error(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&curVal.Id, &curVal.IdField, &curVal.IntVal, &curVal.FloatVal, &curVal.TextVal, &curVal.ByteVal)
		if curIdObject == 0 {
			curIdObject = curVal.Id
		}
		if err != nil {
			glog.Error(err)
			return
		}
		if curIdObject != curVal.Id {
			objs = append(objs, curObj)
			curObj.Values = make([]ItemFieldVal, 0, len(curObj.Fields))
			curIdObject = curVal.Id
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
// If idObject >=0 return Object with Id = idObject else return all object for Item definition idItem
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
