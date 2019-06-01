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

func getStatus(cfg *daikinConfig) (*daikinStat, error) {
	params := []string{}

	uri := "http://" + cfg.Host
	if cfg.Host == cloudHost {
		uri = "https://" + cloudHost
		params = append(params, "id="+cfg.Id)
		params = append(params, "spw="+cfg.Pw)
		params = append(params, fmt.Sprintf("port=%d", cfg.Port))
	}

	var resp string
	var err error
	stat := new(daikinStat)

	switch cfg.Type {

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
		stat, err = parseResp(resp, stat)
		if err != nil {
			return nil, err
		}

	case "circulator":
		resp, err = httpget(uri+"/circulator/get_unit_info", params)
		if err != nil {
			return nil, err
		}
		stat, err = parseResp(resp, stat)
		if err != nil {
			return nil, err
		}

	default:
		err := errors.New("invalid target type")
		return nil, err
	}

	return stat, nil
}

func setControl(cfg *daikinConfig, stat *daikinStat) (*daikinStat, error) {
	uri := "http://" + cfg.Host
	if cfg.Host == cloudHost {
		uri = "https://" + cloudHost
	}

	switch cfg.Type {
	case "aircon":
		uri += "/aircon/set_control_info"
	case "circulator":
		uri += "/circulator/set_control_info"
	default:
		err := errors.New("invalid target type")
		return nil, err
	}

	params := makeParam(cfg, stat)

	resp, err := httpget(uri, params)
	if err != nil {
		return nil, err
	}

	return parseResp(resp, new(daikinStat))
}

func makeParam(cfg *daikinConfig, stat *daikinStat) (params []string) {
	if cfg.Host == cloudHost {
		params = append(params, "id="+cfg.Id)
		params = append(params, "spw="+cfg.Pw)
		params = append(params, fmt.Sprintf("port=%d", cfg.Port))
	}

	if len(stat.power) > 0 {
		params = append(params, daikinParamPower+"="+stat.power)
	} else if len(cfg.curstat.power) > 0 {
		params = append(params, daikinParamPower+"="+cfg.curstat.power)
	} else {
		params = append(params, daikinParamPower+"=0")
	}

	if cfg.Type == "aircon" {
		if len(stat.mode) > 0 {
			params = append(params, daikinParamMode+"="+stat.mode)
		} else if len(cfg.curstat.mode) > 0 {
			params = append(params, daikinParamMode+"="+cfg.curstat.mode)
		} else {
			params = append(params, daikinParamMode+"=0")
		}

		if len(stat.temp) > 0 {
			params = append(params, daikinParamTemp+"="+stat.temp)
		} else if len(cfg.curstat.temp) > 0 {
			params = append(params, daikinParamTemp+"="+cfg.curstat.temp)
		} else {
			params = append(params, daikinParamTemp+"=26.0")
		}

		if len(stat.hum) > 0 {
			params = append(params, daikinParamHum+"="+stat.hum)
		} else if len(cfg.curstat.hum) > 0 {
			params = append(params, daikinParamTemp+"="+cfg.curstat.hum)
		} else {
			params = append(params, daikinParamHum+"=CONTINUE")
		}
	}

	if cfg.Type == "circulator" {
		if len(stat.fan) > 0 {
			params = append(params, daikinParamFan+"="+stat.fan)
		}
	}

	return params
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
	//log.Print(uri)
	//log.Print(d)
	return d, nil
}

/// http.go ends here
