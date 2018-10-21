package aqara

import (
	"encoding/json"
	"errors"
	"net"
)

// Get one of the interface's unicast IP address
func getInterfaceAddr(iface string) (net.IP, error) {
	dev, err := net.InterfaceByName(iface)
	if err != nil {
		return nil, err
	}
	devAddr, err := dev.Addrs()
	if err != nil {
		return nil, err
	}
	if len(devAddr) == 0 {
		return nil, errors.New("interface has no address")
	}

	return devAddr[0].(*net.IPNet).IP, nil
}

func parseBuffer(buffer []byte, gateway *Gateway) (*ReportMessage, error) {
	var message internalReportMessage
	if err := json.Unmarshal(buffer, &message); err != nil {
		return nil, err
	}

	if gateway != nil && message.Token != "" {
		gateway.token = message.Token
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(message.Data), &data); err != nil {
		return nil, err
	}

	if val, ok := data["error"]; ok {
		return nil, errors.New(val.(string))
	}

	return &ReportMessage{
		Model: message.Model,
		Sid:   message.Sid,
		Data:  data,
	}, nil
}
