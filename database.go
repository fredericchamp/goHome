// database.go
package main

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang/glog"
	_ "github.com/mattn/go-sqlite3"
)

// -----------------------------------------------

var dbFileName = ":memory:"

// -----------------------------------------------

type RefValue struct {
	Name  string
	Code  string
	Label string
}

type TItemType int

type TItemId int

const (
	ItemTypeNone TItemType = iota
	ItemEntity
	ItemSensor
	ItemActor
	ItemSensorAct
	ItemImageSensor
)

const ItemIdNone TItemId = 0

type Item struct {
	IdItem        TItemId
	Name          string
	IdProfil      TUserProfil
	IdItemType    TItemType
	IdMasterItem  TItemId
	IconeFileName string
}

type ItemField struct {
	IdField    int
	IdItem     TItemId
	NOrder     int
	Name       string
	IdDataType TDataType
	Label      string
	Helper     string
	Uniq       int
	RefList    string
	Regexp     string
}

type ItemFieldVal struct {
	IdObject int
	IdField  int
	Val      string
}

type HistoSensor struct {
	Ts       time.Time
	IdObject int
	Val      string
}

type HistoActor struct {
	Ts       time.Time
	IdObject int
	IdUser   int
	Param    string
	Res      string
}

// -----------------------------------------------

// execSqlStmtsFromFile : read text file and execute each line as a bd stmt
// ignore empty lines and lines starting with --
func execSqlStmtsFromFile(db *sql.DB, fileName string) (err error) {
	initFile, err := os.OpenFile(fileName, os.O_RDONLY, 0444)
	if err != nil {
		glog.Errorf("Fail to open %s for reading : %s", fileName, err)
		return
	}
	defer initFile.Close()

	lg := 0
	scanner := bufio.NewScanner(initFile)
	for scanner.Scan() {
		lg++
		sqlStmt := strings.TrimSpace(scanner.Text())

		if len(sqlStmt) == 0 || strings.HasPrefix(sqlStmt, "--") {
			continue
		}

		_, err = db.Exec(sqlStmt)
		if err != nil {
			glog.Errorf("File %s:%d : Error excuting sqlStmt (%s) : %s\n", fileName, lg, sqlStmt, err)
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
func openDB() (db *sql.DB, err error) {
	db, err = sql.Open("sqlite3", dbFileName)
	if err != nil {
		glog.Errorf("Failed to open sqlite3(%s) : ", dbFileName, err)
		return
	}
	return
}

// -----------------------------------------------

//  getRefList Read a reference list from DB and return it
func getRefList(db *sql.DB, listname string) (list []RefValue, err error) {
	if db == nil {
		if db, err = openDB(); err != nil {
			return
		}
		defer db.Close()
	}

	rows, err := db.Query("select name, code, label from RefValues where name like ? order by name, code", listname)
	if err != nil {
		glog.Errorf("getRefList query fail (name=%s) : %s ", listname, err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var curVal RefValue
		err = rows.Scan(&curVal.Name, &curVal.Code, &curVal.Label)
		if err != nil {
			glog.Errorf("getRefList scan fail (name=%s) : %s ", listname, err)
			return
		}
		list = append(list, curVal)
	}
	if err = rows.Err(); err != nil {
		glog.Errorf("getRefList rows.Err (name=%s) : %s ", listname, err)
		return
	}

	if len(list) <= 0 {
		err = errors.New(fmt.Sprintf("Ref values not found (name=%s)", listname))
		glog.Errorf("getRefList %s ", err)
		return
	}

	return
}

// getGlobalParam : fetch param value from db table goHome perimeter and name
// If multiple rows received only the first is read ... should necer append given unique index on table
func getGlobalParam(db *sql.DB, perimeter string, name string) (value string, err error) {
	if db == nil {
		if db, err = openDB(); err != nil {
			return
		}
		defer db.Close()
	}

	rows, err := db.Query("select val from goHome where perimeter = ? and name = ?", perimeter, name)
	if err != nil {
		glog.Errorf("getGlobalParam query fail (perimeter=%s,name=%s) : %s ", perimeter, name, err)
		return
	}
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&value)
	if err != nil {
		glog.Errorf("getGlobalParam scam fail  (perimeter=%s,name=%s) : %s ", perimeter, name, err)
		return
	}

	if err = rows.Err(); err != nil {
		glog.Errorf("getGlobalParam rows.Err  (perimeter=%s,name=%s) : %s ", perimeter, name, err)
		return
	}
	return
}

// getGlobalParam : fetch all param names/values from db table goHome for a given perimeter
func getGlobalParamList(db *sql.DB, perimeter string) (valMap map[string]string, err error) {
	if db == nil {
		if db, err = openDB(); err != nil {
			return
		}
		defer db.Close()
	}

	valMap = make(map[string]string)

	rows, err := db.Query("select name, val from goHome where perimeter = ?", perimeter)
	if err != nil {
		glog.Errorf("getGlobalParamList query fail (perimeter=%s) : %s ", perimeter, err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var val string
		err = rows.Scan(&name, &val)
		if err != nil {
			glog.Errorf("getGlobalParamList scan fail (perimeter=%s) : %s ", perimeter, err)
			return
		}
		valMap[name] = val
	}
	if err = rows.Err(); err != nil {
		glog.Errorf("getGlobalParamList rows.Err (perimeter=%s) : %s ", perimeter, err)
		return
	}
	return
}

// -----------------------------------------------
// -----------------------------------------------

// getManageItems select manage Items from DB
// If idItem > 0 return Item with given Id else return all Items
func getManageItems(db *sql.DB, idItemType TItemType, idItem TItemId) (items []Item, err error) {
	if db == nil {
		if db, err = openDB(); err != nil {
			return
		}
		defer db.Close()
	}

	var curItem Item
	var rows *sql.Rows

	switch {
	case idItem > 0:
		rows, err = db.Query("select idItem, Name, idProfil, idItemType, idMasterItem, iconeFileName from Item where idItem = ? ", idItem)
	case idItemType > 0:
		rows, err = db.Query("select idItem, Name, idProfil, idItemType, idMasterItem, iconeFileName from Item where idItemType = ? ", idItemType)
	default:
		rows, err = db.Query("select idItem, Name, idProfil, idItemType, idMasterItem, iconeFileName from Item order by idItem")
	}
	if err != nil {
		glog.Errorf("getManageItems query fail (type=%d,item=%d) : %s ", idItemType, idItem, err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&curItem.IdItem, &curItem.Name, &curItem.IdProfil, &curItem.IdItemType, &curItem.IdMasterItem, &curItem.IconeFileName)
		if err != nil {
			glog.Errorf("getManageItems scan fail (type=%d,item=%d) : %s ", idItemType, idItem, err)
			return
		}
		items = append(items, curItem)
	}
	if err = rows.Err(); err != nil {
		glog.Errorf("getManageItems rows.Err (type=%d,item=%d) : %s ", idItemType, idItem, err)
		return
	}

	if len(items) <= 0 {
		err = errors.New(fmt.Sprintf("Manage Items not found (itemType=%d,item=%d)", idItemType, idItem))
		glog.Errorf("getManageItems %s ", err)
		return
	}

	return
}

// -----------------------------------------------
// -----------------------------------------------

// getItemFields select item fields description.
// If idItem > 0 return item fields for the given item
// Else If idObject > 0 return item fields for the item of the given object
// Else return fields for all item
func getItemFields(db *sql.DB, idItem TItemId, idObject int) (fields []ItemField, err error) {
	if db == nil {
		if db, err = openDB(); err != nil {
			return
		}
		defer db.Close()
	}

	var curField ItemField
	var rows *sql.Rows

	switch {
	case idObject > 0:
		rows, err = db.Query("select f.idField, f.idItem, f.nOrder, f.Name, f.idDataType, f.Label, f.Helper, f.Uniq, f.RefList, f.Regexp from ItemField f, ItemFieldVal v where f.idField = v.idField and v.idObject = ? order by f.nOrder", idObject)
	case idItem > 0:
		rows, err = db.Query("select f.idField, f.idItem, f.nOrder, f.Name, f.idDataType, f.Label, f.Helper, f.Uniq, f.RefList, f.Regexp from ItemField f where f.idItem = ? order by f.nOrder", idItem)
	default:
		rows, err = db.Query("select f.idField, f.idItem, f.nOrder, f.Name, f.idDataType, f.Label, f.Helper, f.Uniq, f.RefList, f.Regexp from ItemField f order by f.idItem, f.nOrder")
	}
	if err != nil {
		glog.Errorf("getItemFields query fail (item=%d,obj=%d) : %s ", idItem, idObject, err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&curField.IdField, &curField.IdItem, &curField.NOrder, &curField.Name, &curField.IdDataType,
			&curField.Label, &curField.Helper, &curField.Uniq, &curField.RefList, &curField.Regexp)
		if err != nil {
			glog.Errorf("getItemFields scan fail (item=%d,obj=%d) : %s ", idItem, idObject, err)
			return
		}
		fields = append(fields, curField)
	}
	if err = rows.Err(); err != nil {
		glog.Errorf("getItemFields rows.Err (item=%d,obj=%d) : %s ", idItem, idObject, err)
		return
	}

	if len(fields) <= 0 {
		err = errors.New(fmt.Sprintf("Item fields not found (item=%d,object=%d)", idItem, idObject))
		glog.Errorf("getItemFields %s ", err)
		return
	}

	return
}

// -----------------------------------------------
// -----------------------------------------------

// getItemFieldValues select values for item fields.
// If idObject > 0 return values for the given object
// Else return all values for all objects with idField is part of idItem description
func getItemFieldValues(db *sql.DB, idItem TItemId, idObject int) (values []ItemFieldVal, err error) {
	if db == nil {
		if db, err = openDB(); err != nil {
			return
		}
		defer db.Close()
	}

	var curVal ItemFieldVal
	var rows *sql.Rows

	switch {
	case idObject > 0:
		rows, err = db.Query("select v.idObject, v.idField, v.Val from ItemFieldVal v, ItemField f where v.idField = f.idField and v.idObject = ? order by v.idObject, f.NOrder", idObject)
	case idItem > 0:
		rows, err = db.Query("select v.idObject, v.idField, v.Val from ItemFieldVal v, ItemField f where v.idField = f.idField and f.idItem = ? order by v.idObject, f.NOrder", idItem)
	default:
		err = errors.New(fmt.Sprintf("getItemFieldValues (idItem=%d idObject=%d) ... wont read all ItemFieldVal from db", idItem, idObject))
	}
	if err != nil {
		glog.Errorf("getItemFieldValues query fail (item=%d,obj=%d) : %s ", idItem, idObject, err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&curVal.IdObject, &curVal.IdField, &curVal.Val)
		if err != nil {
			glog.Errorf("getItemFieldValues scan fail (item=%d,obj=%d) : %s ", idItem, idObject, err)
			return
		}
		values = append(values, curVal)
	}
	if err = rows.Err(); err != nil {
		glog.Errorf("getItemFieldValues rows.Err fail (item=%d,obj=%d) : %s ", idItem, idObject, err)
		return
	}

	return
}

// getLinkedObjIds : return Ids of all objects which idMasterObj is set to idOject
func getLinkedObjIds(db *sql.DB, idObject int) (lstId []int, err error) {
	if db == nil {
		if db, err = openDB(); err != nil {
			return
		}
		defer db.Close()
	}

	rows, err := db.Query("select v.idObject from ItemFieldVal v, ItemField f where v.idField = f.idField and f.Name = 'idMasterObj' and v.Val = ? order by v.idObject", fmt.Sprint(idObject))
	if err != nil {
		glog.Errorf("getLinkedObjIds query fail (obj=%d) : %s ", idObject, err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var idLinkedObject int
		err = rows.Scan(&idLinkedObject)
		if err != nil {
			glog.Errorf("getLinkedObjIds scan fail (obj=%d) : %s ", idObject, err)
			return
		}
		lstId = append(lstId, idLinkedObject)
	}
	if err = rows.Err(); err != nil {
		glog.Errorf("getLinkedObjIds rows.Err (obj=%d) : %s ", idObject, err)
		return
	}

	return
}

// -----------------------------------------------
// -----------------------------------------------

// getHistoSensor : read values from HistoSensor
// if last the return the last available value (with greater timestamp)
// else return all values between [startTS and endTS] (if endTS <= 2016/01/01 returns all values with ts >= startTS)
func getHistoSensor(db *sql.DB, idObject int, last bool, startTS time.Time, endTS time.Time) (values []HistoSensor, err error) {
	if db == nil {
		if db, err = openDB(); err != nil {
			return
		}
		defer db.Close()
	}

	var rows *sql.Rows

	if last {
		rows, err = db.Query("select h.ts, h.idObject, h.Val from HistoSensor h where h.idObject = ? group by h.idObject having h.ts = max(h.ts)", idObject)
	} else {
		if endTS.Before(time.Date(2016, time.January, 1, 0, 0, 0, 0, time.Local)) {
			endTS = time.Now()
		}
		rows, err = db.Query("select h.ts, h.idObject, h.Val from HistoSensor h where h.idObject = ? and h.ts between ? and ? order by h.ts", idObject, startTS.Unix(), endTS.Unix())
	}
	if err != nil {
		glog.Errorf("getHistoSensor query fail (obj=%d,last=%d,start=%s,end=%s) : %s ", idObject, last, startTS, endTS, err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var curVal HistoSensor
		err = rows.Scan(&curVal.Ts, &curVal.IdObject, &curVal.Val)
		if err != nil {
			glog.Errorf("getHistoSensor scan fail (obj=%d,last=%d,start=%s,end=%s) : %s ", idObject, last, startTS, endTS, err)
			return
		}
		values = append(values, curVal)
	}
	if err = rows.Err(); err != nil {
		glog.Errorf("getHistoSensor rows.Err fail (obj=%d,last=%d,start=%s,end=%s) : %s ", idObject, last, startTS, endTS, err)
		return
	}

	return
}

// -----------------------------------------------
// -----------------------------------------------

// getHistActor : read values from HistoActor
// if last the return the last available value (with greater timestamp)
// else return all values between [startTS and endTS] (if endTS <= 2016/01/01 returns all values with ts >= startTS)
func getHistActor(db *sql.DB, idObject int, last bool, startTS time.Time, endTS time.Time) (values []HistoActor, err error) {
	if db == nil {
		if db, err = openDB(); err != nil {
			return
		}
		defer db.Close()
	}

	var rows *sql.Rows

	if last {
		rows, err = db.Query("select h.ts, h.idObject, h.idUser, h.Param, h.Res from HistoActor h where h.idObject = ? group by h.idObject having h.ts = max(h.ts)", idObject)
	} else {
		if endTS.Before(time.Date(2016, time.January, 1, 0, 0, 0, 0, time.Local)) {
			endTS = time.Now()
		}
		rows, err = db.Query("select h.ts, h.idObject, h.idUser, h.Param, h.Res from HistoActor h where h.idObject = ? and h.ts between ? and ? order by h.ts", idObject, startTS.Unix(), endTS.Unix())
	}
	if err != nil {
		glog.Errorf("getHistActor query fail (obj=%d,last=%d,start=%s,end=%s) : %s ", idObject, last, startTS, endTS, err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var curVal HistoActor
		err = rows.Scan(&curVal.Ts, &curVal.IdObject, &curVal.IdUser, &curVal.Param, &curVal.Res)
		if err != nil {
			glog.Errorf("getHistActor scan fail (obj=%d,last=%d,start=%s,end=%s) : %s ", idObject, last, startTS, endTS, err)
			return
		}
		values = append(values, curVal)
	}
	if err = rows.Err(); err != nil {
		glog.Errorf("getHistActor rows.Err (obj=%d,last=%d,start=%s,end=%s) : %s ", idObject, last, startTS, endTS, err)
		return
	}

	return
}

// writeItemFieldValues : write all values to DB
// All values in the [] must belong to the same object
// values[0].IdObject is used as objectid for all records
// if values[0].IdObject <= 0 the next available objectid is reserved and use as objectid for all records
func writeItemFieldValues(db *sql.DB, values []ItemFieldVal) (objectid int, err error) {
	if db == nil {
		if db, err = openDB(); err != nil {
			return
		}
		defer db.Close()
	}

	if len(values) <= 0 {
		err = errors.New("Nothing to write values[] is empty")
		return
	}

	tx, err := db.Begin()
	if err != nil {
		err = errors.New(fmt.Sprintf("writeItemFieldValues db.Begin : %s", err))
		return
	}
	objectid = values[0].IdObject

	if objectid <= 0 {
		// New Id needed
		row := tx.QueryRow("select max(idObject)+1 from ItemFieldVal")
		if row == nil {
			glog.Error("writeItemFieldValues fail to select next objectid")
			tx.Rollback()
			return
		}
		if err = row.Scan(&objectid); err != nil {
			glog.Errorf("writeItemFieldValues fail to scan next objectid : %s", err)
			tx.Rollback()
			return
		}
	} else {
		// Delete existing ItemFieldVal
		_, err = db.Exec(fmt.Sprintf("delete from ItemFieldVal where idObject = %d", objectid))
		if err != nil {
			glog.Errorf("writeItemFieldValues fail to delete ItemFieldVal for objectid=%d : %s", objectid, err)
			tx.Rollback()
			return
		}
	}

	stmt, err := tx.Prepare("insert into ItemFieldVal (idObject, idField, Val) values(?, ?, ?)")
	if err != nil {
		glog.Errorf("writeItemFieldValues fail to prepare insert into ItemFieldVal : %s", err)
		tx.Rollback()
		return
	}
	defer stmt.Close()

	for _, val := range values {
		_, err = stmt.Exec(objectid, val.IdField, val.Val)
		if err != nil {
			glog.Errorf("writeItemFieldValues fail to exec insert into ItemFieldVal (objId=%d, %v) : %s", objectid, val, err)
			tx.Rollback()
			return
		}
	}
	tx.Commit()

	return
}
