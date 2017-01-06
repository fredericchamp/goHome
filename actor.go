// actor.go
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang/glog"
)

func init() {
	RegisterInternalFunc(ActorFunc, "GoHomeExit", GoHomeExit)
	RegisterInternalFunc(ActorFunc, "SendSMS", SendSMS)
	RegisterInternalFunc(ActorFunc, "SendMail", SendMail)
}

// triggerActorById : trigger actor function using ActCmd, restirered parameter 'ActParam' and dynamic param 'param'
func triggerActorById(actorId int, userId int, param string) (result string, err error) {

	objs, err := getHomeObjects(nil, ItemIdNone, actorId)
	if err != nil {
		return
	}
	if len(objs) <= 0 {
		err = errors.New(fmt.Sprintf("No actor with id = %d", actorId))
		glog.Error(err)
	}
	actor := objs[0]

	result, err = triggerObjActor(actor, userId, param)
	return
}

// triggerObjActor : trigger actor function using ActCmd, registered parameter 'ActParam' and dynamic param 'param'
func triggerObjActor(actor HomeObject, userId int, param string) (result string, err error) {
	result = "Failed"
	actName, err := actor.getStrVal("Name")
	if err != nil {
		return
	}
	actCmd, err := actor.getStrVal("ActCmd")
	if err != nil {
		return
	}
	actParam, err := actor.getStrVal("ActParam")
	if err != nil {
		return
	}
	isInternal, err := actor.getIntVal("IsInternal")
	if err != nil {
		return
	}

	if isInternal != 0 {
		result, err = CallInternalFunc(ActorFunc, actCmd, actParam, param)
	} else {
		result, err = ExecExternalCmd(actCmd, actParam, param)
	}

	glog.Infof("Actor : user#%d : %s(%s) => %s", userId, actName, param, result)
	if glog.V(1) {
		glog.Infof("Actor : %s('%s','%s')", actCmd, actParam, param)
	}

	go recordActorResult(actor, userId, param, result)

	return
}

// recordActorResult : store in DB param and result for an actor
func recordActorResult(actor HomeObject, userId int, param string, result string) {
	db, err := openDB()
	if err != nil {
		return
	}
	defer db.Close()

	actorId := actor.getId()

	_, err = db.Exec("insert into HistoActor values ( ?, ?, ?, ?, ?);", time.Now().Unix(), actorId, userId, param, result)
	if err != nil {
		glog.Errorf("Fail to store result (%s) for actor %d : %s ", result, actorId, err)
	}

	if glog.V(1) {
		glog.Infof("recordActorResult : %d - %s - %s", time.Now().Unix(), param, result)
	}
}

// -----------------------------------------------
// -----------------------------------------------

// GoHomeExit : End server execution.
// No built-in autorestart, should use systemd for restart (see setup/goHome.service)
// param2 can be use to give a delay before exit
func GoHomeExit(param1 string, param2 string) (result string, err error) {
	wait := time.Second
	if param1 != "" {
		wait, err = time.ParseDuration(param2)
		if err != nil {
			glog.Errorf("GoHomeExit : error parsigng duration parm : %s", err)
			wait = time.Second
		}
	}

	time.AfterFunc(wait, func() { goHomeExitChan <- true })

	return fmt.Sprintf("Server exit in %v", wait), nil
}

// -----------------------------------------------
// -----------------------------------------------

// SendSMS : send a SMS using param1 device
// param1 : serial port (i.e. /dev/ttyAMA0 on rpi). Other device type may be added in the futur
// param2 : "<phoneNum>|<message>" with phoneNum := "[0-9]+"
func SendSMS(param1 string, param2 string) (result string, err error) {
	serialPort := param1
	pTab := strings.Split(param2, "|")
	if len(pTab) <= 1 {
		err = errors.New("SendSMS bad parameter '" + param2 + "', expecting '<phoneNum>|<message>'")
		glog.Errorf(err.Error())
		result = "bad parameter"
		return
	}
	phoneNum := pTab[0]
	message := strings.Join(pTab[1:], "|")

	result, err = SerialATSMS(serialPort, phoneNum, message)

	return
}

// -----------------------------------------------
// -----------------------------------------------

// SendMail : send a mail
// param1 : JSON param : {"server":"smtp.gmail.com","port":"587","tls":true,"account":"sender@gmail.com","password":"****",......}
// param2 : "<to>|<subject>|<message>" ... so subject can't include char '|'
func SendMail(param1 string, param2 string) (result string, err error) {
	var mail MailInfo

	err = json.Unmarshal([]byte(param1), &mail)
	if err != nil {
		glog.Errorf("Fail to unmarshal MailInfo (%s) : %s", param1, err)
		return
	}

	pTab := strings.Split(param2, "|")
	if len(pTab) <= 1 {
		err = errors.New("SendMail bad parameter '" + param2 + "', expecting '<to>|<subject>|<message>'")
		glog.Errorf(err.Error())
		result = "bad parameter"
		return
	}

	mail.To = pTab[0]
	mail.Subject = pTab[1]
	mail.Message = strings.Join(pTab[2:], "|")

	result, err = smtpSendMail(mail)

	return
}
