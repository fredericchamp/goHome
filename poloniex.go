// poloniex.go
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"
)

const poloPubURL = "https://poloniex.com/public"
const poloGetTickerCmd = "?command=returnTicker"

type StructTicker struct {
	Id            int
	Last          string
	LowestAsk     string
	HighestBid    string
	PercentChange string
	BaseVolume    string
	QuoteVolume   string
	IsFrozen      string
	High24hr      string
	Low24h        string
}

var prevValues map[string]StructTicker

func init() {
	RegisterInternalFunc(SensorFunc, "GetPoloTicker", GetPoloTicker)
	prevValues = make(map[string]StructTicker)
}

// TODO use WAMP pub api : wss://api.poloniex.com and (https://github.com/llchan/go-wamp)

func GetPoloTicker(param1 string, param2 string) (result string, err error) {
	result = "0"

	resp, err := http.Get(fmt.Sprint(poloPubURL, poloGetTickerCmd))
	if err != nil {
		glog.Errorf("Fail to get URL (%s) : %s", fmt.Sprint(poloPubURL, poloGetTickerCmd), err)
		return
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf("Fail to read Body : %s", err)
		return
	}

	var objmap map[string]*json.RawMessage
	err = json.Unmarshal(data, &objmap)
	if err != nil {
		glog.Errorf("Fail to unmarshal Body : %s", err)
		return
	}

	msg, found := objmap[param1]
	if !found {
		err = errors.New(fmt.Sprintf("No value for key (%s)", param1))
		glog.Error(err)
		return
	}

	var oneTicker StructTicker
	err = json.Unmarshal(*msg, &oneTicker)
	if err != nil {
		glog.Errorf("Fail to unmarshal value for map[%s] : %s", param1, err)
		return
	}

	// keep last value
	prevValues[param1] = oneTicker

	if glog.V(2) {
		glog.Infof("GetPoloTicker (%s) (%s) = %s", param1, param2, oneTicker.Last)
	}
	result = oneTicker.Last

	return
}
