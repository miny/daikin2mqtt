/// http.go ---

package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	cloudHost = "daikinsmartdb.jp"
)

const (
	daikinStatPowerOff  = "0"
	daikinStatPowerOn   = "1"
	daikinStatModeAuto  = "0"
	daikinStatModeCool  = "3"
	daikinStatModeHeat  = "4"
	daikinStatFanLow    = "1"
	daikinStatFanMedium = "2"
	daikinStatFanHigh   = "3"
)

type daikinStat struct {
	power   string // power state
	mode    string // mode
	fan     string // fan mode
	temp    string // target temperature
	hum     string // target humidity
	intemp  string // sensor temperature
	inhum   string // sensor humidity
	outtemp string // sensor outdoor temperature
}

func hoge(config daikinConfig) {
	stat, err := getStatus(config)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(stat)
}

func getStatus(config daikinConfig) (*daikinStat, error) {
	stat := new(daikinStat)
	params := []string{}

	uri := "http://" + config.Host
	if config.Host == cloudHost {
		uri = "https://" + cloudHost
		params = append(params, "id="+config.Id)
		params = append(params, "spw="+config.Pw)
		params = append(params, fmt.Sprintf("port=%d", config.Port))
	}

	var resp string
	var err error

	switch config.Type {

	case "aircon":
		resp, err = httpget(uri+"/aircon/get_control_info", params)
		if err != nil {
			return nil, err
		}
		stat, err = parseResp(resp, stat)
		if err != nil {
			return nil, err
		}
		resp, err = httpget(uri+"/aircon/get_sensor_info", params)
		if err != nil {
			return nil, err
		}

	case "circulator":
		resp, err = httpget(uri+"/circulator/get_unit_info", params)
		if err != nil {
			return nil, err
		}

	default:
		err := errors.New("invalid target type")
		return nil, err
	}

	return parseResp(resp, stat)
}

func setControl(config daikinConfig, params []string) (*daikinStat, error) {
	stat := new(daikinStat)

	uri := "http://" + config.Host
	if config.Host == cloudHost {
		uri = "https://" + cloudHost
		params = append(params, "id="+config.Id)
		params = append(params, "spw="+config.Pw)
		params = append(params, fmt.Sprintf("port=%d", config.Port))
	}

	switch config.Type {
	case "aircon":
		uri += "/aircon/set_control_info"
	case "circulator":
		uri += "/circulator/set_control_info"
	default:
		err := errors.New("invalid target type")
		return nil, err
	}

	resp, err := httpget(uri, params)
	if err != nil {
		return nil, err
	}

	return parseResp(resp, stat)
}

func parseResp(resp string, stat *daikinStat) (*daikinStat, error) {
	var ok string

	for _, kv := range strings.Split(resp, ",") {
		pair := strings.Split(kv, "=")
		if len(pair) != 2 {
			continue
		}
		k, v := pair[0], pair[1]
		if len(k) == 0 || len(v) == 0 {
			continue
		}

		switch k {
		case "ret":
			ok = v
		case "pow":
			stat.power = v
		case "mode":
			stat.mode = v
		case "stemp":
			stat.temp = v
		case "shum":
			stat.hum = v
		case "htemp":
			stat.intemp = v
		case "hhum":
			stat.inhum = v
		case "otemp":
			stat.outtemp = v
		case "f_rate":
			stat.fan = v
		}
	}

	if ok == "OK" {
		return stat, nil
	}

	err := errors.New("response is " + ok)
	return nil, err
}

func httpget(uri string, values []string) (string, error) {
	params := strings.Join(values, "&")
	if len(params) > 0 {
		uri += "?" + params
	}

	resp, err := http.Get(uri)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	d, _ := url.QueryUnescape(string(b))
	fmt.Println(uri)
	fmt.Println(d)
	return d, nil
}

/// http.go ends here
