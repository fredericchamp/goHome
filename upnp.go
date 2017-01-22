// upnp.go
package main

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"sync"

	"github.com/golang/glog"
	"github.com/huin/goupnp"
	"github.com/huin/goupnp/dcps/internetgateway1"
)

// -----------------------------------------------

const (
	TagLocalHost = "@localhost@"
)

// -----------------------------------------------

type PortMapping struct {
	srcPort  uint16
	destPort uint16
	destHost string
	protocol string
}

// -----------------------------------------------

var mappingLock sync.Mutex
var mappingTab []PortMapping

var clientIGD *internetgateway1.WANIPConnection1

var localIP string

// -----------------------------------------------

func init() {
	RegisterInternalFunc(SensorFunc, "GetExternalIP", GetExternalIP)
}

//
func upnpSetup(db *sql.DB) (err error) {

	if glog.V(1) {
		glog.Infof("upnpSetup Start")
	}

	localIP, err = GetOutboundIP()
	if err != nil {
		glog.Errorf("GetOutboundIP fail : %s", err)
		return
	}

	rootDevices, err := goupnp.DiscoverDevices("urn:schemas-upnp-org:device:InternetGatewayDevice:1")
	if err != nil {
		glog.Errorf("DiscoverDevices fail : %s", err)
		return
	}
	if len(rootDevices) <= 0 {
		err = errors.New("DiscoverDevices fail to discover a root device")
		glog.Error(err)
		return
	}
	if glog.V(2) {
		for i, device := range rootDevices {
			glog.Infof("DiscoverDevices[%d] = %v\n", i, device)
		}
	}

	// We will not handle here network with multiple Root Device, we use only the first one
	clients, err := internetgateway1.NewWANIPConnection1ClientsByURL(rootDevices[0].Location)
	if err != nil {
		glog.Errorf("NewWANIPConnection1ClientsByURL fail : %s", err)
		return
	}
	if len(rootDevices) <= 0 {
		err = errors.New("NewWANIPConnection1ClientsByURL fail to create a client")
		glog.Error(err)
		return
	}
	if glog.V(2) {
		for i, client := range clients {
			glog.Infof("IGD Client[%d] = %v\n", i, client)
		}
	}

	// We will not handle here multiple client, we use only the first one
	clientIGD = clients[0]

	portMap, err := getGlobalParamList(db, "UPnP")
	if err != nil {
		glog.Errorf("getGlobalParamList(db, UPnP) fail : %s", err)
		return
	}
	for source, target := range portMap {
		srcPort, err := strconv.Atoi(source)
		if err != nil {
			glog.Errorf("Bad source port number (%s) : ", source, err)
			continue
		}
		dest := strings.Split(target, ":")
		if len(dest) != 2 {
			err = errors.New("Mapping target decode fail : " + target)
			glog.Errorf(err.Error())
			continue
		}
		destPort, err := strconv.Atoi(dest[1])
		if err != nil {
			glog.Errorf("Bad dest port number (%s) : ", dest[1], err)
			continue
		}
		destHost := strings.Replace(dest[0], TagLocalHost, localIP, -1)
		addPortMapping(PortMapping{uint16(srcPort), uint16(destPort), destHost, "TCP"})

	}

	if glog.V(1) {
		glog.Infof("upnpSetup Done")
	}

	return nil
}

//
func upnpCleanup() {
	if clientIGD != nil {
		mappingLock.Lock()
		defer mappingLock.Unlock()

		for _, pm := range mappingTab {
			if glog.V(2) {
				glog.Infof("upnpCleanup %d(%s)", pm.srcPort, pm.protocol)
			}
			if err := clientIGD.DeletePortMapping("", pm.srcPort, pm.protocol); err != nil {
				glog.Errorf("DeletePortMapping failed for %d(%s) : %s ", pm.srcPort, pm.protocol, err)
			}
		}
		mappingTab = nil
	}
	if glog.V(1) {
		glog.Infof("upnpCleanup Done")
	}

	return
}

//
func GetExternalIP(param1 string, param2 string) (externalIP string, err error) {

	if clientIGD == nil {
		err = errors.New("GetExternalIP cancel : No clientIGD available")
		return
	}

	externalIP, err = clientIGD.GetExternalIPAddress()
	if err != nil {
		glog.Errorf("GetExternalIPAddress fail : %s", err)
		return
	}
	if glog.V(1) {
		glog.Infof("Found external IP = %s", externalIP)
	}

	return
}

//
func addPortMapping(mapping PortMapping) {
	if clientIGD == nil {
		return
	}
	mappingLock.Lock()
	defer mappingLock.Unlock()

	if glog.V(1) {
		glog.Infof("Adding port mapping (from=%d, to=%s:%d)", mapping.srcPort, mapping.destHost, mapping.destPort)
	}
	err := clientIGD.AddPortMapping("", mapping.srcPort, mapping.protocol, mapping.destPort, mapping.destHost, true, "goHome", 0)
	if err != nil {
		glog.Errorf("AddPortMapping failed : %s ", err)
		return
	}
	mappingTab = append(mappingTab, mapping)
	return
}

// -----------------------------------------------
// -----------------------------------------------
