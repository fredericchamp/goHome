// main.go
package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/golang/glog"
)

const serverName = ""

// rootHTTPHandler : Handle HTTP request to '/'
func rootHTTPHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<p>goHome version %s</p>", goHomeVersion)
	fmt.Fprintf(w, "<p>URL requested : %s</p>", r.URL.Path)
}

// startHTTP is used to start the HTTP server part.
// It takes an int as parameter that define the port number to listen
func startHTTP(port int) {
	http.HandleFunc("/", rootHTTPHandler)
	http.ListenAndServe(fmt.Sprintf("localhost:%d", port), nil) // TODO error handling
	if glog.V(1) {
		glog.Infof("startHTTP(%d) done", port)
	}
}

// -----------------------------------------------

// rootHTTPSHandler : Handle HTTPS request to '/'
func rootHTTPSHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
	fmt.Fprintf(w, "<p>goHome HTTPS server (%s)<p>", time.Now().String())
	fmt.Fprintf(w, "<p>goHome version %s</p>", goHomeVersion)
	fmt.Fprintf(w, "<p>URL requested : %s</p>", r.URL.Path)
}

// startHTTPS is used to start the HTTPS server part.
// It takes an int as parameter that define the port number to listen
// PI : Simple cert generation : openssl req -x509 -nodes -newkey rsa:2048 -keyout certificats/server.key.pem -out certificats/server.crt.pem -days 3650
func startHTTPS(chanExit chan bool) {

	db, err := openDB()
	if err != nil {
		glog.Error(err)
		chanExit <- true
		return
	}
	defer db.Close()

	value, err := getGlobalParam(db, -1, "goHome", "port")
	if err != nil {
		glog.Errorf("Error getting port# : %s", err)
		chanExit <- true
		return
	}
	port, err := strconv.Atoi(value)
	if err != nil {
		glog.Errorf("Error converting port# (%s) : %s", value, err)
		chanExit <- true
		return
	}

	// TODO read certFileNames from DB
	var caCertFile string = "certificats/goHomeCAcert.pem"
	var crtFile string = "certificats/server.crt.pem"
	var keyFile string = "certificats/server.key.pem"

	http.HandleFunc("/", rootHTTPSHandler)

	//-----------------------------
	caCert, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		glog.Errorf("Error reading CA cert (%s) : %s", caCertFile, err)
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
		Addr:      fmt.Sprintf("%s:%d", serverName, port), //":8080",
		TLSConfig: tlsConfig,
	}

	if glog.V(1) {
		glog.Infof("Starting ListenAndServeTLS (%d)", port)
	}

	err = server.ListenAndServeTLS(crtFile, keyFile)
	if err != nil {
		glog.Errorf("Error starting HTTPS ListenAndServeTLS : %s ... exiting", err)
		chanExit <- true
	}

}

func startHTTPS_works(chanExit chan bool) {

	db, err := openDB()
	if err != nil {
		glog.Error(err)
		return
	}
	defer db.Close()

	value, err := getGlobalParam(db, -1, "goHome", "port")
	if err != nil {
		glog.Errorf("Error getting port# : %s", err)
		return
	}
	port, err := strconv.Atoi(value)
	if err != nil {
		glog.Errorf("Error converting port# (%s) : %s", value, err)
		return
	}

	// TODO read certFileNames from DB
	var crtFile string = "server.rsa.crt"
	var keyFile string = "server.rsa.key"

	http.HandleFunc("/", rootHTTPSHandler)

	if glog.V(1) {
		glog.Infof("Starting ListenAndServeTLS (%d)", port)
	}

	err = http.ListenAndServeTLS(fmt.Sprintf("%s:%d", serverName, port), crtFile, keyFile, nil)
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

func authenticate() {
	// client cert

	// get user data + profil
}

func readUser() {
	// all, email, id
}

func writeUser() {
	// all, email, id
}

func registerUserCert() {
	// all, email, id
}

// -----------------------------------------------

func readItemTypeList() {
	// TODO : manque liens entre itemType
}

func readItemDescription() {
	// all, item type, item id

	// items + itemsFields
}

func readObject() {
	// item type, obj id

	// TODO linked objs
}

func writeObject() {

}

func readHistoSensor() {
	// Sensor Id Lst, Sensor Name lst, period, last
}

func readHistoActor() {
	// Actor Id Lst, Actor Name lst, period, last
}

func callActor() {
	// name, id
}
