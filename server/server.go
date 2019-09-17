package main

import (
	"fmt"
	"net"
)

var BufferByte int = 64
var port string = "1235"
var network_device_name string = "en0"

func main() {
	en0, err := net.InterfaceByName(network_device_name)
	Error(err)
	addrs, err := en0.Addrs()
	Error(err)
	ip, _, err := net.ParseCIDR(addrs[0].String())
	Error(err)
	var listen_ipv6_address string =
		"[" + ip.String() + "%" + network_device_name + "]:" + port
	fmt.Println(listen_ipv6_address)
	broadcast_addr, err := net.ResolveUDPAddr("udp6", listen_ipv6_address)
	Error(err)

	listener, err := net.ListenUDP("udp", broadcast_addr)
	Error(err)

	defer listener.Close()

	buffer := make([]byte, BufferByte)
	for {
		length, inbound_from_addr_byte, err := listener.ReadFrom(buffer)
		Error(err)

		inbound_message := string(buffer[:length])
		inbound_from_addr := inbound_from_addr_byte.(*net.UDPAddr).String()
		fmt.Printf("Inbound %v > %v as “%s”\n", inbound_from_addr, inbound_message, inbound_message)
	}
}

func Error(_err error) {
	if _err != nil {
		panic(_err)
	}
}
