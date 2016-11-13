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
	"runtime/debug"
	"strconv"
	"strings"
	"sync"

	"github.com/golang/glog"
)

const poloPubURL = "https://poloniex.com/public"
const poloGetTickerCmd = "?command=returnTicker"

const poloApiURL = "https://poloniex.com/tradingApi"
const poloGetBalanceCmd = "returnCompleteBalances"

// Note : this Mutex will not ensure that request will be process in the correct order by the receiving server
var lastNonceLock sync.Mutex
var lastNonce uint64 = 0

const poloErrTag = `{"error":`
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

	result = oneTicker.Last
	if glog.V(1) {
		glog.Infof("GetPoloTicker (%s) (%s) = %s", param1, param2, result)
	}

	return
}

type PoloStruct struct {
	Key    string
	Secret string
}

// queryPolo : Send POST request to 'poloApiURL'
// 'jsonParam' is a json encoded list of all needed param and values for 'command'
// result is the json body receive in answer to the query
func queryPolo(command string, poloKey PoloStruct, jsonParam string) (result []byte, err error) {

	var parammap map[string]string
	if len(jsonParam) > 0 {
		err = json.Unmarshal([]byte(jsonParam), &parammap)
		if err != nil {
			glog.Errorf("Fail to unmarshal jsonParam '%s' : %s", jsonParam, err)
			glog.V(3).Info(string(debug.Stack()))
			return
		}
	}

	// build post data
	form := url.Values{}
	form.Add("command", command)
	lastNonceLock.Lock()
	lastNonce = lastNonce + 1
	form.Add("nonce", fmt.Sprint(lastNonce))
	lastNonceLock.Unlock()
	for key, val := range parammap {
		form.Add(key, val)
	}

	// Calc hmac sha512
	mac := hmac.New(sha512.New, []byte(poloKey.Secret))
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
	req.Header.Set("Key", poloKey.Key)
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
	result, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf("Fail to read Body : %s", err)
		return
	}

	// On error receive, parse it
	if strings.HasPrefix(string(result[:10]), poloErrTag) {
		var objmap map[string]*json.RawMessage
		err = json.Unmarshal(result, &objmap)
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
				lastNonceLock.Lock()
				lastNonce, err = strconv.ParseUint(strings.Split(strings.TrimPrefix(string(*msg), nonceErrTag), ".")[0], 10, 64)
				lastNonceLock.Unlock()
				if glog.V(2) {
					glog.Info("Nonce error => lastNonce updated, retrying with ", lastNonce)
				}
				result, err = queryPolo(command, poloKey, jsonParam)
			}
			return
		}
	}

	return
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

	data, err := queryPolo(poloGetBalanceCmd, polo, "")
	if err != nil {
		return
	}

	var objmap map[string]*json.RawMessage
	err = json.Unmarshal(data, &objmap)
	if err != nil {
		if len(data) > 50 {
			data = data[:20]
		}
		glog.Errorf("Fail to unmarshal Body '%s' : %s", string(data), err)
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

	if glog.V(1) {
		glog.Infof("GetPoloBalance = %s", result)
	}

	return
}
