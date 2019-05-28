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

type daikinStat struct {
	power bool
}

func hoge(config daikinConfig) {
	status, err := getStatus(config)
	if err != nil {
		fmt.Println("resp error")
	}
	fmt.Println(status)
}

func getStatus(config daikinConfig) (url.Values, error) {
	retval := url.Values{}
	params := []string{}

	uri := "http://" + config.Host
	if config.Host == cloudHost {
		uri = "https://" + cloudHost
		params = append(params, "id="+config.Id)
		params = append(params, "spw="+config.Pw)
		params = append(params, fmt.Sprintf("port=%d", config.Port))
	}

	switch config.Type {
	case "aircon":
		uri += "/aircon/get_control_info"
	case "circulator":
		uri += "/circulator/get_unit_info"
	default:
		err := errors.New("invalid target type")
		return retval, err
	}

	resp, err := httpget(uri, params)
	if err != nil {
		return retval, err
	}

	return parseResp(resp)
}

func setControl(config daikinConfig, params []string) (url.Values, error) {
	retval := url.Values{}

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
		return retval, err
	}

	resp, err := httpget(uri, params)
	if err != nil {
		return retval, err
	}

	return parseResp(resp)
}

func parseResp(resp string) (url.Values, error) {
	retval := url.Values{}
	ok := false
	for _, kv := range strings.Split(resp, ",") {
		pair := strings.Split(kv, "=")
		if len(pair) != 2 {
			continue
		}
		k, v := pair[0], pair[1]
		if len(k) == 0 || len(v) == 0 {
			continue
		}
		if k == "ret" && v == "OK" {
			ok = true
		}
		retval.Add(k, v)
	}

	if ok {
		return retval, nil
	}

	err := errors.New("response is not OK.")
	return retval, err
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
