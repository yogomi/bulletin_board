package main

import (
	"../common_libs/address_helper"
	"fmt"
	"net"
)

var bufferByte = 64
var port = "1235"
var networkDeviceName = "en0"

func main() {
	networkDevice, err := net.InterfaceByName(networkDeviceName)
	Error(err)
	addrs, err := networkDevice.Addrs()
	Error(err)

	broadcastIP, err := addressHelper.GetIPv4BroadcastAddressFromAddressList(addrs)
	Error(err)

	broadcastAddr, err := net.ResolveUDPAddr("udp", broadcastIP.String() + ":" + port)
	Error(err)
	fmt.Println(broadcastIP)

	listener, err := net.ListenUDP("udp", broadcastAddr)
	Error(err)

	defer listener.Close()

	buffer := make([]byte, bufferByte)
	for {
		length, inboundFromAddrByte, err := listener.ReadFrom(buffer)
		Error(err)

		inboundMessage := string(buffer[:length])
		inboundFromAddr := inboundFromAddrByte.(*net.UDPAddr).String()
		fmt.Printf("Inbound %v > %v as “%s”\n", inboundFromAddr, inboundMessage, inboundMessage)
	}
}

// Error は_errをnilかどうかチェックしてnil以外だった場合panicで止める関数
func Error(_err error) {
	if _err != nil {
		panic(_err)
	}
}
