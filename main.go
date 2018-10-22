package main

import (
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/whs/go-aqaramqtt/aqara"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	flagGatewayIP    = kingpin.Flag("ip", "Gateway IP").IP()
	flagGatewaySID   = kingpin.Flag("sid", "Gateway SID").String()
	flagGatewayKey   = kingpin.Flag("key", "Gateway key").OverrideDefaultFromEnvar("AQARA_KEY").Required().String()
	flagInterface    = kingpin.Flag("iface", "Interface to run").Required().String()
	flagMqttServer   = kingpin.Flag("mqtt-server", "MQTT server including protocol (eg. tcp://example.com:1883)").Required().String()
	flagMqttUsername = kingpin.Flag("username", "MQTT username").OverrideDefaultFromEnvar("MQTT_USERNAME").String()
	flagMqttPassword = kingpin.Flag("password", "MQTT password").OverrideDefaultFromEnvar("MQTT_PASSWORD").String()
	flagMqttPrefix   = kingpin.Flag("prefix", "MQTT prefix (without trailing slash)").Default("xiaomi").String()
)

func main() {
	kingpin.Parse()

	// Setup Aqara
	var gateway aqara.Gateway
	if *flagGatewayIP != nil && flagGatewaySID != nil {
		log.Printf("Using configured gateway %v sid %v", flagGatewayIP, *flagGatewaySID)
		gateway = aqara.NewGateway(*flagGatewayIP, *flagGatewaySID, *flagGatewayKey, *flagInterface)
	} else {
		log.Print("Discovering gateway...")
		gateways, err := aqara.Discover(*flagInterface, 1*time.Second)
		if err != nil {
			log.Fatal(err)
		}
		gateway = gateways[0]
		gateway.Key = *flagGatewayKey
		log.Printf("Using gateway %v sid %s", gateway.IP, gateway.Sid)
	}

	// Setup MQTT
	mqttOptions := mqtt.NewClientOptions()
	mqttOptions.AddBroker(*flagMqttServer)
	if flagMqttUsername != nil && flagMqttPassword != nil {
		mqttOptions.SetUsername(*flagMqttUsername)
		mqttOptions.SetPassword(*flagMqttPassword)
	}

	// Start AqaraMQTT
	aqaraMqtt := NewAqaraMqtt(mqttOptions, &gateway, *flagMqttPrefix)
	if err := aqaraMqtt.Start(); err != nil {
		panic(err)
	}
}
