package aqara

import (
	"encoding/json"
	"log"
	"net"
	"time"
)

var gatewayMulticast = net.UDPAddr{
	IP:   net.IPv4(224, 0, 0, 50),
	Port: 4321,
}

type IamMessage struct {
	Port         string `json:"port"`
	Sid          string `json:"sid"`
	Model        string `json:"model"`
	ProtoVersion string `json:"proto_version"`
	IP           string `json:"ip"`
}

// Discover Xiaomi gateway in LAN
func Discover(iface string, timeout time.Duration) ([]Gateway, error) {
	ifaceAddr, err := getInterfaceAddr(iface)
	if err != nil {
		return nil, err
	}

	addr := net.UDPAddr{
		IP:   ifaceAddr,
		Port: 0,
	}

	con, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return nil, err
	}
	defer con.Close()

	log.Printf("sending whois to %v", gatewayMulticast)

	if _, err = con.WriteToUDP([]byte("{\"cmd\":\"whois\"}"), &gatewayMulticast); err != nil {
		return nil, err
	}

	con.SetReadDeadline(time.Now().Add(timeout))
	var buffer = make([]byte, 1024)
	var message IamMessage
	var out []Gateway

	for {
		readLen, remoteAddr, err := con.ReadFromUDP(buffer)

		if err, ok := err.(net.Error); ok && err.Timeout() {
			break
		}

		if err != nil {
			return nil, err
		}

		if err = json.Unmarshal(buffer[:readLen], &message); err != nil {
			log.Printf("Unreadable message from %v: %v", remoteAddr, err)
			continue
		}

		out = append(out, NewGateway(remoteAddr.IP, message.Sid, "", iface))
	}

	return out, nil
}
