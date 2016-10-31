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
	"strconv"
	"strings"
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

// -----------------------------------------------

// startHTTPS is used to start the HTTPS server part.
// It takes an int as parameter that define the port number to listen
// PI : Simple cert generation
//    $ openssl req -x509 -nodes -newkey rsa:2048 -keyout certificats/server.key.pem -out certificats/server.crt.pem -days 3650
//    or
//    $ go run $GOROOT/src/crypto/tls/generate_cert.go --host localhost # will generate key.pem and cert.pem for you.
func startHTTPS(chanExit chan bool) {

	db, err := openDB()
	if err != nil {
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
		chanExit <- true
		return
	}
	value, err := getGlobalParam(db, -1, "goHome", "https_port")
	if err != nil {
		chanExit <- true
		return
	}
	port, err := strconv.Atoi(value)
	if err != nil {
		glog.Errorf("Error converting port# (%s) : %s", value, err)
		chanExit <- true
		return
	}

	serverCrtFileName, err := getGlobalParam(db, -1, "goHome", "server_crt")
	if err != nil {
		chanExit <- true
		return
	}
	serverKeyFileName, err := getGlobalParam(db, -1, "goHome", "server_key")
	if err != nil {
		chanExit <- true
		return
	}
	caCertFileName, err := getGlobalParam(db, -1, "goHome", "ca_crt")
	if err != nil {
		chanExit <- true
		return
	}

	//-----------------------------

	db.Close()

	//-----------------------------

	serverMux := http.NewServeMux() // rather than using http.defaultServeMux
	serverMux.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(fileServerRoot))))
	serverMux.HandleFunc("/", httpsHandler)

	//-----------------------------

	caCert, err := ioutil.ReadFile(caCertFileName)
	if err != nil {
		glog.Errorf("Error reading CA cert (%s) : %s", caCertFileName, err)
		chanExit <- true
		return
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	// TODO accept other CA

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

// file mgt : upload/download : icones, cert, ...
// proxy (video stream)

// -----------------------------------------------
// URL handlers

// itemTypeListHandler : Handle HTTPS request to '/itemTypes'
func itemTypeListHandler(w http.ResponseWriter, r *http.Request, postParam map[string]string, email string, profil userProfil) {
	fmt.Fprintf(w, `{"ItemEntity":1,"ItemSensor":2,"ItemActor":3,"ItemSensorAct":4,"ItemStreamSensor":5}`)
}

// itemListHandler : Handle HTTPS request to '/items'
func itemListHandler(w http.ResponseWriter, r *http.Request, postParam map[string]string, email string, profil userProfil) {
	// requested url can be :
	// "/items"					=> return all existing items
	// "/items/itemTypeId/nn"	=> return items with idItemType = nn
	// "/items/itemId/nn"	=> return items with id = nn
	var items []Item
	var err error
	switch { // TODO use postParam
	case strings.Contains(r.URL.Path, "/itemTypeId/"):
		urlPart := strings.Split(r.URL.Path, "/")
		itemTypeId, err := strconv.Atoi(urlPart[len(urlPart)-1])
		if err != nil {
			fmt.Fprintf(w, `{"error":"Fail to parse itemTypeId from (%s)"}`, r.URL.Path)
			return
		}
		items, err = getManageItems(nil, -1, itemType(itemTypeId))

	case strings.Contains(r.URL.Path, "/itemId/"):
		urlPart := strings.Split(r.URL.Path, "/")
		itemId, err := strconv.Atoi(urlPart[len(urlPart)-1])
		if err != nil {
			fmt.Fprintf(w, `{"error":"Fail to parse itemId from (%s)"}`, r.URL.Path)
			return
		}
		items, err = getManageItems(nil, itemId, ItemNone)

	default:
		items, err = getManageItems(nil, -1, ItemNone)

	}
	if err != nil {
		fmt.Fprintf(w, `{"error":"itemListHandler failed for (%s)"}`, r.URL.Path)
		return
	}

	jsonEncoded, err := json.Marshal(profilFilteredItems(profil, items))
	if err != nil {
		fmt.Fprintf(w, `{"error":"json.Marshal(items) failed"}`, email, profil)
		return
	}

	w.Write(jsonEncoded)

}

// objectListHandler : Handle HTTPS request to '/objects'
func objectListHandler(w http.ResponseWriter, r *http.Request, postParam map[string]string, email string, profil userProfil) {
	// requested url can be :
	// "/objects"				=> unused, return error
	// "/objects/itemTypeId/nn"	=> return all objects for items with idItemType = nn
	// "/objects/itemId/nn"		=> return all objects for idItem = nn
	// "/objects/objectId/nn"	=> return object for id = nn
	var objs []HomeObject
	var err error
	switch { // TODO use postParam + check : r.URL.Path[len(r.URL.Path)-1]
	case strings.Contains(r.URL.Path, "/itemTypeId/"):
		urlPart := strings.Split(r.URL.Path, "/")
		itemTypeId, err := strconv.Atoi(urlPart[len(urlPart)-1])
		if err != nil {
			fmt.Fprintf(w, `{"error":"Fail to parse itemTypeId from (%s)"}`, r.URL.Path)
			return
		}
		objs, err = getHomeObjectsForType(nil, itemType(itemTypeId))

	case strings.Contains(r.URL.Path, "/itemId/"):
		urlPart := strings.Split(r.URL.Path, "/")
		itemId, err := strconv.Atoi(urlPart[len(urlPart)-1])
		if err != nil {
			fmt.Fprintf(w, `{"error":"Fail to parse itemId from (%s)"}`, r.URL.Path)
			return
		}
		objs, err = getHomeObjects(nil, -1, itemId)

	case strings.Contains(r.URL.Path, "/objectId/"):
		urlPart := strings.Split(r.URL.Path, "/")
		objId, err := strconv.Atoi(urlPart[len(urlPart)-1])
		if err != nil {
			fmt.Fprintf(w, `{"error":"Fail to parse objectId from (%s)"}`, r.URL.Path)
			return
		}
		objs, err = getHomeObjects(nil, objId, -1)

	default:
		err = errors.New("dummy")
	}
	if err != nil {
		fmt.Fprintf(w, `{"error":"objectListHandler failed for (%s)"}`, r.URL.Path)
		return
	}

	jsonEncoded, err := json.Marshal(profilFilteredObjects(profil, objs))
	if err != nil {
		fmt.Fprintf(w, `{"error":"json.Marshal(objs) failed"}`)
		return
	}

	w.Write(jsonEncoded)

}

// httpsHandler : Handle HTTPS request
func httpsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")

	email, profil := getUserEmailAndProfil(r.TLS.PeerCertificates)

	// TODO retreive postParam
	postParam := map[string]string{}

	switch {
	case profil > ProfilNone && strings.HasPrefix(r.URL.Path, "/itemTypes"):
		itemTypeListHandler(w, r, postParam, email, profil)
	case profil > ProfilNone && strings.HasPrefix(r.URL.Path, "/items"):
		itemListHandler(w, r, postParam, email, profil)
	case profil > ProfilNone && strings.HasPrefix(r.URL.Path, "/objects"):
		objectListHandler(w, r, postParam, email, profil)
	case profil > ProfilNone && strings.HasPrefix(r.URL.Path, "/histoSensor"):
		// TODO histoSensorHandler(w, r, postParam, email, profil)
	case profil > ProfilNone && strings.HasPrefix(r.URL.Path, "/histoActor"):
		// TODO histoActorHandler(w, r, postParam, email, profil)
	default:
		fmt.Fprintf(w, header)
		fmt.Fprintf(w, "<p>goHome HTTPS server<p>\n")
		fmt.Fprintf(w, "<p>goHome version %s</p>\n", goHomeVersion)
		fmt.Fprintf(w, "<p>URL requested : %s (no dedidated handler)</p>\n", r.URL.Path)
		if profil <= ProfilNone {
			fmt.Fprintf(w, "<p>Unknown client or insufficient privileges (email='%s', profil=%d)</p>", email, profil)
		} else {
			fmt.Fprintf(w, "<p>Client with email from cert '%s' found => profil=%d</p>\n", email, profil)
		}
		fmt.Fprintf(w, footer, time.Now().String())
	}
}

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

	// TODO if it's a user => loadUsers(true) when finish
}
