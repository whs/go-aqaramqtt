package main

import (
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/whs/go-aqaramqtt/aqara"
)

type AqaraMqtt struct {
	mqtt            mqtt.Client
	gateway         aqara.Gateway
	gatewayListener aqara.GatewayListener
	Prefix          string

	aqaraChan chan aqara.ListenResponse
}

// NewAqaraMqtt return new instance of AqaraMqtt
func NewAqaraMqtt(mqttOptions *mqtt.ClientOptions, gateway *aqara.Gateway, prefix string) *AqaraMqtt {
	mqttOptions.SetWill(fmt.Sprintf("%s/status", prefix), "0", 1, true)
	mqttOptions.SetOnConnectHandler(func(client mqtt.Client) {
		if token := client.Publish(fmt.Sprintf("%s/status", prefix), 1, true, "1"); token.Wait() && token.Error() != nil {
			log.Printf("LWT publish fail: %v", token.Error())
		}
	})
	mqttClient := mqtt.NewClient(mqttOptions)
	// mqtt.DEBUG = log.New(os.Stdout, "[mqtt] ", log.LstdFlags)
	// mqtt.ERROR = mqtt.DEBUG
	// mqtt.WARN = mqtt.DEBUG
	// mqtt.CRITICAL = mqtt.DEBUG

	return &AqaraMqtt{
		mqtt:            mqttClient,
		gateway:         *gateway,
		gatewayListener: aqara.NewGatewayListener([]aqara.Gateway{*gateway}),
		aqaraChan:       make(chan aqara.ListenResponse),
		Prefix:          prefix,
	}
}

// Start AqaraMqtt server
func (a *AqaraMqtt) Start() error {
	go a.listenAqara()
	a.connectMqtt()
	go a.streamAqara()

	if err := a.registerDevices(); err != nil {
		return err
	}

	return nil
}

func (a *AqaraMqtt) listenAqara() {
	err := a.gatewayListener.Listen(a.aqaraChan)
	if err != nil {
		panic(err)
	}
}

func (a *AqaraMqtt) connectMqtt() {
	log.Print("Connecting to MQTT")
	if token := a.mqtt.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	log.Print("Connected to MQTT")
}

func (a *AqaraMqtt) registerDevices() error {
	if token := a.mqtt.Subscribe(fmt.Sprintf("%s/gateway/%s/set_rgb", a.Prefix, a.gateway.Sid), 0, a.setRgbHandler); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	if token := a.mqtt.Subscribe(fmt.Sprintf("%s/gateway/%s/set_mid", a.Prefix, a.gateway.Sid), 0, a.setMidHandler); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	sids, err := a.gateway.GetDevices()
	if err != nil {
		return err
	}

	for _, sid := range sids {
		status, err := a.gateway.GetDeviceStatus(sid)
		log.Printf("registering %s", sid)
		if err != nil {
			log.Printf("cannot get device %s status: %v", sid, err)
			continue
		}

		switch status.Model {
		case "plug":
			if token := a.mqtt.Subscribe(fmt.Sprintf("%s/plug/%s/set_status", a.Prefix, sid), 0, a.setStatusHandler); token.Wait() && token.Error() != nil {
				return token.Error()
			}
		}

		if err := a.publishReport(status); err != nil {
			return err
		}
	}

	return nil
}

func (a *AqaraMqtt) streamAqara() {
	for {
		message := <-a.aqaraChan
		if err := a.publishReport(&message.Message); err != nil {
			log.Println(err)
		}
	}
}

func (a *AqaraMqtt) publishReport(report *aqara.ReportMessage) error {
	for k, v := range report.Data {
		if token := a.mqtt.Publish(fmt.Sprintf("%s/%s/%s/%s", a.Prefix, report.Model, report.Sid, k), 0, true, fmt.Sprintf("%s", v)); token.Wait() && token.Error() != nil {
			return token.Error()
		}
	}

	return nil
}

func (a *AqaraMqtt) setRgbHandler(_ mqtt.Client, message mqtt.Message) {
	log.Print(message)
}

func (a *AqaraMqtt) setMidHandler(_ mqtt.Client, message mqtt.Message) {
	log.Print(message)
}

func (a *AqaraMqtt) setStatusHandler(_ mqtt.Client, message mqtt.Message) {
	log.Print(message)
}
