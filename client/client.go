package main

import (
	"fmt"
	"net"
	"os"
)

var server_address string = "[fe80::]:1235"

func main() {
	broadcast_addr, err := net.ResolveUDPAddr("udp6", server_address)
	Error(err)

	sender, err := net.DialUDP("udp", nil, broadcast_addr)
	Error(err)
	defer sender.Close()

	message, err := os.Hostname()
	sender.Write([]byte(message))
	fmt.Println(message)
}

func Error(_err error) {
	if _err != nil {
		panic(_err)
	}
}
