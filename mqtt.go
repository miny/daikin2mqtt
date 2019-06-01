/// mqtt.go ---

// mqtt topic
//
// aircon:
// [type]/[name]/power
//              /power/set
//              /mode
//              /mode/set
//              /temperature
//              /temperature/set
// sensor/[type]/[name]/temperature
//                     /humidity
//                     /outtemp
//
// circulator:
// [type]/[name]/power
//              /power/set
//              /fanmode
//              /fanmode/set

package main

import (
	"errors"
	"fmt"
	"strconv"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const mqttBroker = "tcp://localhost:1883"

// receive topics
var (
	topicAircon = []string{
		"power/set",
		"mode/set",
		"temperature/set",
	}
	topicCirculator = []string{
		"power/set",
		"fanmode/set",
	}
)

func makeTopics(config []daikinConfig) (map[string]byte, error) {
	topics := make(map[string]byte)

	for _, cfg := range config {
		if len(cfg.Type) == 0 || len(cfg.Name) == 0 {
			err := errors.New("invalid target type or name.")
			return topics, err
		}
		if cfg.Type == "aircon" {
			for _, t := range topicAircon {
				s := fmt.Sprintf("%s/%s/%s", cfg.Type, cfg.Name, t)
				topics[s] = byte(0)
			}
		}
		if cfg.Type == "circulator" {
			for _, t := range topicCirculator {
				s := fmt.Sprintf("%s/%s/%s", cfg.Type, cfg.Name, t)
				topics[s] = byte(0)
			}
		}
	}

	return topics, nil
}

func mqttInit(config []daikinConfig) (mqtt.Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(mqttBroker)
	opts.SetAutoReconnect(true)
	opts.SetDefaultPublishHandler(
		func(c mqtt.Client, msg mqtt.Message) {
			recvmsg <- [2]string{msg.Topic(), string(msg.Payload())}
		})

	client := mqtt.NewClient(opts)
	if t := client.Connect(); t.Wait() && t.Error() != nil {
		return nil, t.Error()
	}

	topics, err := makeTopics(config)
	if err != nil {
		return nil, err
	}

	if t := client.SubscribeMultiple(topics, nil); t.Wait() && t.Error() != nil {
		return nil, t.Error()
	}

	return client, nil
}

func mqttSendOne(client mqtt.Client, topic string, value interface{}) {
	var s string
	switch value.(type) {
	case string:
		s = value.(string)
	case bool:
		if value.(bool) {
			s = "on"
		} else {
			s = "off"
		}
	default:
		s = fmt.Sprint(value)
	}
	token := client.Publish(topic, 0, true, s)
	token.Wait()
}

func mqttSendAircon(client mqtt.Client, cfg *daikinConfig, stat *daikinStat) {
	s := fmt.Sprintf("%s/%s", cfg.Type, cfg.Name)

	if stat.power == daikinStatPowerOn {
		mqttSendOne(client, s+"/power", "on")
		switch stat.mode {
		case daikinStatModeAuto:
			mqttSendOne(client, s+"/mode", "auto")
		case daikinStatModeCool:
			mqttSendOne(client, s+"/mode", "cool")
		case daikinStatModeHeat:
			mqttSendOne(client, s+"/mode", "heat")
		}
	} else {
		mqttSendOne(client, s+"/power", "off")
		mqttSendOne(client, s+"/mode", "off")
	}
	if chktemp(stat.temp) {
		mqttSendOne(client, s+"/temperature", stat.temp)
	}

	s = fmt.Sprintf("sensor/%s/%s", cfg.Type, cfg.Name)

	if chktemp(stat.intemp) {
		mqttSendOne(client, s+"/temperature", stat.intemp)
	}
	if chktemp(stat.inhum) {
		mqttSendOne(client, s+"/humidity", stat.inhum)
	}
	if chktemp(stat.outtemp) {
		mqttSendOne(client, s+"/outtemp", stat.outtemp)
	}
}

func mqttSendCirculator(client mqtt.Client, cfg *daikinConfig, stat *daikinStat) {
	s := fmt.Sprintf("%s/%s", cfg.Type, cfg.Name)

	if stat.power == daikinStatPowerOn {
		mqttSendOne(client, s+"/power", "on")
	} else {
		mqttSendOne(client, s+"/power", "off")
	}
	switch stat.fan {
	case daikinStatFanLow:
		mqttSendOne(client, s+"/fanmode", "low")
	case daikinStatFanMedium:
		mqttSendOne(client, s+"/fanmode", "medium")
	case daikinStatFanHigh:
		mqttSendOne(client, s+"/fanmode", "high")
	}
}

func chktemp(s string) bool {
	if len(s) == 0 {
		return false
	}
	_, err := strconv.ParseFloat(s, 32)
	return err == nil
}

/// mqtt.go ends here
