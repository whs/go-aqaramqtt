package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/whs/go-aqaramqtt/aqara"
)

const MidVol = 50

// AqaraMqtt bridges between Xiaomi Aqara and MQTT
type AqaraMqtt struct {
	mqtt            mqtt.Client
	gateway         *aqara.Gateway
	gatewayListener aqara.GatewayListener
	Prefix          string

	aqaraChan chan aqara.ListenResponse
	lastRgb   *HassColorMessage
}

// NewAqaraMqtt return new instance of AqaraMqtt
func NewAqaraMqtt(mqttOptions *mqtt.ClientOptions, gateway *aqara.Gateway, prefix string) *AqaraMqtt {
	out := &AqaraMqtt{
		gateway:         gateway,
		gatewayListener: aqara.NewGatewayListener([]*aqara.Gateway{gateway}),
		aqaraChan:       make(chan aqara.ListenResponse),
		Prefix:          prefix,
	}
	out.initMqtt(mqttOptions)

	return out
}

func (a *AqaraMqtt) initMqtt(mqttOptions *mqtt.ClientOptions) {
	mqttOptions.SetWill(fmt.Sprintf("%s/status", a.Prefix), "0", 1, true)
	mqttOptions.SetOnConnectHandler(func(client mqtt.Client) {
		if token := client.Publish(fmt.Sprintf("%s/status", a.Prefix), 1, true, "1"); token.Wait() && token.Error() != nil {
			log.Printf("LWT publish fail: %v", token.Error())
		}

		if err := a.registerDevices(); err != nil {
			log.Print(err)
		}
	})
	a.mqtt = mqtt.NewClient(mqttOptions)
	// mqtt.DEBUG = log.New(os.Stdout, "[mqtt] ", log.LstdFlags)
	// mqtt.ERROR = mqtt.DEBUG
	// mqtt.WARN = mqtt.DEBUG
	// mqtt.CRITICAL = mqtt.DEBUG
}

// Start AqaraMqtt server
func (a *AqaraMqtt) Start() error {
	go a.listenAqara()
	a.connectMqtt()
	a.streamAqara()
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
		var value string

		if report.Model == "gateway" && k == "rgb" {
			value = formatHassRgb(int64(v.(float64)))
		} else if report.Model == "switch" && k == "status" {
			value = v.(string)
			if value == "long_click_release" {
				value = "release"
			}
		} else if report.Model == "motion" && k == "no_motion" {
			k = "status"
			value = "no_motion"
		} else if k == "voltage" && report.Model != "plug" {
			value = convertVoltage(v.(float64))
		} else {
			value = fmt.Sprintf("%v", v)
		}

		if token := a.mqtt.Publish(fmt.Sprintf("%s/%s/%s/%s", a.Prefix, report.Model, report.Sid, k), 0, true, value); token.Wait() && token.Error() != nil {
			return token.Error()
		}

		// switch click event must be followed immediately by release event
		if report.Model == "switch" && k == "status" && value != "long_click_press" && value != "release" {
			if token := a.mqtt.Publish(fmt.Sprintf("%s/switch/%s/status", a.Prefix, report.Sid), 0, true, "release"); token.Wait() && token.Error() != nil {
				return token.Error()
			}
		}
	}

	return nil
}

func (a *AqaraMqtt) setRgbHandler(_ mqtt.Client, message mqtt.Message) {
	log.Print(message)
	sid := getSidFromMqttTopic(message.Topic())

	if a.gateway.Sid != sid {
		log.Printf("sid %s not of the gateway %s. skipping", sid, a.gateway.Sid)
		return
	}

	rgb, hassMessage, err := parseHassRgbMessage(message.Payload(), a.lastRgb)
	if err != nil {
		log.Print(err)
		return
	}
	a.lastRgb = hassMessage

	out, err := a.gateway.SetRGB(rgb)
	if err != nil {
		log.Print(err)
	} else {
		log.Print(out)
	}
}

func (a *AqaraMqtt) setMidHandler(_ mqtt.Client, message mqtt.Message) {
	log.Print(message)
	sid := getSidFromMqttTopic(message.Topic())

	if a.gateway.Sid != sid {
		log.Printf("sid %s not of the gateway %s. skipping", sid, a.gateway.Sid)
		return
	}

	mid, err := strconv.ParseUint(string(message.Payload()), 10, 32)
	if err != nil {
		log.Print(err)
		return
	}

	out, err := a.gateway.SetMID(uint(mid), MidVol)
	if err != nil {
		log.Print(err)
	} else {
		log.Print(out)
	}
}

func (a *AqaraMqtt) setStatusHandler(_ mqtt.Client, message mqtt.Message) {
	log.Print(message)
	sid := getSidFromMqttTopic(message.Topic())

	out, err := a.gateway.SetStatus(sid, string(message.Payload()))
	if err != nil {
		log.Print(err)
	} else {
		log.Print(out)
	}
}

func getSidFromMqttTopic(topic string) string {
	parts := strings.Split(topic, "/")
	return parts[2]
}

// HassColorMessage are used by Home Assistant
// https://www.home-assistant.io/components/light.mqtt_json/
type HassColorMessage struct {
	Brightness *uint8        `json:"brightness"`
	Color      *colorMessage `json:"color"`
	State      string        `json:"state"`
}

type colorMessage struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
}

func formatHassRgb(rgb int64) string {
	a := uint8((rgb & 0xFF000000) >> 24)
	r := uint8((rgb & 0x00FF0000) >> 16)
	g := uint8((rgb & 0x0000FF00) >> 8)
	b := uint8(rgb & 0x000000FF)

	var state string
	if a > 0 {
		state = "ON"
	} else {
		state = "OFF"
	}

	message := HassColorMessage{
		Brightness: &a,
		State:      state,
		Color: &colorMessage{
			R: r,
			G: g,
			B: b,
		},
	}

	out, err := json.Marshal(message)
	if err != nil {
		log.Print(err)
		return ""
	}
	return string(out)
}

func parseHassRgbMessage(message []byte, lastRgb *HassColorMessage) (uint64, *HassColorMessage, error) {
	var parsed HassColorMessage
	var zero uint8
	var max uint8 = 100
	if err := json.Unmarshal(message, &parsed); err != nil {
		return 0, nil, err
	}

	if parsed.State == "OFF" {
		parsed.Brightness = &zero
	}

	if lastRgb != nil {
		if parsed.Brightness == nil {
			parsed.Brightness = lastRgb.Brightness
		}
		if parsed.Color == nil {
			parsed.Color = lastRgb.Color
		}
	} else {
		if parsed.Brightness == nil {
			parsed.Brightness = &zero
		}
		if parsed.Color == nil {
			parsed.Color = &colorMessage{R: 0, G: 0, B: 0}
		}
	}

	if *parsed.Brightness == 0 && parsed.State == "ON" {
		parsed.Brightness = &max
	}

	var out uint64
	out |= uint64(*parsed.Brightness) << 24
	out |= uint64(parsed.Color.R) << 16
	out |= uint64(parsed.Color.G) << 8
	out |= uint64(parsed.Color.B)
	return out, &parsed, nil
}

var maxVolt float64 = 3300
var minVolt float64 = 2800

func convertVoltage(voltage float64) string {
	value := math.Max(math.Min(voltage, maxVolt), minVolt)
	out := ((value - minVolt) / (maxVolt - minVolt)) * 100
	return fmt.Sprintf("%.0f", (math.Round(out*100) / 100))
}
