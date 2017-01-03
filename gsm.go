// gsm.go
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/tarm/serial"
)

var gsmPortLock sync.Mutex
var gsmPort *serial.Port
var gsmDevice string

// -----------------------------------------------
// GSM operation
// Currently handeling only one GSM module on serial port
// -----------------------------------------------

func gsmSetup(db *sql.DB) (err error) {

	if db == nil {
		if db, err = openDB(); err != nil {
			return
		}
		defer db.Close()
	}

	gsmDevice, err = getGlobalParam(db, "GSM", "device")
	if err != nil {
		glog.Errorf("gsmSetup : no device found => GSM disable")
		err = nil
		return
	}

	serialConf := &serial.Config{
		Name:        gsmDevice,
		Baud:        9600,
		Size:        8,
		Parity:      serial.ParityNone,
		StopBits:    serial.Stop1,
		ReadTimeout: time.Millisecond * 100,
	}
	gsmPort, err = serial.OpenPort(serialConf)
	if err != nil {
		glog.Errorf("gsmSetup openPort '%s' failed : %s", gsmDevice, err)
		gsmDevice = ""
		gsmPort = nil
		return
	}

	if err = gsmActivate(db); err != nil {
		gsmDevice = ""
		gsmPort = nil
		return
	}

	if err = gsmRegister(); err != nil {
		gsmDevice = ""
		gsmPort = nil
		return
	}

	if glog.V(1) {
		glog.Infof("gsmSetup Done")
	}

	return
}

//
func gsmCleanup() {
	if gsmPort != nil {
		if err := gsmPort.Close(); err != nil {
			glog.Errorf("gsmCleanup error closing device : %s", err)
		}
	}
	gsmDevice = ""
	gsmPort = nil
	if glog.V(1) {
		glog.Infof("gsmCleanup Done")
	}

	return
}

//
func gsmActorOnOff(db *sql.DB) (err error) {
	return gsmActor(db, "onOffActorId")
}

//
func gsmActorReset(db *sql.DB) (err error) {
	return gsmActor(db, "resetActorId")
}

//
func gsmActor(db *sql.DB, actorParam string) (err error) {
	strActorId, err := getGlobalParam(db, "GSM", actorParam)
	if err != nil {
		glog.Errorf("gsmActor : actor '%s' not found => no ation", actorParam)
		err = nil
		return
	}

	actorId, err := strconv.Atoi(strActorId)
	if err != nil {
		glog.Errorf("gsmActor : bad actorId '%s' for '%s' : %s", strActorId, actorParam, err)
		return
	}

	_, err = triggerActorById(actorId, -1, "")

	return
}

//
func gsmActivate(db *sql.DB) (err error) {

	// First check
	if err = gsmSendCmdAT("AT\r", "OK\r", time.Millisecond*1500); err != nil {

		// No response on first check => device may be off, try to turn it on
		err = gsmActorOnOff(db)

		// Wait for device to start
		time.Sleep(time.Second * 9)

		// Second check
		if err = gsmSendCmdAT("AT\r", "OK\r", time.Millisecond*1500); err != nil {
			// Still no response
			err = errors.New("gsmActivate failed (OnOff)")
			glog.Error(err.Error())
			return
		}
	}

	// GSM module initialization

	// Reset to factory settings
	if err = gsmSendCmdAT("AT&F\r", "OK\r", time.Second*2); err != nil {
		err = errors.New("gsmActivate failed (factory settings)")
		glog.Error(err.Error())
		return
	}

	// Request calling line identification
	if err = gsmSendCmdAT("AT+CLIP=1\r", "OK\r", time.Second*2); err != nil {
		err = errors.New("gsmActivate failed (line identification)")
		glog.Error(err.Error())
		return
	}

	// Module error code 0->disable; 1->numeric; 2->verbose
	if err = gsmSendCmdAT("AT+CMEE=0\r", "OK\r", time.Second*2); err != nil {
		err = errors.New("gsmActivate failed (error code)")
		glog.Error(err.Error())
		return
	}

	// Set the SMS mode to text
	if err = gsmSendCmdAT("AT+CMGF=1\r", "OK\r", time.Second*2); err != nil {
		err = errors.New("gsmActivate failed (SMS mode to text)")
		glog.Error(err.Error())
		return
	}

	// Disable messages about new SMS from the GSM module
	if err = gsmSendCmdAT("AT+CNMI=2,0\r", "OK\r", time.Second*2); err != nil {
		err = errors.New("gsmActivate failed (disable messages about new SMS)")
		glog.Error(err.Error())
		return
	}

	// send AT command to init memory for SMS in the SIM card
	// response: +CPMS: <usedr>,<totalr>,<usedw>,<totalw>,<useds>,<totals>
	if err = gsmSendCmdAT("AT+CPMS=\"SM\",\"SM\",\"SM\"\r", "OK\r", time.Second*10); err != nil {
		err = errors.New("gsmActivate failed (init memory for SMS)")
		glog.Error(err.Error())
		return
	}

	// select phonebook memory storage
	if err = gsmSendCmdAT("AT+CPBS=\"SM\"\r", "OK\r", time.Second*2); err != nil {
		err = errors.New("gsmActivate failed (phonebook memory storage)")
		glog.Error(err.Error())
		return
	}

	if glog.V(1) {
		glog.Infof("gsmActivate => Device is ready")
	}

	return
}

//
func gsmRegister() (err error) {
	// Register to the network
	// response: "+CREG: 0,1" or "+CREG: 0,2" or "+CREG: 0,5"
	if err = gsmSendCmdAT("AT+CREG?\r", "OK\r", time.Second*5); err != nil {
		err = errors.New("gsmRegister failed")
		glog.Error(err.Error())
		return
	}
	return
}

//
func gsmReset() (err error) {
	db, err := openDB()
	if err != nil {
		return
	}
	defer db.Close()

	// Hard reset
	if err = gsmActorReset(db); err != nil {
		return
	}

	// Wait for device to start
	time.Sleep(time.Second * 9)

	// Check
	if err = gsmSendCmdAT("AT\r", "OK\r", time.Millisecond*1500); err != nil {
		// No response, try Device activation
		if err = gsmActivate(db); err != nil {
			return
		}
	}

	if err = gsmRegister(); err != nil {
		return
	}

	return
}

// gsmWaitForCR : Read from gsmPort until cr received or timeout expired
func gsmWaitForCR(cr string, timeout time.Duration) (err error) {
	if gsmPort == nil {
		err = errors.New("gsmWaitForCR : gsmPort not initialized")
		glog.Errorf(err.Error())
		return
	}

	chanStop := make(chan bool)

	time.AfterFunc(timeout, func() { chanStop <- true })

	var received string
	buf := make([]byte, 32)
	for true {
		select {
		case <-chanStop:
			err = errors.New(fmt.Sprintf("gsmWaitForCR : timeout (%v)", timeout))
			return
		default:
			var n int
			n, err = gsmPort.Read(buf)
			if err != nil && err.Error() != "EOF" {
				err = errors.New(fmt.Sprintf("gsmWaitForCR : %s", err))
				return
			}
			if n > 0 {
				received = fmt.Sprintf("%s%s", received, string(buf[0:n]))
				if strings.Contains(received, cr) {
					return
				}
			}
		}
	}

	return
}

// gsmSendCmdAT : Send cmdAT command and wait for cr
// If no cr received before timeout then return an error
func gsmSendCmdAT(cmdAT string, cr string, timeout time.Duration) (err error) {

	if glog.V(2) {
		glog.Infof("gsmSendCmdAT '(%s','%s',%v)", strings.Replace(cmdAT, "\r", "\\r", -1), strings.Replace(cr, "\r", "\\r", -1), timeout)
	}

	if gsmPort == nil {
		err = errors.New("gsmSendCmdAT : gsmPort not initialized")
		glog.Errorf(err.Error())
		return
	}

	// Only 1 thread writing cmd to gsmPort at a time
	gsmPortLock.Lock()
	defer gsmPortLock.Unlock()

	gsmPort.Flush()

	n, err := gsmPort.Write([]byte(cmdAT))
	if err != nil || n != len(cmdAT) {
		glog.Errorf("gsmSendCmdAT Write failed : %s", err)
		return
	}

	if err = gsmWaitForCR(cr, timeout); err != nil {
		glog.Errorf("gsmSendCmdAT failed : %s", err)
		return
	}

	if glog.V(2) {
		glog.Infof("gsmSendCmdAT done (%s)", strings.Replace(cmdAT, "\r", "\\r", -1))
	}

	return
}

// SerialATSMS : send a SMS to phoneNum using AT cmd send on serial device serialPort
func SerialATSMS(serialPort string, phoneNum string, message string) (result string, err error) {

	if gsmDevice != serialPort {
		err = errors.New(fmt.Sprintf("SerialATSMS unknown device '%s', existing device is '%s'", serialPort, gsmDevice))
		glog.Error(err.Error())
		result = "Fail"
		return
	}

	if err = gsmSendCmdAT("AT+CMGS=\""+phoneNum+"\"\r", ">", time.Second*10); err != nil {
		result = "Fail"
		return
	}

	// \x1a == (char)26 == ^Z
	if err = gsmSendCmdAT(message+"\x1a\r", "OK\r", time.Second*10); err != nil {
		gsmReset() // Try reset to prepare next try
		result = "Fail"
		return
	}

	result = "Done"

	return
}
