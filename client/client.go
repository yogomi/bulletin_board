package main

import (
	"../common_libs/address_helper"
	"fmt"
	"net"
	"os"
)

var port = "1235"
var networkDeviceName = "en0"

func main() {
	networkDevice, err := net.InterfaceByName(networkDeviceName)
	Error(err)
	addrs, err := networkDevice.Addrs()
	Error(err)

	broadcastIP, err := addressHelper.GetIPv4BroadcastAddressFromAddressList(addrs)
	Error(err)
	fmt.Println(broadcastIP)

	broadcastAddr, err := net.ResolveUDPAddr("udp", broadcastIP.String() + ":" + port)
	Error(err)

	sender, err := net.DialUDP("udp", nil, broadcastAddr)
	Error(err)
	defer sender.Close()

	message, err := os.Hostname()
	sender.Write([]byte(message))
	fmt.Println(message)
}

// Error は_errをnilかどうかチェックしてnil以外だった場合panicで止める関数
func Error(_err error) {
	if _err != nil {
		panic(_err)
	}
}
