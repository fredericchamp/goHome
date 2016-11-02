// main.go
package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/golang/glog"
)

const header = `<!-- HEADER -->
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>goHome</title>
</head>
<body>
`

const footer = `<!-- FOOTER -->
<br><br>
<p align="center">-*-</p>
<p align="center">%s</p>
</body>
</html>
`

const (
	paramNameCmd   = "command"
	paramNameElem  = "element"
	paramNameKeyT  = "keytype"
	paramNameKeyV  = "keyval"
	paramNameJsonP = "jsonparam"
	paramNameObjId = "objectid"
	paramNameStart = "startts"
	paramNameEnd   = "endts"
)

const (
	apiReadCommand      = "Read"
	apiSaveCommand      = "Save"
	apiDeleteCommand    = "Delete"
	apiActionCommand    = "Action"
	apiSensorValCommand = "Sensor Val"
)

const (
	elemItemType    = "itemType"
	elemItem        = "item"
	elemObject      = "object"
	elemSensorValue = "sensorValue"
	elemActorResult = "actorResult"
)

const (
	keyItemType = "itemType"
	keyItem     = "item"
	keyObject   = "object"
)

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

func getFormIntVal(form url.Values, key string, idx int) (intVal int, err error) {
	strVal, err := getFormStrVal(form, key, idx)
	if err != nil {
		return
	}
	intVal, err = strconv.Atoi(strVal)
	if err != nil {
		err = errors.New(fmt.Sprintf(`{"error":"Fail to parse %s id from (%s)"}`, key, strVal))
		return
	}
	return
}

// apiReadHistoValues
func apiReadHistoValues(w http.ResponseWriter, r *http.Request, element string, userObj HomeObject, profil userProfil) (err error) {
	objectId, err := getFormIntVal(r.Form, paramNameObjId, 0)
	if err != nil {
		fmt.Fprintf(w, `{"error":"%s"}`, err.Error())
		return
	}
	startTS, err := getFormIntVal(r.Form, paramNameStart, 0)
	if err != nil {
		return
	}
	endTS, err := getFormIntVal(r.Form, paramNameEnd, 0)
	if err != nil {
		return
	}

	last := true // TODO

	// TODO : check user has access to objectId given his profil

	var sVals []HistoSensor
	var aVals []HistoActor
	switch element {
	case elemSensorValue:
		sVals, err = getHistoSensor(nil, objectId, last, time.Unix(int64(startTS), 0), time.Unix(int64(endTS), 0))
	case elemActorResult:
		aVals, err = getHistActor(nil, objectId, last, time.Unix(int64(startTS), 0), time.Unix(int64(endTS), 0))
	default:
		defaultResponse(w, r, userObj, profil, err)
		return
	}
	if err != nil {
		fmt.Fprintf(w, `{"error":"apiReadHistoValues failed for (%s)"}`, r.Form)
		return
	}

	if glog.V(1) {
		glog.Infof("Read histo %s (%d, %d, %d)", element, objectId, startTS, endTS)
	}

	var jsonEncoded []byte
	switch element {
	case elemSensorValue:
		jsonEncoded, err = json.Marshal(sVals)
	case elemActorResult:
		jsonEncoded, err = json.Marshal(aVals)
	}
	if err != nil {
		fmt.Fprint(w, `{"error":"json.Marshal(_vals) failed"}`)
		return
	}

	w.Write(jsonEncoded)
	return
}

func getPostId(keytype string, id int) (itemTypeId itemType, itemId int, objectId int) {
	switch keytype {
	case keyItemType:
		itemTypeId = itemType(id)
	case keyItem:
		itemId = id
	case keyObject:
		objectId = id
	default:
	}
	return
}

// apiReadObjects
func apiReadObjects(w http.ResponseWriter, r *http.Request, userObj HomeObject, profil userProfil) (err error) {
	keytype, err := getFormStrVal(r.Form, paramNameKeyT, 0)
	if err != nil {
		fmt.Fprintf(w, `{"error":"%s"}`, err.Error())
		return
	}

	keyval, err := getFormIntVal(r.Form, paramNameKeyV, 0)
	if err != nil {
		return
	}
	itemTypeId, itemId, objectId := getPostId(keytype, keyval)

	objs, err := getHomeObjects(nil, objectId, itemId, itemTypeId)
	if err != nil {
		fmt.Fprintf(w, `{"error":"apiReadObjects failed for (%s)"}`, r.Form)
		return
	}
	if glog.V(1) {
		glog.Infof("Read objects (%d, %d, %d)", objectId, itemId, itemTypeId)
	}

	jsonEncoded, err := json.Marshal(profilFilteredObjects(profil, objs))
	if err != nil {
		fmt.Fprint(w, `{"error":"json.Marshal(objs) failed"}`)
		return
	}

	w.Write(jsonEncoded)
	return
}

// apiReadItems
func apiReadItems(w http.ResponseWriter, r *http.Request, userObj HomeObject, profil userProfil) (err error) {
	keytype, err := getFormStrVal(r.Form, paramNameKeyT, 0)
	if err != nil {
		fmt.Fprintf(w, `{"error":"%s"}`, err.Error())
		return
	}

	keyval, err := getFormIntVal(r.Form, paramNameKeyV, 0)
	if err != nil {
		return
	}
	itemTypeId, itemId, _ := getPostId(keytype, keyval)

	items, err := getManageItems(nil, itemId, itemTypeId)
	if err != nil {
		fmt.Fprintf(w, `{"error":"apiReadItems failed for (%s)"}`, r.Form)
		return
	}
	if glog.V(1) {
		glog.Infof("Read managed items (%d, %d)", itemId, itemTypeId)
	}

	jsonEncoded, err := json.Marshal(profilFilteredItems(profil, items))
	if err != nil {
		fmt.Fprint(w, `{"error":"json.Marshal(items) failed"}`)
		return
	}

	w.Write(jsonEncoded)
	return
}

// apiRead
func apiRead(w http.ResponseWriter, r *http.Request, userObj HomeObject, profil userProfil) {

	element, err := getFormStrVal(r.Form, paramNameElem, 0)
	if err != nil {
		fmt.Fprintf(w, `{"error":"%s"}`, err.Error())
		return
	}

	switch element {
	case elemItemType:
		fmt.Fprintf(w, `{"ItemEntity":1,"ItemSensor":2,"ItemActor":3,"ItemSensorAct":4,"ItemStreamSensor":5}`)
		if glog.V(1) {
			glog.Info("Read itemTypes")
		}
	case elemItem:
		apiReadItems(w, r, userObj, profil)
	case elemObject:
		apiReadObjects(w, r, userObj, profil)
	case elemSensorValue:
	case elemActorResult:
		apiReadHistoValues(w, r, element, userObj, profil)
	default:
		defaultResponse(w, r, userObj, profil, err)
		return
	}
}

// apiSave
func apiSave(w http.ResponseWriter, r *http.Request, userObj HomeObject, profil userProfil) {

	element, err := getFormStrVal(r.Form, paramNameElem, 0)
	if err != nil {
		fmt.Fprintf(w, `{"error":"%s"}`, err.Error())
		return
	}

	switch element {
	case elemItemType:
		fmt.Fprintf(w, `{"error":"apiSave : can not save '%s'"}`, element)
		return
	case elemItem: // TODO save
		// Unmarchal jsonparam to Item
		// check item content
		// check profil rights on item
		// save item to DB (update or insert)
	case elemObject: // TODO save
		// Unmarchal jsonparam to []ItemFieldVal
		// Identify corresponding item
		// check profil rights on item
		// Build HomeObject and check []ItemFieldVal regarding item description
		// If id is provided :
		//		check itemId in DB match itemId deduce from jsonparam content
		//		check profil rights on existing object
		//		check profil rights on update object
		//		update db
		// Else
		//		check profil rights on new object
		//		insert into DB with new Id allocation
	case elemSensorValue: // TODO save
		// Unmarchal jsonparam to SensorVal
		// load sensor using getHomeObjects
		// add sensor.linkedObjs if not loaded by getHomeObjects
		// Call handleSensorValue(time.Now(), sensor, value)
	default:
		defaultResponse(w, r, userObj, profil, err)
		return
	}
}

// apiDelete
func apiDelete(w http.ResponseWriter, r *http.Request, userObj HomeObject, profil userProfil) {

	objectId, err := getFormIntVal(r.Form, paramNameObjId, 0)
	if err != nil {
		fmt.Fprintf(w, `{"error":"%s"}`, err.Error())
		return
	}

	objs, err := getHomeObjects(nil, objectId, 0, ItemNone)
	if err != nil {
		fmt.Fprintf(w, `{"error":"apiDelete Object for id=%d : not found"}`, objectId)
		return
	}

	objs = profilFilteredObjects(profil, objs)
	if len(objs) <= 0 {
		fmt.Fprintf(w, `{"error":"apiDelete Object for id=%d : insufficient privileges"}`, objectId)
		return
	}

	// TODO delete object objectId from DB

}

// -----------------------------------------------

// defaultResponse : when a request is not properly form or handle
func defaultResponse(w http.ResponseWriter, r *http.Request, userObj HomeObject, profil userProfil, err error) {
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	fmt.Fprintf(w, header)
	fmt.Fprintf(w, "<p>goHome HTTPS server<p>\n")
	fmt.Fprintf(w, "<p>goHome version %s</p>\n", goHomeVersion)
	fmt.Fprintf(w, "<p>URL requested : %s </p>\n", r.URL.Path)
	fmt.Fprintf(w, "<p>Post params : %s</p>\n", r.Form)
	fmt.Fprintf(w, "<p>User : %s</p>\n", userObj)

	fmt.Fprintf(w, footer, time.Now().String())
}

// -----------------------------------------------
// URL handlers

// TODO proxy (video stream)

// apiHandler : request API requests
func apiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")

	userObj, err := getUserFromCert(r.TLS.PeerCertificates)
	if err != nil {
		fmt.Fprintf(w, `{"error":"%s"}`, err.Error())
		return
	}

	profil, err := checkApiUser(userObj)
	if err != nil {
		fmt.Fprintf(w, `{"error":"%s"}`, err.Error())
		return
	}

	err = r.ParseForm()
	if err != nil {
		fmt.Fprintf(w, `{"error":"Form parse error '%s' for (%s)"}`, err.Error(), r.Form)
		return
	}
	if glog.V(2) {
		glog.Info(r.Form)
	}

	if len(r.Form) < 5 {
		fmt.Fprintf(w, `{"error":"Missing parameters in (%s)"}`, r.Form)
		return
	}

	command, err := getFormStrVal(r.Form, paramNameCmd, 0)
	if err != nil {
		fmt.Fprintf(w, `{"error":"Command not found in (%s)"}`, r.Form)
		return
	}

	switch command {
	case apiReadCommand:
		apiRead(w, r, userObj, profil)
	case apiSaveCommand:
		apiSave(w, r, userObj, profil)
	case apiDeleteCommand:
		apiDelete(w, r, userObj, profil)
	case apiSensorValCommand:
	case apiActionCommand:
	default:
		defaultResponse(w, r, userObj, profil, errors.New(fmt.Sprintf(`{"error":"Unhandle command '%s' in (%s)"}`, command, r.Form)))
		return
	}
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
	// defer db.Close() dont do that, if ListenAndServeTLS run fine, it wont return !

	//-----------------------------
	// Read global param from DB

	fileServerRoot, err := getGlobalParam(db, -1, "goHome", "fileserver_root")
	if err != nil {
		chanExit <- true
		return
	}
	if glog.V(1) {
		glog.Infof("FileServer root dir = '%s'", fileServerRoot)
	}

	serverName, err := getGlobalParam(db, -1, "goHome", "server_name")
	if err != nil {
		glog.Errorf("Error in startHTTPS ... exiting", err)
		chanExit <- true
		return
	}
	value, err := getGlobalParam(db, -1, "goHome", "https_port")
	if err != nil {
		glog.Errorf("Error in startHTTPS ... exiting", err)
		chanExit <- true
		return
	}
	port, err := strconv.Atoi(value)
	if err != nil {
		glog.Errorf("Error converting port# (%s) ... exiting : %s", value, err)
		chanExit <- true
		return
	}

	serverCrtFileName, err := getGlobalParam(db, -1, "goHome", "server_crt")
	if err != nil {
		glog.Errorf("Error in startHTTPS ... exiting", err)
		chanExit <- true
		return
	}
	serverKeyFileName, err := getGlobalParam(db, -1, "goHome", "server_key")
	if err != nil {
		glog.Errorf("Error in startHTTPS ... exiting", err)
		chanExit <- true
		return
	}
	caCertFileName, err := getGlobalParam(db, -1, "goHome", "ca_crt")
	if err != nil {
		glog.Errorf("Error in startHTTPS ... exiting", err)
		chanExit <- true
		return
	}

	//-----------------------------

	db.Close()

	//-----------------------------

	serverMux := http.NewServeMux() // Create dedicated ServeMux, rather than using http.defaultServeMux
	// Note : access to "/api", apiHandler required a registered user in DB
	serverMux.HandleFunc("/api", apiHandler)
	// Note : access to "/", FileServer only required a valid certificat (valid regarding caCertPool)
	serverMux.Handle("/", http.FileServer(http.Dir(fileServerRoot)))

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
		Addr:      fmt.Sprintf("%s:%d", serverName, port),
		Handler:   serverMux,
		TLSConfig: tlsConfig,
		// TODO : use custom logger => ErrorLog: goHomeHttpLogger,
	}

	if glog.V(1) {
		glog.Infof("Starting ListenAndServeTLS (https://%s:%d)", serverName, port)
	}

	err = server.ListenAndServeTLS(serverCrtFileName, serverKeyFileName)
	if err != nil {
		glog.Errorf("Error starting HTTPS ListenAndServeTLS : %s ... exiting", err)
		chanExit <- true
	}
}

// -----------------------------------------------
// -----------------------------------------------

// -----------------------------------------------
// Utilities

// -----------------------------------------------

func callActor() {
	// TODO callActor( name | id ) + param
}

func setSensorVal() {
	// TODO setSensorVal( name | id ) + val
}

// -----------------------------------------------

func writeObject() {

	// TODO writeObject if it's a user => loadUsers(true) when finish
}

func deleteObject() {

	// TODO deleteObject if it's a user => loadUsers(true) when finish
}
