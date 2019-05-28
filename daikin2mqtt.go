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

type daikinConfig struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Host string `json:"host"`
	Id   string `json:"id"`
	Pw   string `json:"pw"`
	Port uint   `json:"port"`
}

var (
	exitCode = 0
	recvmsg  = make(chan [2]string)
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

	client, err := mqttInit(config)
	if err != nil {
		exitCode = 1
		fmt.Println(err)
		return
	}

	// mqtt receive message loop
	for {
		updateStatus(config, client)

		select {
		case msg := <-recvmsg:
			topic, payload := msg[0], msg[1]
			cfg := matchConfig(config, topic)
			subtopic := topic[len(cfg.Type)+len(cfg.Name)+3:]
			fmt.Println("subtopic", subtopic)
			fmt.Println("payload", payload)
		case <-time.After(5 * time.Minute):
		}
	}

	for _, cfg := range config {
		hoge(cfg)
	}
}

func updateStatus(config []daikinConfig, client mqtt.Client) {
	for _, cfg := range config {
		stat, err := getStatus(cfg)
		if err != nil {
			fmt.Println(err)
			continue
		}

		switch cfg.Type {
		case "aircon":
			mqttSendAircon(client, cfg, stat)
		case "circulator":
			mqttSendCirculator(client, cfg, stat)
		}
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
	for _, cfg := range config {
		s := fmt.Sprintf("%s/%s", cfg.Type, cfg.Name)
		if strings.Index(topic, s) == 0 {
			return &cfg
		}
	}
	return nil
}

/// daikin2mqtt.go ends here
