/// daikin2mqtt.go ---

// power (pow=) 0:off 1:on
// aircon mode (mode=) 0:auto 2:dehum 3:cool 4:heat fan:6
// circulator fanmode (f_rate=) 1:low 2:medium 3:high

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	exitCode = 0
	recvmsg  = make(chan [2]string, 5)
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

	client, err := mqttInit(config)
	if err != nil {
		exitCode = 1
		fmt.Println(err)
		return
	}

	// mqtt receive message loop
	for {
		select {
		case msg := <-recvmsg:
			topic, payload := msg[0], msg[1]
			cfg := matchConfig(config, topic)
			subtopic := topic[len(cfg.Type)+len(cfg.Name)+2:]
			controlTarget(cfg, subtopic, payload)
			updateStatusOne(cfg, client)

		case <-time.After(5 * time.Minute):
			updateStatus(config, client)
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func updateStatusOne(cfg *daikinConfig, client mqtt.Client) {
	stat, err := getStatus(cfg)
	if err != nil {
		fmt.Println(err)
		return
	}

	cfg.curstat = *stat

	switch cfg.Type {
	case "aircon":
		mqttSendAircon(client, cfg, stat)
	case "circulator":
		mqttSendCirculator(client, cfg, stat)
	}
}

func updateStatus(config []daikinConfig, client mqtt.Client) {
	for i, _ := range config {
		updateStatusOne(&config[i], client)
	}
}

func controlTarget(cfg *daikinConfig, topic, payload string) {
	stat := new(daikinStat)

	switch topic {
	case "power/set":
		if (payload == "on" && cfg.curstat.power == daikinStatPowerOn) ||
			(payload == "off" && cfg.curstat.power == daikinStatPowerOff) {
			return
		}
		switch payload {
		case "on":
			stat.power = daikinStatPowerOn
		case "off":
			stat.power = daikinStatPowerOff
		}
	case "mode/set":
		switch payload {
		case "off", "offline", "False":
			stat.power = daikinStatPowerOff
		case "auto":
			stat.power = daikinStatPowerOn
			stat.mode = daikinStatModeAuto
		case "cool":
			stat.power = daikinStatPowerOn
			stat.mode = daikinStatModeCool
		case "heat":
			stat.power = daikinStatPowerOn
			stat.mode = daikinStatModeHeat
		}
	case "temperature/set":
		stat.temp = payload
	case "fanmode/set":
		switch payload {
		case "low":
			stat.fan = daikinStatFanLow
		case "medium":
			stat.fan = daikinStatFanMedium
		case "high":
			stat.fan = daikinStatFanHigh
		}
	}

	for i := 0; i < 5; i++ {
		_, err := setControl(cfg, stat)
		if err == nil {
			break
		}
		fmt.Println(err)
		time.Sleep(3 * time.Second)
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

func matchConfig(config []daikinConfig, topic string) *daikinConfig {
	for i, cfg := range config {
		s := fmt.Sprintf("%s/%s", cfg.Type, cfg.Name)
		if strings.Index(topic, s) == 0 {
			return &config[i]
		}
	}
	return nil
}

/// daikin2mqtt.go ends here
