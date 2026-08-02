package mqttkit

import (
	"sync"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/example/go-ai-scaffold/pkg/class/exception"
	"github.com/example/go-ai-scaffold/pkg/cli/configkey"
	"github.com/example/go-ai-scaffold/pkg/library/cryptokit"
	"github.com/example/go-ai-scaffold/pkg/service/configkit"
)

var client *Client

var _once sync.Once

func New() *Client {
	_once.Do(func() {
		if configkit.GetString(configkey.MQTTBroker) == "" {
			panic(exception.New("请填写broker"))
		}
		client = NewClient(ConnectParam{
			Broker:   configkit.GetString(configkey.MQTTBroker),
			Id:       configkit.GetString(configkey.MQTTClientID, cryptokit.ID()),
			Username: configkit.GetString(configkey.MQTTUsername),
			Pwd:      configkit.GetString(configkey.MQTTPwd),
		})
	})
	return client
}

func Subscribe(topic string, qos byte, callback MQTT.MessageHandler) {
	if client == nil {
		New()
	}
	client.Subscribe(topic, qos, callback)
}

func Publish(topic string, qos byte, retained bool, payload any) error {
	if client == nil {
		New()
	}
	return client.Publish(topic, qos, retained, payload)
}

func GetClient() MQTT.Client {
	return client.C
}
