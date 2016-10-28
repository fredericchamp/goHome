// poloniex.go
package main

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/golang/glog"
)

const poloPubURL = "https://poloniex.com/public"
const poloGetTickerCmd = "?command=returnTicker"

const poloApiURL = "https://poloniex.com/tradingApi"
const poloGetBalanceCmd = "returnCompleteBalances"

const unixNano201601010000 = 1451606400000000000

var lastNonce uint64 = 0

const nonceErrTag = `"Nonce must be greater than `

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

func init() {
	RegisterInternalFunc(SensorFunc, "GetPoloTicker", GetPoloTicker)
	RegisterInternalFunc(SensorFunc, "GetPoloBalance", GetPoloBalance)
}

// TODO use WAMP pub api : wss://api.poloniex.com and (https://github.com/llchan/go-wamp)

// GetPoloTicker : Sensor like func to get 'last' rate for a pair (fetched from poloniex pubilc URL)
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

	if glog.V(1) {
		glog.Infof("GetPoloTicker (%s) (%s) = %s", param1, param2, oneTicker.Last)
	}
	result = oneTicker.Last

	return
}

type PoloStruct struct {
	Key    string
	Secret string
}

type PoloBalanceLine struct {
	Available string
	OnOrders  string
	BtcValue  string
}

// GetPoloBalance : Sensor like func to get complete balance for a poloniex account (using key & secret from parameters)
func GetPoloBalance(param1 string, param2 string) (result string, err error) {
	var polo PoloStruct
	err = json.Unmarshal([]byte(param1), &polo)
	if err != nil {
		glog.Errorf("Fail to unmarshal PoloStruct (%s) : %s", param1, err)
		return
	}
	if glog.V(2) {
		glog.Info("polo ", polo)
	}

	// build post data
	form := url.Values{}
	form.Add("command", poloGetBalanceCmd)
	lastNonce = lastNonce + 1
	form.Add("nonce", fmt.Sprint(lastNonce))

	// Calc hmac sha512
	mac := hmac.New(sha512.New, []byte(polo.Secret))
	mac.Write([]byte(form.Encode()))
	hexSign := hex.EncodeToString(mac.Sum(nil))

	// Prepare HTTP Request
	req, err := http.NewRequest(http.MethodPost, poloApiURL, strings.NewReader(form.Encode()))
	if err != nil {
		glog.Errorf("Fail to create new http request : %s", err)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/4.0 (compatible; Poloniex goHome server)")
	req.Header.Set("Key", polo.Key)
	req.Header.Set("Sign", hexSign)
	if glog.V(2) {
		glog.Info("req ", req)
	}

	// Exec HTTP Request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		glog.Errorf("Fail to post (%s) : %s", poloApiURL, err)
		return
	}
	defer resp.Body.Close()

	// Read response
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

	msg, found := objmap["error"]
	if found {
		err = errors.New(string(*msg))
		glog.Error(err)
		// check for nonce error
		if strings.Contains(string(*msg), nonceErrTag) {
			lastNonce, err = strconv.ParseUint(strings.Split(strings.TrimPrefix(string(*msg), nonceErrTag), ".")[0], 10, 64)
			if glog.V(2) {
				glog.Info("Nonce error => lastNonce updated, retrying with ", lastNonce)
			}
			result, err = GetPoloBalance(param1, param2)
		}
		return
	}

	var total float64 = 0
	for key, msg := range objmap {
		var oneLine PoloBalanceLine
		err = json.Unmarshal(*msg, &oneLine)
		if err != nil {
			glog.Errorf("Fail to unmarshal value for map[%s] : %s", key, err)
			return
		}
		amount, err1 := strconv.ParseFloat(oneLine.BtcValue, 64)
		if err1 != nil {
			err = err1
			glog.Errorf("Fail to parse float64 val map[%s]=%s : %s", key, string(*msg), err)
			return
		}
		total = total + amount
	}

	result = fmt.Sprintf("%.8f", total)

	return
}
