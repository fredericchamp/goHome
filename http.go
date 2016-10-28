// main.go
package main

import (
	"fmt"
	"net/http"

	"github.com/golang/glog"
)

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
