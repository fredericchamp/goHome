// util.go
package main

import (
	"crypto/tls"
	"database/sql"
	"errors"
	"fmt"
	"net"
	//	"net/mail"
	"net/smtp"
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
// MAIL
// -----------------------------------------------
type MailInfo struct {
	Server   string
	Host     string
	Port     int
	Tls      bool
	Account  string
	Password string
	From     string // TODO : should use : FromAddr mail.Adresse
	//	ReplyTo  string // TODO : should use : ReplyToAddr mail.Adresse
	To      string // TODO : should use : ToAddr mail.Adresse
	Subject string
	Message string
}

func smtpSendMail(mailInfo MailInfo) (result string, err error) {

	// Check param
	if mailInfo.Host == "" ||
		mailInfo.Account == "" ||
		mailInfo.Password == "" ||
		mailInfo.To == "" {
		err = errors.New(fmt.Sprintf("Bad parameter : %v", mailInfo))
		result = "Bad parameter"
		return
	}
	if mailInfo.From == "" {
		mailInfo.From = mailInfo.Account
	}
	if mailInfo.Port == 0 {
		mailInfo.Port = 25 // default SMTP port
	}
	if mailInfo.Server == "" {
		mailInfo.Server = fmt.Sprintf("%s:%d", mailInfo.Host, mailInfo.Port)
	}

	var clientSmtp *smtp.Client

	if glog.V(2) {
		glog.Infof("mailInfo=%v", mailInfo)
	}

	if mailInfo.Tls {
		tlsconfig := &tls.Config{
			//InsecureSkipVerify: true, // for testing with self sign server certificate
			ServerName: mailInfo.Host,
		}

		conn, err := tls.Dial("tcp", mailInfo.Server, tlsconfig)
		if err != nil {

			// If direct tls Dial fail, fall back to StartTLS method
			conn, err := net.Dial("tcp", mailInfo.Server)
			if err != nil {
				glog.Errorf("smtpSendMail : net.Dial fail : %s", err)
				return "tls error", err
			}

			clientSmtp, err = smtp.NewClient(conn, mailInfo.Host)
			if err != nil {
				glog.Errorf("smtpSendMail : smtp.NewClient fail : %s", err)
				return "NewClient error", err
			}

			err = clientSmtp.StartTLS(tlsconfig)
			if err != nil {
				glog.Errorf("smtpSendMail : StartTLS fail : %s", err)
				return "StartTLS error", err
			}

		} else {

			clientSmtp, err = smtp.NewClient(conn, mailInfo.Host)
			if err != nil {
				glog.Errorf("smtpSendMail : smtp.NewClient fail : %s", err)
				return "NewClient error", err
			}
		}

	} else {

		clientSmtp, err = smtp.Dial(mailInfo.Server)
		if err != nil {
			glog.Errorf("smtpSendMail : smtp.Dial fail : %s", err)
			return "Dial error", err
		}
	}
	defer clientSmtp.Quit()

	// Auth
	auth := smtp.PlainAuth("", mailInfo.Account, mailInfo.Password, mailInfo.Host)
	if err = clientSmtp.Auth(auth); err != nil {
		glog.Errorf("smtpSendMail : smtp.PlainAuth fail : %s", err)
		return "PlainAuth error", err
	}

	// From
	if err = clientSmtp.Mail(mailInfo.From); err != nil {
		glog.Errorf("smtpSendMail : .Mail fail : %s", err)
		return ".Mail error", err
	}

	// To
	if err = clientSmtp.Rcpt(mailInfo.To); err != nil {
		glog.Errorf("smtpSendMail : .Rcpt fail : %s", err)
		return ".Rcpt error", err
	}

	// Data
	w, err := clientSmtp.Data()
	if err != nil {
		glog.Errorf("smtpSendMail : .Data fail : %s", err)
		return ".Data error", err
	}

	// Setup message
	message := fmt.Sprintf("From: %s\r\n", mailInfo.From)
	message += fmt.Sprintf("To: %s\r\n", mailInfo.To)
	message += fmt.Sprintf("Subject: %s\r\n", mailInfo.Subject)
	message += fmt.Sprintf("\r\n%s", mailInfo.Message)

	_, err = w.Write([]byte(message))
	if err != nil {
		glog.Errorf("smtpSendMail : .Write fail : %s", err)
		return ".Write error", err
	}

	err = w.Close()
	if err != nil {
		glog.Errorf("smtpSendMail : .Close fail : %s", err)
		return ".Close error", err
	}

	return
}
