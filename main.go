package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/golang/glog"
)

// -----------------------------------------------
const (
	ProfilAdmin = 1 + iota
	ProfilUser
)

const (
	ItemEntity = 1 + iota
	ItemSensor
	ItemActor
	ItemSensorAct
)

const (
	DBTypeBool = 1 + iota
	DBTypeInt
	DBTypeFloat
	DBTypeText
	DBTypeDateTime
	DBTypeBlob
)

const (
	DurationMS = "ms"
	DurationS  = "s"
	DurationM  = "m"
	DurationH  = "h"
	DurationD  = "d"
)

// -----------------------------------------------
// Version is x.y.z where
// x = release version
// y = function module version (with IHM impact)
// z = patch/fix version
const goHomeVersion = "0.1.1"

// -----------------------------------------------
// -----------------------------------------------

const defaultVLog = "1"
const defaultLogDir = "/var/goHome/log"
const defaultSqlite3File = "/var/goHome/goHome.sqlite3"

// -----------------------------------------------
// -----------------------------------------------

var dbfile = flag.String("sqlite3", defaultSqlite3File, "full path to sqlite3 database file")
var debug = flag.Bool("debug", false, "run in debug mode")

func usage() {
	//	fmt.Fprintf(os.Stderr, "usage: %s -stderrthreshold=[INFO|WARN|FATAL] -log_dir=[string]\n", os.Args[0])
	//	flag.PrintDefaults()
}

func init() {
	flag.CommandLine.Init(os.Args[0], flag.ContinueOnError)
	flag.Usage = usage
	err := flag.CommandLine.Parse(os.Args[1:])
	if err != nil {
		if err == flag.ErrHelp {
			fmt.Fprintf(os.Stderr, "usage: %s -stderrthreshold=[INFO|WARN|FATAL] -log_dir=[string]\n", os.Args[0])
			flag.PrintDefaults()
			os.Exit(2)
		} else {
			glog.Warning("Bad parameter in command line")
		}
	}

	if "" == flag.Lookup("log_dir").Value.String() {
		flag.Lookup("log_dir").Value.Set(defaultLogDir)
	}
	if "" == flag.Lookup("v").Value.String() {
		flag.Lookup("v").Value.Set(defaultVLog)
	}

}

// -----------------------------------------------
// -----------------------------------------------

// handleSignal Handle OS signals.
// send true to the done chanel when server should end
func handleSignal(done chan bool) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	s := <-c
	glog.Infof("Got signal [%s] ... exiting", s)
	signal.Reset()
	done <- true
}

// -----------------------------------------------
// -----------------------------------------------

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
}

// -----------------------------------------------
// -----------------------------------------------

func main() {
	defer glog.Flush()
	fmt.Printf("pid=%d\n", os.Getpid())

	done := make(chan bool)

	if glog.V(2) {
		glog.Info("sqlite3 file = ", *dbfile)
		glog.Info("debug = ", *debug)
	}

	go handleSignal(done)

	err := initDBFile(*dbfile)
	if err != nil {
		glog.Error("Could not init DB ... exiting")
		return
	}
	db, err := openDB()
	if err != nil {
		glog.Error(err)
		return
	}
	defer db.Close()

	value, err := getGlobalParam(db, -1, "goHome", "port")
	if err != nil {
		glog.Errorf("Error getting port# : %s ... exiting", err)
		return
	}
	port, err := strconv.Atoi(value)
	if err != nil {
		glog.Errorf("Error converting port# (%s) : %s ... exiting", value, err)
		return
	}
	if glog.V(2) {
		glog.Info("Port# = ", port)
	}
	//go startHTTP(port)

	/*
		items, err := getManageItems(db, -1, -1)
		if err != nil {
			glog.Errorf("Error getting manage items : %s ... exiting", err)
			return
		}
		glog.Info("items = ", items)

		fields, err := getItemFields(db, -1, -1)
		if err != nil {
			glog.Errorf("Error getting item fields : %s ... exiting", err)
			return
		}
		glog.Info("fields = ", fields)

			obj1, err := getDBObjects(db, 9, -1)
			if err != nil {
				glog.Errorf("Error getting obj 9 : %s ... exiting", err)
				return
			}
			glog.Info("obj 9 = ", obj1)

			objs, err := getDBObjects(db, -1, 2)
			if err != nil {
				glog.Errorf("Error getting objs for item 2 : %s ... exiting", err)
				return
			}
			glog.Info("--------------------------------\nfields = ", objs[1].Fields)
			for _, v := range objs {
				glog.Info("--------------------------------\nvalues = ", v.Values)
			}
	*/

	err = sensorSetup(db)
	if err != nil {
		glog.Errorf("SensorSetup failed : %s ... exiting", err)
		return
	}
	defer sensorCleanup()

	err = actorSetup(db)
	if err != nil {
		glog.Errorf("ActorSetup failed : %s ... exiting", err)
		return
	}
	defer actorCleanup()

	triggerActor("Portal", "test 1")
	triggerActor("SendSMS", "test 2")
	triggerActor("Garage", "test 3")

	<-done
	glog.Info("Bye !")
}
