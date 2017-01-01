// gsm.go
package main

import (
	"database/sql"
	"errors"
	"fmt"
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

	if err = gsmTurnOn(); err != nil {
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
func gsmTurnOn() (err error) {
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

// gsmSendCmdAT : Send cmdAT command and wait for "OK\r"
// If no "OK\r" received before timeout then return an error
func gsmSendCmdAT(cmdAT string, cr string, timeout time.Duration) (err error) {

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

	if glog.V(1) {
		glog.Infof("gsmSendCmdAT done (%s)", cmdAT)
	}

	return
}

//

// SerialATSMS : send a SMS to phoneNum using AT cmd send on serial device serialPort
func SerialATSMS(serialPort string, phoneNum string, message string) (result string, err error) {

	if err = gsmSendCmdAT("AT+CMGS=\""+phoneNum+"\"\r", "\r>", time.Second*10); err != nil {
		result = "Fail"
		return
	}

	// \x1a == (char)26 == ^Z
	if err = gsmSendCmdAT(message+"\x1a\r", "OK\r", time.Second*10); err != nil {
		result = "Fail"
		return
	}

	//	if (RPIOK != rpiGsmAtCmd(fdSerial, szATCmd, 10000, 5000, "OK"))
	//	{
	//		rpiGsmReset( fdSerial); // SMS send failed => try rpiGsmReset
	//		return RPIKO;
	//	}

	result = "Done"

	return
}
