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
//                     /compressor
//                     /outtemp
//
// circulator:
// [type]/[name]/power
//              /power/set
//              /fanmode
//              /fanmode/set
//

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

/// mqtt.go ends here
