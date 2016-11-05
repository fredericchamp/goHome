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

	"github.com/golang/glog"
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

func writeApiError(w http.ResponseWriter, errMsg string) {
	w.Write(apiError(errMsg))
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
//		writeJsonError(w, err.Error())
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
		writeApiError(w, err.Error())
		return
	}

	profil, err := checkApiUser(userObj)
	if err != nil {
		writeApiError(w, err.Error())
		return
	}

	if err = r.ParseForm(); err != nil {
		writeApiError(w, fmt.Sprintf("Form parse error '%s' for (%s)", err, r.Form))
		return
	}
	if glog.V(2) {
		glog.Infof("User profil %d, Form=", profil, r.Form)
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
		/*
			items, err := getManageItems(nil, jsonCmde.Itemtypeid, jsonCmde.Itemid)
			if err != nil {
				writeJsonError(w, fmt.Sprintf("%s failed for (type=%d, item=%d) : %s", jsonCmde.Command, jsonCmde.Itemtypeid, jsonCmde.Itemid, err))
				return
			}

			jsonEncoded, err := json.Marshal(profilFilteredItems(profil, items))
			if err != nil {
				writeJsonError(w, fmt.Sprintf("%s failed for (type=%d, item=%d) : %s", jsonCmde.Command, jsonCmde.Itemtypeid, jsonCmde.Itemid, err))
				return
			}
			w.Write(jsonEncoded)
		*/
		w.Write(fctApiReadItem(profil, jsonCmde))
		return

	case apiReadObject:
		if glog.V(1) {
			glog.Infof("%s (type=%d, item=%d, obj=%d)", jsonCmde.Command, jsonCmde.Itemtypeid, jsonCmde.Itemid, jsonCmde.Objectid)
		}
		/*
			objs, err := getHomeObjects(nil, jsonCmde.Itemtypeid, jsonCmde.Itemid, jsonCmde.Objectid)
			if err != nil {
				writeJsonError(w, fmt.Sprintf("%s failed for (%s) : %s", jsonCmde.Command, r.Form, err))
				return
			}

			jsonEncoded, err := json.Marshal(profilFilteredObjects(profil, objs))
			if err != nil {
				writeJsonError(w, fmt.Sprintf("%s failed for (%s) : %s", jsonCmde.Command, r.Form, err))
				return
			}

			w.Write(jsonEncoded)
		*/
		w.Write(fctApiReadObject(profil, jsonCmde))
		return

	case apiReadSensor:
		if glog.V(1) {
			glog.Infof("%s (objectid=%d)", jsonCmde.Command, jsonCmde.Objectid)
		}
		/*
			objs, err := getHomeObjects(nil, ItemNone, -1, jsonCmde.Objectid)
			if err != nil {
				writeJsonError(w, fmt.Sprintf("%s failed for (%s) : %s", jsonCmde.Command, r.Form, err))
				return
			}
			if len(objs) <= 0 {
				writeJsonError(w, fmt.Sprintf("%s failed for (%s) : object not found", jsonCmde.Command, r.Form))
				return
			}
			sensor := objs[0]

			err = checkAccessToObject(profil, sensor)
			if err != nil {
				writeJsonError(w, err.Error())
				return
			}

			value, err := readSensoValue(sensor)
			if err != nil {
				writeJsonError(w, fmt.Sprintf("%s failed for (%s) : %s", jsonCmde.Command, r.Form, err))
				return
			}

			jsonEncoded, err := json.Marshal(value)
			if err != nil {
				writeJsonError(w, fmt.Sprintf("%s failed for (%s) : %s", jsonCmde.Command, r.Form, err))
				return
			}

			w.Write(jsonEncoded)
		*/
		w.Write(fctApiReadSensor(profil, jsonCmde))
		return

	case apiReadHistoVal:
		if glog.V(1) {
			glog.Infof("%s (obj=%d, start=%d, end=%d)", jsonCmde.Command, jsonCmde.Objectid, jsonCmde.Startts, jsonCmde.Endts)
		}
		/*
			last := false
			if (jsonCmde.Startts <= 0 && jsonCmde.Endts <= 0) || jsonCmde.Startts > time.Now().Unix() {
				last = true
			}

			err = checkAccessToObjectId(profil, jsonCmde.Objectid)
			if err != nil {
				writeJsonError(w, err.Error())
				return
			}

			sVals, err := getHistoSensor(nil, jsonCmde.Objectid, last, time.Unix(jsonCmde.Startts, 0), time.Unix(jsonCmde.Endts, 0))
			if err != nil {
				writeJsonError(w, fmt.Sprintf("%s failed for (%s) : %s", jsonCmde.Command, r.Form, err))
				return
			}

			jsonEncoded, err := json.Marshal(sVals)
			if err != nil {
				writeJsonError(w, fmt.Sprintf("%s failed for (%s) : %s", jsonCmde.Command, r.Form, err))
				return
			}

			w.Write(jsonEncoded)
		*/
		w.Write(fctApiReadHistoVal(profil, jsonCmde))
		return

	case apiReadActorRes:
		if glog.V(1) {
			glog.Infof("%s (objectid=%d, start=%d, end=%d)", jsonCmde.Command, jsonCmde.Objectid, jsonCmde.Startts, jsonCmde.Endts)
		}
		/*
			last := false
			if (jsonCmde.Startts <= 0 && jsonCmde.Endts <= 0) || jsonCmde.Startts > time.Now().Unix() {
				last = true
			}

			err = checkAccessToObjectId(profil, jsonCmde.Objectid)
			if err != nil {
				writeJsonError(w, err.Error())
				return
			}

			aVals, err := getHistActor(nil, jsonCmde.Objectid, last, time.Unix(jsonCmde.Startts, 0), time.Unix(jsonCmde.Endts, 0))
			if err != nil {
				writeJsonError(w, fmt.Sprintf("%s failed for (%s) : %s", jsonCmde.Command, r.Form, err))
				return
			}

			jsonEncoded, err := json.Marshal(aVals)
			if err != nil {
				writeJsonError(w, fmt.Sprintf("%s failed for (%s) : %s", jsonCmde.Command, r.Form, err))
				return
			}

			w.Write(jsonEncoded)
		*/
		w.Write(fctApiReadActorRes(profil, jsonCmde))
		return

	case apiSaveItem: // TODO
		// Unmarchal jsonparam to Item
		// check item content
		// check profil rights on item
		// save item to DB (update or insert)
		writeApiError(w, fmt.Sprintf("Command %s not ready", jsonCmde.Command))
		return

	case apiSaveObject:
		if glog.V(1) {
			glog.Infof("%s (type=%d, item=%d, obj=%d)", jsonCmde.Command, jsonCmde.Itemtypeid, jsonCmde.Itemid, jsonCmde.Objectid)
		}
		/*
			var objIn HomeObject
			if err = json.Unmarshal([]byte(jsonCmde.Jsonparam), &objIn); err != nil {
				writeJsonError(w, fmt.Sprintf("%s fail to unmarshal jsonparam (%s) : %s", jsonCmde.Command, jsonCmde.Jsonparam, err))
				return
			}

			// load object from db
			objs, err := getHomeObjects(nil, jsonCmde.Itemtypeid, jsonCmde.Itemid, jsonCmde.Objectid)
			if err != nil {
				writeJsonError(w, fmt.Sprintf("%s fail to load matching object (%s) : %s", jsonCmde.Command, jsonCmde.Objectid, err))
				return
			}
			if len(objs) != 1 {
				writeJsonError(w, fmt.Sprintf("%s should have only 1 matching object, not %d", jsonCmde.Command, len(objs)))
				return
			}
			objDb := objs[0]

			// check profil access rights on item
			if err = checkAccessToObjectId(profil, objDb.Fields[0].IdItem); err != nil {
				writeJsonError(w, err.Error())
				return
			}

			// Check objIn fieldsand objDb fields are the same
			if !reflect.DeepEqual(objIn.Fields, objDb.Fields) {
				writeJsonError(w, fmt.Sprintf("%s received []Fields does not match []Fields in DB for itemid=%d", jsonCmde.Command, jsonCmde.Itemid))
				return
			}

			if err = objIn.ValidateValues(objIn.Values); err != nil {
				writeJsonError(w, fmt.Sprintf("%s : %s", jsonCmde.Command, err.Error()))
				return
			}

			if jsonCmde.Objectid > 0 {
				// check profil access rights on existing object
				if err = checkAccessToObject(profil, objDb); err != nil {
					writeJsonError(w, err.Error())
					return
				}
			}

			// check profil access rights on new object
			if err = checkAccessToObject(profil, objIn); err != nil {
				writeJsonError(w, err.Error())
				return
			}

			// write object to DB
			if jsonCmde.Objectid, err = writeObject(objIn); err != nil {
				writeJsonError(w, err.Error())
				return
			}

			// Code copy from apiReadObject [-----------------------------
			objs, err = getHomeObjects(nil, ItemNone, -1, jsonCmde.Objectid)
			if err != nil {
				writeJsonError(w, fmt.Sprintf("%s failed for (objectid=%d) : %s", jsonCmde.Command, jsonCmde.Objectid, err))
				return
			}

			jsonEncoded, err := json.Marshal(profilFilteredObjects(profil, objs))
			if err != nil {
				writeJsonError(w, fmt.Sprintf("%s json.Marshal failed for (%s) : %s", jsonCmde.Command, jsonCmde.Objectid, err))
				return
			}

			w.Write(jsonEncoded)
			// ----------------------------------------------------------]

		*/
		w.Write(fctApiSaveObject(profil, jsonCmde))
		return

	case apiDeleteItem, apiDeleteObject:
		writeApiError(w, "Delete not available : use apiSave* to toggle IsActive flag or use manual access to DB")
		return

	case apiSendSensorVal: // TODO
		//		if glog.V(1) {
		//			glog.Infof("%s (type=%d, item=%d, obj=%d)", jsonCmde.Command, jsonCmde.Objectid, jsonCmde.Value)
		//		}

		//		objs, err := getHomeObjects(nil, ItemNone, -1, jsonCmde.Objectid)
		//		if err != nil {
		//			writeJsonError(w, fmt.Sprintf("%s failed for (%s) : %s", jsonCmde.Command, r.Form, err))
		//			return
		//		}

		// Unmarchal jsonparam to SensorVal
		// load sensor using getHomeObjects
		// Call handleSensorValue(ts, sensor, value)
		writeApiError(w, fmt.Sprintf("Command %s not ready", jsonCmde.Command))
		return

	case apiTriggerActor: // TODO
		writeApiError(w, fmt.Sprintf("Command %s not ready", jsonCmde.Command))
		return

	default:
		writeApiError(w, fmt.Sprintf("Unhandle command '%s' in (%s)", jsonCmde.Command, r.Form))
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

	if err = server.ListenAndServeTLS(serverCrtFileName, serverKeyFileName); err != nil {
		glog.Errorf("Error starting HTTPS ListenAndServeTLS : %s ... exiting", err)
		chanExit <- true
	}
}
