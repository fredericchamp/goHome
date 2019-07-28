// http.go
package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
)

var htmlEscaper = strings.NewReplacer(
	`&`, "&amp;",
	`'`, "&#39;",
	`<`, "&lt;",
	`>`, "&gt;",
	`"`, "&#34;",
)
 
// -----------------------------------------------
const htmlHeader = `
<!-- HEADER -->
<html>
<head>
	<title>goHome</title>
	<meta http-equiv="Content-Type" content="text/html; charset=iso-8859-1">
	<meta name="description" content="goHome">
	<meta name="viewport" content="width=device-width, initial-scale=1">
</head>
<body 
	style="background-image: url(/images/PlageMoorea.jpg);background-repeat: no-repeat;background-size: 100%%;"
	onload="%s"
>
	<div align="center" onclick="window.location.reload(true);" >%s</div>
	<br><br>
	<div id="maindiv" align="center">
`

const htmlUserCode = `
		<form action="#" method="POST">
			<input type="password" id="usercode" name="usercode" value="%s">
			<input type="submit" value="Submit">
		</form>
`

const htmlAction = `
		<table><tr>
			<td align="center">
				<form action="#" method="POST">
					<input type="hidden" id="usercode" name="usercode" value="%s">
					<input type="hidden" id="objectid" name="objectid" value="3">
					<input type="image" src="/images/portail.jpg" border="1px" width="120" height="120">
				</form>
			</td>
			<td align="center">
				<form action="#" method="POST">
					<input type="hidden" id="usercode" name="usercode" value="%s">
					<input type="hidden" id="objectid" name="objectid" value="4">
					<input type="image" src="/images/garage.jpg"  border="1px" width="120" height="120">
				</form>
			</td>
		</tr><tr>
			<td align="center" colspan="2" >
				<br><br>
				<form action="#" method="POST">
					<input type="hidden" id="usercode" name="usercode" value="%s">
					<input type="hidden" id="objectid" name="objectid" value="0">
					<input type="image" src="/capture/simpleAlarm.jpg" border="2px" width="512" height="384">
				</form>
			</td>
		</tr></table>
`

const htmlFooter = `
	</div>
<!-- FOOTER -->
	<p align="center">-*-</p>
	<p align="center">%s</p>
</body>
</html>
`

// -----------------------------------------------

func getFormStrVal(form url.Values, key string, idx int) (strVal string, err error) {
	val, found := form[key]
	if !found {
		err = errors.New(fmt.Sprintf(`{"error":"Missing '%s' in form data (%s)"}`, key, form))
		return
	}
	strVal = val[idx]
	return
}

func writeApiError(w http.ResponseWriter, errMsg string) {
	if glog.V(1) {
		glog.Infof("writeApiError : '%s'", errMsg)
	}
	w.Write(apiError(errMsg))
	return
}

// -----------------------------------------------

// defaultResponse : for testing only
//func defaultResponse(w http.ResponseWriter, r *http.Request) {
//	const header = `<!-- HEADER -->
//<html>
//<head>
//<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
//<meta name="viewport" content="width=device-width, initial-scale=1">
//<title>goHome</title>
//</head>
//<body>
//`
//	const footer = `<!-- FOOTER -->
//<br><br>
//<p align="center">-*-</p>
//<p align="center">%s</p>
//</body>
//</html>
//`
//	fmt.Fprintf(w, header)
//	fmt.Fprintf(w, "<p>goHome HTTPS server<p>\n")
//	fmt.Fprintf(w, "<p>goHome version %s</p>\n", goHomeVersion)
//	fmt.Fprintf(w, "<p>URL requested : %s </p>\n", r.URL.Path)
//	fmt.Fprintf(w, "<p>Post params : %s</p>\n", r.Form)
//	fmt.Fprintf(w, footer, time.Now().String())
//}

// -----------------------------------------------

// sendInitPage : landing page for simpleResponse
func sendInitPage(w http.ResponseWriter) {
	fmt.Fprintf(w, htmlHeader, "", "Utilisateur inconnu")
	fmt.Fprintf(w, htmlUserCode, "")
	fmt.Fprintf(w, htmlFooter, time.Now().Format("2 Jan 2006 15:04:05") )
	return
}

// sendActionPage : returning defaut action page
func sendActionPage(w http.ResponseWriter, onload string, userName string, userCode string ) {
	fmt.Fprintf(w, htmlHeader, onload, userName)
	fmt.Fprintf(w, htmlAction, userCode, userCode, userCode)
	fmt.Fprintf(w, htmlFooter, time.Now().Format("2 Jan 2006 15:04:05") )
	return
}
// -----------------------------------------------

// simpleResponse : for simple HTTP client that can't handle modern css
func simpleResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")

	// Parse received request
	if err := r.ParseForm(); err != nil {
		writeApiError(w, fmt.Sprintf("Form parse error '%s' for (%s)", err, r.Form))
		return
	}
	if glog.V(2) {
		glog.Infof("Paramaters = %v",r.Form)
	}

	// Look for a user code
	userCode, err := getFormStrVal(r.Form, "usercode", 0)
	if err != nil {
		// No userCode found => return page with form to input userCode
		if glog.V(2) {
			glog.Infof("No userCode found : %v",err)
		}
		sendInitPage(w)
		return
	}

	// Check userCode validity
	userObj, err := getUserFromCode(nil, userCode)
	if err != nil {
		// Invalid userCode
		if glog.V(2) {
			glog.Infof("Invalid userCode %v",err)
		}
		sendInitPage(w)
		return
	}

	userName, err := userObj.getStrVal("FirstName")
	if err != nil {
		if glog.V(2) {
			glog.Infof("Get user FirstName : %v",err)
		}
		userName = "FirstName"
	}

	// -------------------------- Action --------------------------

	// Found a user for given userCode, get profil
	profil, err := checkApiUser(userObj)
	if err != nil {
		if glog.V(2) {
			glog.Infof("Fail to check Api access : %v",err)
		}
		sendActionPage(w, "", userName, userCode)
		return
	}

	// Check if any action (objectid) received
	objectidStr, err := getFormStrVal(r.Form, "objectid", 0)
	if err != nil {
		// No objectid found => return page with form to input userCode
		if glog.V(2) {
			glog.Infof("No valid ObjectId found : %s - %v", objectidStr, err)
		}
		sendActionPage(w, "", userName, userCode)
		return
	}
	objectid, err := strconv.Atoi(objectidStr)
	if err != nil {
		// Not an int objectidStr
		if glog.V(2) {
			glog.Infof("Bad objectid found : %v",err)
		}
		sendActionPage(w, "", userName, userCode)
		return
	}

	// Check user access to object
	if err := checkAccessToObjectId(profil, objectid); err != nil {
		if glog.V(2) {
			glog.Infof("Acces api check fail : %v",err)
		}
		sendActionPage(w, "", userName, userCode)
		return
	}

	// Trigger asked action
	result, err := triggerActorById(objectid, userObj.getId(), "")
	if err != nil {
		if glog.V(2) {
			glog.Infof("Trigger actionfail : %v",err)
		}
		sendActionPage(w, "", userName, userCode)
		return
	}

	if glog.V(2) {
		glog.Infof("Actor result : %s",result)
	}

	sendActionPage(w, fmt.Sprintf("alert('Action %d : %s')",objectid,htmlEscaper.Replace(result)), userName, userCode)
	return
}

// -----------------------------------------------

// apiHandler : handle API requests
func apiHandler(w http.ResponseWriter, r *http.Request) {
	var userObj HomeObject
	w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")

	// Parse received request
	if err := r.ParseForm(); err != nil {
		writeApiError(w, fmt.Sprintf("Form parse error '%s' for (%s)", err, r.Form))
		return
	}
	command, err := getFormStrVal(r.Form, "command", 0)
	if err != nil {
		writeApiError(w, err.Error())
		return
	}

	var jsonCmde apiCommandSruct
	err = json.Unmarshal([]byte(command), &jsonCmde)
	if err != nil {
		writeApiError(w, fmt.Sprintf("Fail to unmarshal apiCommandSruct (%s) : %s", command, err))
		return
	}

	if glog.V(2) {
		glog.Info(jsonCmde)
	}

	// Handle user info
	if glog.V(2) {
		glog.Infof("User code = %s", jsonCmde.UserCode)
	}

	userCert := r.TLS.PeerCertificates
	if ( len(userCert) > 0 ) {
		userObj, err = getUserFromCert(r.TLS.PeerCertificates)
	} else { // if no user cetificat provided, fall back to user code
		userObj, err = getUserFromCode(nil, jsonCmde.UserCode)
	}
	if err != nil {
		writeApiError(w, err.Error())
		return
	}

	profil, err := checkApiUser(userObj)
	if err != nil {
		writeApiError(w, err.Error())
		return
	}

	if glog.V(2) {
		glog.Infof("User profil %d, Form=%v", profil, r.Form)
	}

	// Handle received command
	switch jsonCmde.Command {

	case apiReadRefList:
		if glog.V(2) {
			glog.Infof("%s (name=%s)", jsonCmde.Command, jsonCmde.Jsonparam)
		}
		w.Write(fctApiRefList(jsonCmde))
		return

	case apiReadCurrentUser:
		if glog.V(2) {
			glog.Infof("%s (user objectid=%d)", jsonCmde.Command, userObj.Values[0].IdObject)
		}
		w.Write(apiObjectResponse(profil, userObj))
		return

	case apiReadItem:
		if glog.V(2) {
			glog.Infof("%s (item=%d)", jsonCmde.Command, jsonCmde.Itemid)
		}
		w.Write(fctApiReadItem(profil, jsonCmde))
		return

	case apiReadObject:
		if glog.V(2) {
			glog.Infof("%s (item=%d, obj=%d)", jsonCmde.Command, jsonCmde.Itemid, jsonCmde.Objectid)
		}
		w.Write(fctApiReadObject(profil, jsonCmde))
		return

	case apiReadSensor:
		if glog.V(2) {
			glog.Infof("%s (objectid=%d)", jsonCmde.Command, jsonCmde.Objectid)
		}
		w.Write(fctApiGetSensorVal(profil, jsonCmde, true))
		return

	case apiGetSensorLastVal:
		if glog.V(2) {
			glog.Infof("%s (objectid=%d)", jsonCmde.Command, jsonCmde.Objectid)
		}
		w.Write(fctApiGetSensorVal(profil, jsonCmde, false))
		return

	case apiReadHistoVal:
		if glog.V(2) {
			glog.Infof("%s (obj=%d, start=%d, end=%d)", jsonCmde.Command, jsonCmde.Objectid, jsonCmde.Startts, jsonCmde.Endts)
		}
		w.Write(fctApiReadHistoVal(profil, jsonCmde))
		return

	case apiReadActorRes:
		if glog.V(2) {
			glog.Infof("%s (objectid=%d, start=%d, end=%d)", jsonCmde.Command, jsonCmde.Objectid, jsonCmde.Startts, jsonCmde.Endts)
		}
		w.Write(fctApiReadActorRes(profil, jsonCmde))
		return

	case apiSaveItem:
		writeApiError(w, fmt.Sprintf("Command %s not ready", jsonCmde.Command))
		return

	case apiSaveObject:
		if glog.V(2) {
			glog.Infof("%s (item=%d, obj=%d)", jsonCmde.Command, jsonCmde.Itemid, jsonCmde.Objectid)
		}
		w.Write(fctApiSaveObject(profil, jsonCmde))
		return

	case apiDeleteItem, apiDeleteObject:
		writeApiError(w, "Delete not available : use apiSave* to toggle IsActive flag or use manual access to DB")
		return

	case apiSendSensorVal: // TODO
		//		if glog.V(1) {
		//			glog.Infof("%s (type=%d, item=%d, obj=%d)", jsonCmde.Command, jsonCmde.Objectid, jsonCmde.Value)
		//		}

		//		objs, err := getHomeObjects(nil, ItemNone, ItemIdNone, jsonCmde.Objectid)
		//		if err != nil {
		//			writeJsonError(w, fmt.Sprintf("%s failed for (%s) : %s", jsonCmde.Command, r.Form, err))
		//			return
		//		}

		// Unmarchal jsonparam to SensorVal
		// load sensor using getHomeObjects
		// Call handleSensorValue(ts, sensor, value)
		writeApiError(w, fmt.Sprintf("Command %s not ready", jsonCmde.Command))
		return

	case apiTriggerActor:
		if glog.V(2) {
			glog.Infof("%s (item=%d, obj=%d)", jsonCmde.Command, jsonCmde.Itemid, jsonCmde.Objectid)
		}
		w.Write(fctApiTriggerActor(profil, userObj.getId(), jsonCmde))
		return

	default:
		writeApiError(w, fmt.Sprintf("Unhandle command '%s' in (%s)", jsonCmde.Command, r.Form))
		return
	}

//	return

}

// -----------------------------------------------

// startHTTPS is used to start the HTTPS server part.
// It takes a chanel as parameter used to signal error (which is suppose to cause goHome exit after cleanup)
// PI : Simple cert generation for TLS
//    $ openssl req -x509 -nodes -newkey rsa:2048 -keyout certificats/server.key.pem -out certificats/server.crt.pem -days 3650
//    or
//    $ go run $GOROOT/src/crypto/tls/generate_cert.go --host localhost # will generate key.pem and cert.pem for you.
func startHTTPS(chanExit chan bool) {

	db, err := openDB()
	if err != nil {
		glog.Errorf("Error in startHTTPS ... exiting", err)
		chanExit <- true
		return
	}
	// defer db.Close() don't use that, if ListenAndServeTLS run fine, it wont return !

	//-----------------------------
	// Read global param from DB

	fileServerRoot, err := getGlobalParam(db, "Http", "fileserver_root")
	if err != nil {
		glog.Errorf("Error in startHTTPS ... exiting : %s", err)
		chanExit <- true
		return
	}
	if glog.V(1) {
		glog.Infof("FileServer root dir = '%s'", fileServerRoot)
	}

	strPortNum, err := getGlobalParam(db, "Http", "https_port")
	if err != nil {
		glog.Errorf("Error in startHTTPS ... exiting : %s", err)
		chanExit <- true
		return
	}

	strPortNumSimple, err := getGlobalParam(db, "Http", "simple_port")
	if err != nil {
		glog.Infof("Cant get simple_port, not starting simple server : %s", err)
		strPortNumSimple = ""
	}

	serverCrtFileName, err := getGlobalParam(db, "Http", "server_crt")
	if err != nil {
		glog.Errorf("Error in startHTTPS ... exiting : %s", err)
		chanExit <- true
		return
	}
	serverKeyFileName, err := getGlobalParam(db, "Http", "server_key")
	if err != nil {
		glog.Errorf("Error in startHTTPS ... exiting : %s", err)
		chanExit <- true
		return
	}
	caCertFileName, err := getGlobalParam(db, "Http", "ca_crt")
	if err != nil {
		glog.Errorf("Error in startHTTPS ... exiting : %s", err)
		chanExit <- true
		return
	}

	//-----------------------------

	serverMux := http.NewServeMux() // Create dedicated ServeMux, rather than using http.defaultServeMux

	//-----------------------------

	proxyMap, err := getGlobalParamList(db, "Proxy")
	if err != nil {
		glog.Errorf("Error in startHTTPS ... exiting : %s", err)
		chanExit <- true
		return
	}
	for source, target := range proxyMap {
		urlTarget, err := url.Parse(target)
		if err != nil {
			glog.Errorf("startHTTPS fail to parse proxy url target(%s) : %s", target, err)
		}
		if glog.V(1) {
			glog.Infof("Adding proxy (from=%s, to=%s)", source, target)
		}

		serverMux.Handle(source, http.StripPrefix(source, httputil.NewSingleHostReverseProxy(urlTarget)))
	}

	// Note : access to "/api", apiHandler required a registered user in DB
	serverMux.HandleFunc("/api", apiHandler)
	//serverMux.HandleFunc("/tst/", defaultResponse)

	serverMux.Handle("/", http.FileServer(http.Dir(fileServerRoot)))

	//-----------------------------

	db.Close()

	//-----------------------------

	caCert, err := ioutil.ReadFile(caCertFileName)
	if err != nil {
		glog.Errorf("Error reading CA cert (%s)  ... exiting : %s", caCertFileName, err)
		chanExit <- true
		return
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Setup HTTPS client
	tlsConfig := &tls.Config{
		ClientCAs: caCertPool,
		// NoClientCert
		// RequestClientCert
		// RequireAnyClientCert
		// VerifyClientCertIfGiven
		// RequireAndVerifyClientCert
		ClientAuth: tls.RequireAndVerifyClientCert,
	}
	tlsConfig.BuildNameToCertificate()

	server := &http.Server{
		Addr:      ":" + strPortNum,
		Handler:   serverMux,
		TLSConfig: tlsConfig,
		// TODO : use custom logger => ErrorLog: goHomeHttpLogger,
	}


	if ( strPortNumSimple != "" ) {

		serverMuxSimple := http.NewServeMux() // Create dedicated ServeMux, rather than using http.defaultServeMux
		serverMuxSimple.HandleFunc("/simple", simpleResponse)
		serverMuxSimple.HandleFunc("/api", apiHandler)
		serverMuxSimple.Handle("/", http.FileServer(http.Dir(fileServerRoot)))

		// Setup HTTPS client for "simple Client"
		tlsConfigSimple := &tls.Config{
			ClientCAs: caCertPool,
			// NoClientCert
			// RequestClientCert
			// RequireAnyClientCert
			// VerifyClientCertIfGiven
			// RequireAndVerifyClientCert
			ClientAuth: tls.VerifyClientCertIfGiven,
		}
		tlsConfigSimple.BuildNameToCertificate()

		serverSimple := &http.Server{
			Addr:      ":" + strPortNumSimple,
			Handler:   serverMuxSimple,
			TLSConfig: tlsConfigSimple,
			// TODO : use custom logger => ErrorLog: goHomeHttpLogger,
		}

		if glog.V(1) {
			glog.Infof("Starting ListenAndServeTLS (https://*:%s)", strPortNumSimple)
		}

		go serverSimple.ListenAndServeTLS(serverCrtFileName, serverKeyFileName)
	}

	if glog.V(1) {
		glog.Infof("Starting ListenAndServeTLS (https://*:%s)", strPortNum)
	}

	if err = server.ListenAndServeTLS(serverCrtFileName, serverKeyFileName); err != nil {
		glog.Errorf("Error starting HTTPS ListenAndServeTLS : %s ... exiting", err)
		chanExit <- true
	}

}
