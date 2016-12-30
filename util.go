// util.go
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
)

// -----------------------------------------------
// Backup
// -----------------------------------------------

const (
	TagArchiveName = "@archiveName@"
	TagBackupDir   = "@backupDir@"
)

//
func backupSetup(db *sql.DB, datetimeParam string) (err error) {

	// Read backup date/time from DB. Syntax close to crontab entries, see nextMatchingDatetime
	if len(datetimeParam) <= 0 {
		if db == nil {
			if db, err = openDB(); err != nil {
				return
			}
			defer db.Close()
		}
		datetimeParam, err = getGlobalParam(db, "Backup", "date/time")
		if err != nil {
			return
		}
	}
	if len(datetimeParam) <= 0 {
		err = errors.New("backup date/time parameter missing")
		glog.Errorf("backupSetup : %s", err)
		return
	}

	// Calc next date & time matching datetimeParam
	nextBackupAt, err := nextMatchingDatetime(datetimeParam, time.Now())
	if err != nil {
		return
	}

	// Set next backup to run when expected
	time.AfterFunc(nextBackupAt.Sub(time.Now()), doBackup)

	if glog.V(1) {
		glog.Infof("backupSetup Done")
	}

	return nil
}

// doBackup : perform backup using paramter in db
func doBackup() {
	db, err := openDB()
	if err != nil {
		glog.Errorf("doBackup : fail to open BDD : %s", err)
		// TODO : send Mail or SMS
		return
	}
	defer db.Close()

	// Read backup parameters
	backupParam, err := getGlobalParamList(db, "Backup")
	if err != nil {
		// TODO : send Mail or SMS
		return
	}

	backupDir, exist := backupParam["dir"]
	if !exist {
		glog.Error("doBackup : backup dir parameter missing")
		// TODO : send Mail or SMS
		return
	}

	// Backup database
	err = backupDB("", backupDir)
	if err != nil {
		// TODO : send Mail or SMS
		return
	}

	// Backup files :
	// Will Exec each string Val for parameter with Name like 'files_%'
	for name, cmd := range backupParam {
		if strings.HasPrefix(name, "files_") {
			cmd = strings.Replace(cmd, TagBackupDir, backupDir, -1)
			_, err = execCommand(cmd)
			if err != nil {
				// TODO : send Mail or SMS
				return
			}
		}
	}

	// Build archive backup file if "archive" parameter is present
	archiveName := fmt.Sprintf("%s_%s", time.Now().Format("20060102_150405"), "goHome.tar.gz")
	archiveName = filepath.Join(os.TempDir(), archiveName)

	cmd, archExist := backupParam["archive"]
	if archExist {
		cmd = strings.Replace(cmd, TagArchiveName, archiveName, -1)
		cmd = strings.Replace(cmd, TagBackupDir, backupDir, -1)
		_, err = execCommand(cmd)
		if err != nil {
			// TODO : send Mail or SMS
			return
		}
	}

	// Externalize backup if "externalize" parameter is present
	cmd, extExist := backupParam["externalize"]
	if extExist {
		cmd = strings.Replace(cmd, TagArchiveName, archiveName, -1)
		cmd = strings.Replace(cmd, TagBackupDir, backupDir, -1)
		_, err = execCommand(cmd)
		if err != nil {
			// TODO : send Mail or SMS
			return
		}
	}

	// Cleanup
	cmd, exist = backupParam["cleanup"]
	if exist {
		cmd = strings.Replace(cmd, TagArchiveName, archiveName, -1)
		cmd = strings.Replace(cmd, TagBackupDir, backupDir, -1)
		_, err = execCommand(cmd)
		if err != nil {
			// TODO : send Mail or SMS
			return
		}
	}

	// Setup next backup
	if err := backupSetup(db, backupParam["date/time"]); err != nil {
		glog.Errorf("doBackup : fail to setup next backup : %s", err)
		// TODO : send Mail or SMS
	}

	if glog.V(1) {
		if archExist {
			glog.Infof("backup Done => %s (externalize=%v)", archiveName, extExist)
		} else {
			glog.Infof("backup Done => %s (externalize=%v)", backupDir, extExist)
		}
	}

	return
}

// -----------------------------------------------
// -----------------------------------------------

// nextMatchingdatetime Calc next date & time after now matching datetimeParam
// datetimeParam : "(*|[0-9]+) (*|[0-9]+) (*|[0-9]+) ..." (simple crontab form, only "m h dom" are used)
func nextMatchingDatetime(datetimeParam string, now time.Time) (nextAt time.Time, err error) {

	datetimeParam = cleanSpaces(datetimeParam)
	timecrit := strings.Split(datetimeParam, " ") // m h dom ...

	m := now.Minute()
	if timecrit[0] != "*" {
		m, err = strconv.Atoi(timecrit[0])
		if err != nil {
			glog.Errorf("nextMatchingDatetime : bad date/time 'm' : %s : %s", datetimeParam, err)
			return
		}
	}
	h := now.Hour()
	if timecrit[1] != "*" {
		h, err = strconv.Atoi(timecrit[1])
		if err != nil {
			glog.Errorf("nextMatchingDatetime : bad date/time 'h' : %s : %s", datetimeParam, err)
			return
		}
	}
	d := now.Day()
	if timecrit[2] != "*" {
		d, err = strconv.Atoi(timecrit[2])
		if err != nil {
			glog.Errorf("nextMatchingDatetime : bad date/time 'dom' : %s : %s", datetimeParam, err)
			return
		}
	}

	nextAt = time.Date(now.Year(), now.Month(), d, h, m, 0, 0, now.Location())

	maxDayOfMonth := 31
	for nextAt.Before(now) {
		if timecrit[0] == "*" && nextAt.Minute() < 59 {
			nextAt = nextAt.Add(time.Minute)
		} else if timecrit[1] == "*" && nextAt.Hour() < 23 {
			if timecrit[0] == "*" {
				nextAt = nextAt.Add(time.Minute)
			} else {
				nextAt = nextAt.Add(time.Hour)
			}
		} else if timecrit[2] == "*" && nextAt.Day() < maxDayOfMonth {
			curMonth := nextAt.Month()
			plus := nextAt
			if timecrit[0] == "*" && timecrit[1] == "*" {
				plus = plus.Add(time.Minute)
			} else if timecrit[0] == "*" && timecrit[1] != "*" {
				plus = time.Date(plus.Year(), plus.Month(), plus.Day(), plus.Hour(), 0, 0, 0, now.Location())
				plus = plus.Add(24 * time.Hour)
			} else if timecrit[1] == "*" {
				plus = nextAt.Add(time.Hour)
			} else {
				plus = plus.Add(24 * time.Hour)
			}
			if curMonth == plus.Month() {
				nextAt = plus
			} else {
				maxDayOfMonth = nextAt.Day()
			}
		} else {
			y := nextAt.Year()
			mo := int(nextAt.Month()) + 1
			if mo > 12 {
				mo = 1
				y += 1
			}
			m = nextAt.Minute()
			if timecrit[0] == "*" {
				m = 0
			}
			h = nextAt.Hour()
			if timecrit[1] == "*" {
				h = 0
			}
			d = nextAt.Day()
			if timecrit[2] == "*" {
				d = 0
			}

			nextAt = time.Date(y, time.Month(mo), d, h, m, 0, 0, now.Location())
		}
	}

	if glog.V(1) {
		glog.Infof("nextMatchingdatetime '%s' = %v", datetimeParam, nextAt)
	}

	return
}

// -----------------------------------------------
// -----------------------------------------------

// cleanSpaces remove leading and trailing and replace any sequence of multiple spaces by one space
func cleanSpaces(in string) (out string) {
	out = strings.TrimSpace(in)
	ln := len(out) + 1
	for ln > len(out) {
		ln = len(out)
		out = strings.Replace(out, "  ", " ", -1)
	}
	return
}

// -----------------------------------------------
// -----------------------------------------------

// execCommand : execute cmd string and return output (stdout+stderr)
func execCommand(cmd string) (result string, err error) {
	cmdTab := strings.Split(cleanSpaces(cmd), " ")

	cmdStruct := exec.Command(cmdTab[0], cmdTab[1:]...)

	resultBytes, err := cmdStruct.CombinedOutput()
	if err != nil {
		glog.Errorf("execCommand fail : %v : %v", cmdTab, err)
		glog.Errorf("CombinedOutput = %s", string(resultBytes))
		return
	}

	return string(resultBytes), nil
}

// -----------------------------------------------
// -----------------------------------------------
