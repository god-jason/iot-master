package iot

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/god-jason/iot-master/pkg/log"
)

var client mqtt.Client

func checkMqtt() {
	if client == nil {
		opts := mqtt.NewClientOptions()
		opts.AddBroker("tcp://www.santaiyun.com:1883")
		opts.SetClientID(fmt.Sprintf("santaiyun-mqtt-%s", rand.Int()))
		opts.SetUsername("admin")
		opts.SetPassword("pzdadmin123")
		opts.SetConnectRetry(true) //重试
		opts.SetKeepAlive(50 * time.Second)

		client = mqtt.NewClient(opts)
		token := client.Connect()
		token.Wait()
		if token.Error() != nil {
			log.Println(token.Error())
		}
	}
}

func santaiyunReport(id string, name string, value int) {
	checkMqtt()
	data := map[string]any{
		id: []map[string]any{{
			"name":  name,
			"value": value,
		}},
	}
	buf, _ := json.Marshal(data)
	client.Publish("mtopic", 0, false, buf)
}
