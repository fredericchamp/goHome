// actor.go
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"net/http"
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
	if glog.V(2) {
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

	if glog.V(2) {
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

// SendSMS : send a SMS using a Android device with SMS Gateway Ultimate (https://play.google.com/store/apps/details?id=com.icecoldapps.smsgatewayultimate)
// param1 : SMS Gateway <serveur[:port]> i.e. 192.168.43.1:1116
// param2 : "<phoneNum> <message>" with phoneNum := "[0-9]+"
func SendSMS(param1 string, param2 string) (result string, err error) {
	var Url *url.URL
	Url, err = url.Parse("http://" + param1 + "/send.html")
	if err != nil {
		err = errors.New("SendSMS bad url '" + param1 + "', " + err.Error())
		glog.Errorf(err.Error())
		result = "bad parameter"
		return
	}
	pTab := strings.Split(param2, " ")
	if len(pTab) <= 1 {
		err = errors.New("SendSMS bad parameter '" + param2 + "', expecting '<phoneNum> <message>'")
		glog.Errorf(err.Error())
		result = "bad parameter"
		return
	}
	parameters := url.Values{}
    parameters.Add("smstype", "sms")
    parameters.Add("smsto", pTab[0])
	parameters.Add("smsbody", strings.Join(pTab[1:], " "))
    Url.RawQuery = parameters.Encode()

	resp, err := http.Get(Url.String())
	if err != nil {
		err = errors.New("SendSMS gateway err '" + Url.String() + "', " + err.Error())
		glog.Errorf(err.Error())
		result = "SendSMS gateway err"
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if ! strings.Contains(fmt.Sprintf("%s",body),"The SMS has been sent") {
		result = "Missing Gateway confirmation"
		return
	}

	result = "Done"
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
