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

// execSqlStmtsFromFile : read text file and execute each line as a bd stmt
// ignore empty lines and lines starting with --
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
			glog.Errorf("File %s : Error excuting sqlStmt (%s) : %s\n", fileName, sqlStmt, err)
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

// getGlobalParam : fetch param value from db table goHome
// if id > 0 query db using this id (ignoring perimeter and name parameter) else query using perimeter and name
// If multiple rows received only the first is read
func getGlobalParam(db *sql.DB, idParam int, perimeter string, name string) (value string, err error) {
	if db == nil {
		if db, err = openDB(); err != nil {
			return
		}
		defer db.Close()
	}

	var rows *sql.Rows

	if idParam > 0 {
		rows, err = db.Query("select value from goHome where idParam = ?", idParam)
	} else {
		rows, err = db.Query("select value from goHome where perimeter = ? and name = ?", perimeter, name)
	}
	if err != nil {
		glog.Errorf("Failed to read global param (%d,%s,%s) : %s ", idParam, perimeter, name, err)
		return
	}
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&value)
	if err != nil {
		glog.Errorf("Failed to read global param (%d,%s,%s) : %s ", idParam, perimeter, name, err)
		return
	}

	if err = rows.Err(); err != nil {
		glog.Errorf("Failed to read global param (%d,%s,%s) : %s ", idParam, perimeter, name, err)
		return
	}
	return
}

// -----------------------------------------------
// -----------------------------------------------

type Item struct {
	IdItem       int
	Name         string
	IdProfil     int
	IdItemType   int
	ItemTable    string
	IdMasterItem int
	Icone        []byte
}

// getManageItems select manage Items from DB
// If idItem > 0 return Item with given Id else return all Items
func getManageItems(db *sql.DB, idItemType itemType, idItem int) (items []Item, err error) {
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
		rows, err = db.Query("select idItem, Name, idProfil, idItemType, idMasterItem, icone from Item where idItem = ? ", idItem)
	case idItemType > 0:
		rows, err = db.Query("select idItem, Name, idProfil, idItemType, idMasterItem, icone from Item where idItemType = ? ", idItemType)
	default:
		rows, err = db.Query("select idItem, Name, idProfil, idItemType, idMasterItem, icone from Item order by idItem")
	}
	if err != nil {
		glog.Error(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&curItem.IdItem, &curItem.Name, &curItem.IdProfil, &curItem.IdItemType, &curItem.IdMasterItem, &curItem.Icone)
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

	if len(items) <= 0 {
		err = errors.New(fmt.Sprintf("Item not found (itemType=%d,item=%d)", idItemType, idItem))
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
	Rules      string
}

// getItemFields select item fields description.
// If idItem > 0 return item fields for the given item
// Else If idObject > 0 return item fields for the item of the given object
// Else return fields for all item
func getItemFields(db *sql.DB, idItem int, idObject int) (fields []ItemField, err error) {
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
		rows, err = db.Query("select f.idField, f.idItem, f.nOrder, f.Name, f.idDataType, f.Helper, f.Rules from ItemField f, ItemFieldVal v where f.idField = v.idField and v.idObject = ? order by nOrder", idObject)
	case idItem > 0:
		rows, err = db.Query("select idField, idItem, nOrder, Name, idDataType, Helper, Rules from ItemField where idItem = ? order by nOrder", idItem)
	default:
		rows, err = db.Query("select idField, idItem, nOrder, Name, idDataType, Helper, Rules from ItemField order by idItem, nOrder")
	}
	if err != nil {
		glog.Error(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&curField.IdField, &curField.IdItem, &curField.NOrder, &curField.Name, &curField.IdDataType, &curField.Helper, &curField.Rules)
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

	if len(fields) <= 0 {
		err = errors.New(fmt.Sprintf("Item fields not found (item=%d,object=%d)", idItem, idObject))
		return
	}

	return
}

// -----------------------------------------------
// -----------------------------------------------

type ItemFieldVal struct {
	IdObject int
	IdField  int
	IntVal   int
	FloatVal float32
	TextVal  string
	ByteVal  []byte
}

// getItemFieldValues select values for item fields.
// If idObject > 0 return values for the given object
// Else return all values for all objects with idField is part of idItem description
func getItemFieldValues(db *sql.DB, idItem int, idObject int) (values []ItemFieldVal, err error) {
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
		rows, err = db.Query("select v.idObject, v.idField, v.intVal, v.floatVal, v.textVal, v.byteVal from ItemFieldVal v, ItemField f where v.idField = f.idField and v.idObject = ? order by v.idObject, f.NOrder", idObject)
	case idItem > 0:
		rows, err = db.Query("select v.idObject, v.idField, v.intVal, v.floatVal, v.textVal, v.byteVal from ItemFieldVal v, ItemField f where v.idField = f.idField and f.idItem = ? order by v.idObject, f.NOrder", idItem)
	default:
		err = errors.New(fmt.Sprintf("getItemFieldValues (idItem=%d idObject=%d) ... wont read all ItemFieldVal from db", idItem, idObject))
	}
	if err != nil {
		glog.Error(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&curVal.IdObject, &curVal.IdField, &curVal.IntVal, &curVal.FloatVal, &curVal.TextVal, &curVal.ByteVal)
		if err != nil {
			glog.Error(err)
			return
		}
		values = append(values, curVal)
	}
	if err = rows.Err(); err != nil {
		glog.Error(err)
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

	rows, err := db.Query("select v.idObject from ItemFieldVal v, ItemField f where v.idField = f.idField and f.Name = 'idMasterObj' and v.intVal = ? order by v.idObject", idObject)
	if err != nil {
		glog.Error(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var idLinkedObject int
		err = rows.Scan(&idLinkedObject)
		if err != nil {
			glog.Error(err)
			return
		}
		lstId = append(lstId, idLinkedObject)
	}
	if err = rows.Err(); err != nil {
		glog.Error(err)
		return
	}

	return
}

// -----------------------------------------------
// -----------------------------------------------

type HistoSensor struct {
	Ts       time.Time
	IdObject int
	IntVal   int
	FloatVal float32
	TextVal  string
}

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
		rows, err = db.Query("select h.ts, h.idObject, h.intVal, h.floatVal, h.textVal from HistoSensor h where h.idObject = ? group by h.idObject having h.ts = max(h.ts)", idObject)
	} else {
		if endTS.Before(time.Date(2016, time.January, 1, 0, 0, 0, 0, time.Local)) {
			endTS = time.Now()
		}
		rows, err = db.Query("select h.ts, h.idObject, h.intVal, h.floatVal, h.textVal from HistoSensor h where h.idObject = ? and h.ts between ? and ? order by h.ts", idObject, startTS.Unix(), endTS.Unix())
	}
	if err != nil {
		glog.Error(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var curVal HistoSensor
		err = rows.Scan(&curVal.Ts, &curVal.IdObject, &curVal.IntVal, &curVal.FloatVal, &curVal.TextVal)
		if err != nil {
			glog.Error(err)
			return
		}
		values = append(values, curVal)
	}
	if err = rows.Err(); err != nil {
		glog.Error(err)
		return
	}

	return
}

// -----------------------------------------------
// -----------------------------------------------

type HistoActor struct {
	Ts       time.Time
	IdObject int
	IdUser   int
	Param    string
	Result   string
}

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
		rows, err = db.Query("select h.ts, h.idObject, h.idUser, h.Param, h.Result from HistoActor h where h.idObject = ? group by h.idObject having h.ts = max(h.ts)", idObject)
	} else {
		if endTS.Before(time.Date(2016, time.January, 1, 0, 0, 0, 0, time.Local)) {
			endTS = time.Now()
		}
		rows, err = db.Query("select h.ts, h.idObject, h.idUser, h.Param, h.Result from HistoActor h where h.idObject = ? and h.ts between ? and ? order by h.ts", idObject, startTS.Unix(), endTS.Unix())
	}
	if err != nil {
		glog.Error(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var curVal HistoActor
		err = rows.Scan(&curVal.Ts, &curVal.IdObject, &curVal.IdUser, &curVal.Param, &curVal.Result)
		if err != nil {
			glog.Error(err)
			return
		}
		values = append(values, curVal)
	}
	if err = rows.Err(); err != nil {
		glog.Error(err)
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
		err = errors.New(fmt.Sprintf("Fail to write values, db.Begin : %s", err))
		return
	}
	objectid = values[0].IdObject

	if objectid <= 0 {
		// New Id needed
		row := tx.QueryRow("select max(idObject)+1 from ItemFieldVal")
		if row == nil {
			glog.Error("Fail to select next objectid")
			tx.Rollback()
			return
		}
		if err = row.Scan(&objectid); err != nil {
			glog.Errorf("Fail to scan next objectid : %s", err)
			tx.Rollback()
			return
		}
	} else {
		// Delete existing ItemFieldVal
		_, err = db.Exec(fmt.Sprintf("delete from ItemFieldVal where idObject = %d", objectid))
		if err != nil {
			glog.Errorf("Fail to delete ItemFieldVal for objectid=%d : %s", objectid, err)
			tx.Rollback()
			return
		}
	}

	stmt, err := tx.Prepare("insert into ItemFieldVal (idObject, idField, intVal, floatVal, textVal, byteVal) values(?, ?, ?, ?, ?, ?)")
	if err != nil {
		glog.Errorf("Fail to prepare insert into ItemFieldVal : %s", err)
		tx.Rollback()
		return
	}
	defer stmt.Close()

	for _, val := range values {
		_, err = stmt.Exec(objectid, val.IdField, val.IntVal, val.FloatVal, val.TextVal, val.ByteVal)
		if err != nil {
			glog.Errorf("Fail to exec insert into ItemFieldVal (objId=%d, %v) : %s", objectid, val, err)
			tx.Rollback()
			return
		}
	}
	tx.Commit()

	return
}
