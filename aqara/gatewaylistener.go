package aqara

import (
	"errors"
	"fmt"
	"log"
	"net"
)

// GatewayListener provide a way to listen to a list of gateways
type GatewayListener struct {
	Gateways []Gateway
}

// NewGatewayListener create a GatewayListener instance
func NewGatewayListener(gateways []Gateway) GatewayListener {
	return GatewayListener{
		Gateways: gateways,
	}
}

func (g *GatewayListener) validateGatewaySameIface() error {
	if len(g.Gateways) == 0 {
		return errors.New("no gateway configured for GatewayListener")
	}

	expected := g.Gateways[0].Iface
	for _, item := range g.Gateways {
		if item.Iface != expected {
			return fmt.Errorf("configured gateways have different interface. use several GatewayListener to listen.\nexpected %v, found %v", expected, item.Iface)
		}
	}

	return nil
}

// Listen start listening for traffic from the gateways
func (g *GatewayListener) Listen(c chan ListenResponse) error {
	if err := g.validateGatewaySameIface(); err != nil {
		return err
	}

	iface, err := net.InterfaceByName(g.Gateways[0].Iface)
	if err != nil {
		return err
	}

	con, err := net.ListenMulticastUDP("udp4", iface, &net.UDPAddr{
		IP:   net.IPv4(224, 0, 0, 50),
		Port: 9898,
	})
	defer con.Close()

	buffer := make([]byte, 1024)

	for {
		size, addr, err := con.ReadFromUDP(buffer)
		if err != nil {
			return err
		}

		gateway := g.getGatewayFromAddr(addr)
		if gateway == nil {
			log.Printf("message from unknown gateway %v", addr)
			continue
		}

		message, err := parseBuffer(buffer[:size], gateway)
		if err != nil {
			log.Printf("error parsing message %v", err)
			continue
		}
		c <- ListenResponse{
			Message: *message,
			Gateway: gateway,
		}
	}
}

func (g *GatewayListener) getGatewayFromAddr(addr *net.UDPAddr) *Gateway {
	for _, gateway := range g.Gateways {
		if gateway.IP.Equal(addr.IP) {
			return &gateway
		}
	}
	return nil
}
