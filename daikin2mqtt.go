/// daikin2mqtt.go ---

// mqtt topic
//
// [type]/[name]/power
//              /power/set
//              /mode
//              /mode/set
//              /temperature
//              /temperature/set
//              /fanspeed
//              /fanspeed/set
// sensor/[type]/[name]/temperature
//                     /humidity
//                     /compressor
//
// aircon mode 0:auto 2:dehum 3:cool 4:heat fan:6
// aircon fanspeed 1:low 2:medium 3:high
// circulator fanspeed 1:low 2:medium 3:high

package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type daikinConfig struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Host string `json:"host"`
	Id   string `json:"id"`
	Pw   string `json:"pw"`
	Port uint   `json:"port"`
}

type stateCirculator struct {
	power bool
}

var (
	exitCode = 0
)

func main() {
	mainFunc()
	os.Exit(exitCode)
}

func mainFunc() {
	if len(os.Args) < 2 {
		fmt.Printf("usage: %s <FILE>\n", os.Args[0])
		exitCode = 1
		return
	}

	config, err := readConfig(os.Args[1])
	if err != nil {
		exitCode = 1
		fmt.Println(err)
		return
	}

	fmt.Println(config)

	for _, cfg := range config {
		hoge(cfg)
	}
}

func readConfig(fn string) ([]daikinConfig, error) {
	var config []daikinConfig

	fp, err := os.Open(fn)
	if err != nil {
		return config, err
	}

	jsondec := json.NewDecoder(fp)
	jsondec.Decode(&config)

	return config, nil
}

/// daikin2mqtt.go ends here
