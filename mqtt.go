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
		for _, t := range topicAircon {
			s := fmt.Sprintf("%s/%s/%s", cfg.Type, cfg.Name, t)
			topics[s] = byte(0)
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
			s = "ON"
		} else {
			s = "OFF"
		}
	default:
		s = fmt.Sprint(value)
	}
	token := client.Publish(topic, 0, true, s)
	token.Wait()

	//fmt.Printf("%s: %s\n", topic, s)
}

func mqttSendAircon(client mqtt.Client, cfg daikinConfig, stat *daikinStat) {
	s := fmt.Sprintf("%s/%s", cfg.Type, cfg.Name)

	if stat.power {
		mqttSendOne(client, s+"/power", "ON")
		switch stat.mode {
		case daikinStatModeAuto:
			mqttSendOne(client, s+"/mode", "auto")
		case daikinStatModeCool:
			mqttSendOne(client, s+"/mode", "cool")
		case daikinStatModeHeat:
			mqttSendOne(client, s+"/mode", "heat")
		}
	} else {
		mqttSendOne(client, s+"/power", "OFF")
		mqttSendOne(client, s+"/mode", "off")
	}
	mqttSendOne(client, s+"/temperature", stat.temp)

	s = fmt.Sprintf("sensor/%s/%s", cfg.Type, cfg.Name)

	mqttSendOne(client, s+"/temperature", stat.intemp)
	mqttSendOne(client, s+"/humidity", stat.inhum)
	mqttSendOne(client, s+"/outtemp", stat.outtemp)
}

func mqttSendCirculator(client mqtt.Client, cfg daikinConfig, stat *daikinStat) {
	s := fmt.Sprintf("%s/%s", cfg.Type, cfg.Name)

	mqttSendOne(client, s+"/power", stat.power == daikinStatPowerOn)
	switch stat.fan {
	case daikinStatFanLow:
		mqttSendOne(client, s+"/fanmode", "low")
	case daikinStatFanMedium:
		mqttSendOne(client, s+"/fanmode", "medium")
	case daikinStatFanHigh:
		mqttSendOne(client, s+"/fanmode", "high")
	}
}

/// mqtt.go ends here
