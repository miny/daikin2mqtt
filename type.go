/// type.go ---

package main

type daikinConfig struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Host    string `json:"host"`
	Id      string `json:"id"`
	Pw      string `json:"pw"`
	Port    uint   `json:"port"`
	curstat daikinStat
}

const (
	daikinStatPowerOff  = "0"
	daikinStatPowerOn   = "1"
	daikinStatModeAuto  = "0"
	daikinStatModeCool  = "3"
	daikinStatModeHeat  = "4"
	daikinStatFanLow    = "1"
	daikinStatFanMedium = "2"
	daikinStatFanHigh   = "3"

	daikinParamPower   = "pow"
	daikinParamMode    = "mode"
	daikinParamFan     = "f_rate"
	daikinParamTemp    = "stemp"
	daikinParamHum     = "shum"
	daikinParamInTemp  = "htemp"
	daikinParamInHum   = "hhum"
	daikinParamOutTemp = "otemp"
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

/// type.go ends here
