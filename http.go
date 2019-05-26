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

func getStatus(config daikinConfig) (url.Values, error) {
	retval := url.Values{}
	params := url.Values{}
	url := ""
	if config.Host == cloudHost {
		url = "https://" + cloudHost
		params.Add("id", config.Id)
		params.Add("spw", config.Pw)
		params.Add("port", fmt.Sprintf("%d", config.Port))
	} else {
		url = "http://" + config.Host
	}

	switch config.Type {
	case "aircon":
		url += "/aircon/get_control_info"
	case "circulator":
		url += "/circulator/get_unit_info"
	default:
		err := errors.New("invalid target type")
		return retval, err
	}

	resp, err := httpget(url, params)
	if err != nil {
		return retval, err
	}

	fmt.Println(resp)
	return parseResp(resp)
}

func httpget(url string, values url.Values) (string, error) {
	params := values.Encode()
	if len(params) > 0 {
		url += "?" + params
	}

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	return string(b), nil
}

/// http.go ends here
