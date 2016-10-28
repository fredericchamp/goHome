// main.go
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/golang/glog"
)

// -----------------------------------------------
const (
	ProfilAdmin = 1 + iota
	ProfilUser
)

type itemType int

const (
	ItemEntity itemType = 1 + iota
	ItemSensor
	ItemActor
	ItemSensorAct

// TODO	ItemStreamSensor
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

// Reminder : v flags for glog
// -v=2
// -vmodule=actor=2,sens*=1

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

// signalSetup Handle OS signals.
// send true to the done chanel when server should end
func signalSetup(done chan bool) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	if glog.V(1) {
		glog.Info("signalSetup done")
	}
	s := <-c
	glog.Infof("Got signal [%s] ... exiting", s)
	signal.Reset()
	done <- true
}

// -----------------------------------------------
// -----------------------------------------------

func main() {
	defer glog.Flush()
	defer glog.Info("Bye !")

	done := make(chan bool)

	if glog.V(1) {
		glog.Info("sqlite3 file = ", *dbfile)
		if *debug {
			glog.Info("debug mode ON ")
		}
	}

	go signalSetup(done)

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

	go startHTTP(port)

	err = sensorSetup(db)
	if err != nil {
		glog.Errorf("sensorSetup failed : %s ... exiting", err)
		return
	}
	defer sensorCleanup()

	err = actorSetup(db)
	if err != nil {
		glog.Errorf("actorSetup failed : %s ... exiting", err)
		return
	}
	defer actorCleanup()

	glog.Infof("---*--- %s up and running ---*---", filepath.Base(os.Args[0]))

	<-done
	defer glog.Info("Main done")
}
