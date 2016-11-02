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

// -----------------------------------------------

// Accepted Format = {"command":"api...", "itemtypeid":id, "itemid":id, "objectid":id, "startts":ts, "endts":ts, "jsonparam":{...}}

type apiCommand string

const (
	apiReadItemType  apiCommand = "ReadItemTypes"
	apiReadItem                 = "ReadItems"
	apiReadObject               = "ReadObject"
	apiReadSensorVal            = "ReadSensorVal"
	apiReadActorRes             = "ReadActorRes"
	apiSaveItem                 = "SaveItems"
	apiSaveObject               = "SaveObject"
	apiDeleteItem               = "DeleteItems"
	apiDeleteObject             = "DeleteObject"
	apiSendSensorVal            = "SendSensorVal"
	apiTriggerActor             = "TriggerActor"
)

type apiCommandSruct struct {
	Command    apiCommand
	Itemtypeid itemType
	Itemid     int
	Objectid   int
	Startts    int64
	Endts      int64
	Jsonparam  string
}

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

// -----------------------------------------------

// defaultResponse : when a request is not properly form or handle ... was useful during dev
//func defaultResponse(w http.ResponseWriter, r *http.Request, userObj HomeObject, profil userProfil, err error) {
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
//	if err != nil {
//		fmt.Fprint(w, err.Error())
//		return
//	}
//	fmt.Fprintf(w, header)
//	fmt.Fprintf(w, "<p>goHome HTTPS server<p>\n")
//	fmt.Fprintf(w, "<p>goHome version %s</p>\n", goHomeVersion)
//	fmt.Fprintf(w, "<p>URL requested : %s </p>\n", r.URL.Path)
//	fmt.Fprintf(w, "<p>Post params : %s</p>\n", r.Form)
//	fmt.Fprintf(w, "<p>User : %s</p>\n", userObj)
//	fmt.Fprintf(w, footer, time.Now().String())
//}

// -----------------------------------------------
// URL handlers

// TODO proxy (video stream)

// apiHandler : handle API requests
func apiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")

	userObj, err := getUserFromCert(r.TLS.PeerCertificates)
	if err != nil {
		fmt.Fprintf(w, `{"error":"%s"}`, err)
		return
	}

	profil, err := checkApiUser(userObj)
	if err != nil {
		fmt.Fprintf(w, `{"error":"%s"}`, err)
		return
	}

	err = r.ParseForm()
	if err != nil {
		fmt.Fprintf(w, `{"error":"Form parse error '%s' for (%s)"}`, err, r.Form)
		return
	}
	if glog.V(2) {
		glog.Infof("User profil %d, Form=", profil, r.Form)
	}

	command, err := getFormStrVal(r.Form, "command", 0)
	if err != nil {
		fmt.Fprintf(w, `{"error":"%s"}`, err)
		return
	}

	var jsonCmde apiCommandSruct
	err = json.Unmarshal([]byte(command), &jsonCmde)
	if err != nil {
		glog.Errorf("Fail to unmarshal apiCommandSruct (%s) : %s", command, err)
		return
	}

	if glog.V(2) {
		glog.Info(jsonCmde)
	}

	switch jsonCmde.Command {

	case apiReadItemType:
		if glog.V(1) {
			glog.Info(jsonCmde.Command)
		}
		fmt.Fprint(w, `{"ItemEntity":1,"ItemSensor":2,"ItemActor":3,"ItemSensorAct":4,"ItemStreamSensor":5}`)
		return

	case apiReadItem:
		if glog.V(1) {
			glog.Infof("%s (type=%d, item=%d)", jsonCmde.Command, jsonCmde.Itemtypeid, jsonCmde.Itemid)
		}

		items, err := getManageItems(nil, jsonCmde.Itemtypeid, jsonCmde.Itemid)
		if err != nil {
			fmt.Fprintf(w, `{"error":"%s failed for (%s) : %s"}`, jsonCmde.Command, r.Form, err)
			return
		}

		jsonEncoded, err := json.Marshal(profilFilteredItems(profil, items))
		if err != nil {
			fmt.Fprintf(w, `{"error":"%s failed for (%s) : %s"}`, jsonCmde.Command, r.Form, err)
			return
		}
		w.Write(jsonEncoded)
		return

	case apiReadObject:
		if glog.V(1) {
			glog.Infof("%s (type=%d, item=%d, obj=%d)", jsonCmde.Command, jsonCmde.Itemtypeid, jsonCmde.Itemid, jsonCmde.Objectid)
		}

		objs, err := getHomeObjects(nil, jsonCmde.Itemtypeid, jsonCmde.Itemid, jsonCmde.Objectid)
		if err != nil {
			fmt.Fprintf(w, `{"error":"%s failed for (%s) : %s"}`, jsonCmde.Command, r.Form, err)
			return
		}

		jsonEncoded, err := json.Marshal(profilFilteredObjects(profil, objs))
		if err != nil {
			fmt.Fprintf(w, `{"error":"%s failed for (%s) : %s"}`, jsonCmde.Command, r.Form, err)
			return
		}

		w.Write(jsonEncoded)
		return

	case apiReadSensorVal:
		if glog.V(1) {
			glog.Infof("%s (%d, %d, %d)", jsonCmde.Command, jsonCmde.Objectid, jsonCmde.Startts, jsonCmde.Endts)
		}

		last := false
		if (jsonCmde.Startts <= 0 && jsonCmde.Endts <= 0) || jsonCmde.Startts > time.Now().Unix() {
			last = true
		}

		err = checkAccessToObjectId(profil, jsonCmde.Objectid)
		if err != nil {
			fmt.Fprintf(w, `{"error":"%s"}`, err)
			return
		}

		sVals, err := getHistoSensor(nil, jsonCmde.Objectid, last, time.Unix(jsonCmde.Startts, 0), time.Unix(jsonCmde.Endts, 0))
		if err != nil {
			fmt.Fprintf(w, `{"error":"%s failed for (%s) : %s"}`, jsonCmde.Command, r.Form, err)
			return
		}

		jsonEncoded, err := json.Marshal(sVals)
		if err != nil {
			fmt.Fprintf(w, `{"error":"%s failed for (%s) : %s"}`, jsonCmde.Command, r.Form, err)
			return
		}

		w.Write(jsonEncoded)
		return

	case apiReadActorRes:
		if glog.V(1) {
			glog.Infof("%s (%d, %d, %d)", jsonCmde.Command, jsonCmde.Objectid, jsonCmde.Startts, jsonCmde.Endts)
		}

		last := false
		if (jsonCmde.Startts <= 0 && jsonCmde.Endts <= 0) || jsonCmde.Startts > time.Now().Unix() {
			last = true
		}

		err = checkAccessToObjectId(profil, jsonCmde.Objectid)
		if err != nil {
			fmt.Fprintf(w, `{"error":"%s"}`, err)
			return
		}

		aVals, err := getHistActor(nil, jsonCmde.Objectid, last, time.Unix(jsonCmde.Startts, 0), time.Unix(jsonCmde.Endts, 0))
		if err != nil {
			fmt.Fprintf(w, `{"error":"%s failed for (%s) : %s"}`, jsonCmde.Command, r.Form, err)
			return
		}

		jsonEncoded, err := json.Marshal(aVals)
		if err != nil {
			fmt.Fprintf(w, `{"error":"%s failed for (%s) : %s"}`, jsonCmde.Command, r.Form, err)
			return
		}

		w.Write(jsonEncoded)
		return

	case apiSaveItem: // TODO
		// Unmarchal jsonparam to Item
		// check item content
		// check profil rights on item
		// save item to DB (update or insert)
		fmt.Fprintf(w, `{"error":"Command %s not ready"}`, jsonCmde.Command)
		return

	case apiSaveObject: // TODO
		// Unmarchal jsonparam to []ItemFieldVal
		// Identify corresponding item
		// check profil rights on item
		// Build HomeObject and check []ItemFieldVal regarding item description
		// If id is provided :
		//		check itemId in DB match itemId deduce from jsonparam content
		//		check profil rights on existing object values
		//		check profil rights on new object values
		//		update db
		// Else
		//		check profil rights on new object
		//		insert into DB with new Id allocation
		fmt.Fprintf(w, `{"error":"Command %s not ready"}`, jsonCmde.Command)
		return

	case apiDeleteItem, apiDeleteObject:
		fmt.Fprint(w, `{"error":"Delete not available : use apiSave* to toggle IsActive flag or use manual access to DB"}`)
		return

	case apiSendSensorVal: // TODO
		// Unmarchal jsonparam to SensorVal
		// load sensor using getHomeObjects
		// Call handleSensorValue(ts, sensor, value)
		fmt.Fprintf(w, `{"error":"Command %s not ready"}`, jsonCmde.Command)
		return

	case apiTriggerActor: // TODO
		fmt.Fprintf(w, `{"error":"Command %s not ready"}`, jsonCmde.Command)
		return

	default:
		fmt.Fprintf(w, `{"error":"Unhandle command '%s' in (%s)"}`, jsonCmde.Command, r.Form)
		return
	}

	return

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
	serverMux.HandleFunc("/json", apiHandler)
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
