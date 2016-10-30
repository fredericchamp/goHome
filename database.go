// database.go
package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
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
func openDB() (db *sql.DB, err error) {
	db, err = sql.Open("sqlite3", dbFileName)
	if err != nil {
		glog.Errorf("Failed to open sqlite3(%s) : ", dbFileName, err)
		return
	}
	return
}

// getGlobalParam : fetch param value from db table goHome
// if id > 0 query db using this id (ignoring scope and name parameter) else query using scope and name
// If multiple rows received only the first is read
func getGlobalParam(db *sql.DB, id int, scope string, name string) (value string, err error) {
	if db == nil {
		db, err = openDB()
		if err != nil {
			return
		}
		defer db.Close()
	}

	var rows *sql.Rows

	if id > 0 {
		rows, err = db.Query("select value from goHome where id = ?", id)
	} else {
		rows, err = db.Query("select value from goHome where scope = ? and name = ?", scope, name)
	}
	if err != nil {
		glog.Errorf("Failed to read global param (%d,%s,%s) : %s ", id, scope, name, err)
		return
	}
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&value)
	if err != nil {
		glog.Errorf("Failed to read global param (%d,%s,%s) : %s ", id, scope, name, err)
		return
	}

	if err = rows.Err(); err != nil {
		glog.Errorf("Failed to read global param (%d,%s,%s) : %s ", id, scope, name, err)
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
// If idItem > 0 return Item with given Id else return all Items
func getManageItems(db *sql.DB, idItem int, idItemType itemType) (items []Item, err error) {
	if db == nil {
		db, err = openDB()
		if err != nil {
			return
		}
		defer db.Close()
	}

	var curItem Item
	var rows *sql.Rows

	switch {
	case idItem > 0:
		rows, err = db.Query("select id, Name, idProfil, idItemType, idMasterItem, icone from Item where id = ? ", idItem)
	case idItemType > 0:
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
	Rules      string
}

// getItemFields select item fields description.
// If idItem > 0 return item fields for the given item
// Else If idObject > 0 return item fields for the item of the given object
// Else return fields for all item
func getItemFields(db *sql.DB, idItem int, idObject int) (fields []ItemField, err error) {
	if db == nil {
		db, err = openDB()
		if err != nil {
			return
		}
		defer db.Close()
	}

	var curField ItemField
	var rows *sql.Rows

	switch {
	case idItem > 0:
		rows, err = db.Query("select id, idItem, nOrder, Name, idDataType, Helper, Rules from ItemField where idItem = ? order by nOrder", idItem)
	case idObject > 0:
		rows, err = db.Query("select f.id, f.idItem, f.nOrder, f.Name, f.idDataType, f.Helper, f.Rules from ItemField f, ItemFieldVal v where f.id = v.idField and v.idObject = ? order by nOrder", idObject)
	default:
		rows, err = db.Query("select id, idItem, nOrder, Name, idDataType, Helper, Rules from ItemField order by idItem, nOrder")
	}
	if err != nil {
		glog.Error(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&curField.Id, &curField.IdItem, &curField.NOrder, &curField.Name, &curField.IdDataType, &curField.Helper, &curField.Rules)
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

// getItemFieldValues select values for item fields.
// If idObject > 0 return values for the given object
// Else return all values for all objects with idField is part of idItem description
func getItemFieldValues(db *sql.DB, idItem int, idObject int) (values []ItemFieldVal, err error) {
	if db == nil {
		db, err = openDB()
		if err != nil {
			return
		}
		defer db.Close()
	}

	var curVal ItemFieldVal
	var rows *sql.Rows

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
