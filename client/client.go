package main

import (
	"../common/annouce_type"
	"../common_libs/address_helper"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"os/signal"
	"time"
)

func WaitSignal(endFlag *bool) {
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	*endFlag = true
}

var intervalTime time.Duration = 10
var port = "1235"
var networkDeviceName = "ローカル エリア接続* 4"

func main() {
	interfaces, _ := net.Interfaces()
	fmt.Println(interfaces)

	var firstInterval = true
	var watchingBroadcastIP net.IP = nil
	var sender net.Conn = nil
	var endFlag bool = false

	go WaitSignal(&endFlag)
	fmt.Println(endFlag)

	for endFlag == false {
		if firstInterval {
			firstInterval = false
		} else {
			time.Sleep(intervalTime * time.Second)
		}
		networkDevice, err := net.InterfaceByName(networkDeviceName)
		if err != nil {
			watchingBroadcastIP = nil
			fmt.Printf("Device %s is not found.\n", networkDeviceName)
			continue
		}
		addrs, err := networkDevice.Addrs()
		if err != nil {
			watchingBroadcastIP = nil
			fmt.Printf("Failed to find address on device. %s\n", err)
			continue
		}

		broadcastIP, err := addressHelper.GetIPv4BroadcastAddressFromAddressList(addrs)
		if err != nil {
			watchingBroadcastIP = nil
			fmt.Printf("Failed to find address on device. %s\n", err)
			continue
		}
		if watchingBroadcastIP.String() != broadcastIP.String() {
			watchingBroadcastIP = broadcastIP

			broadcastAddr, err := net.ResolveUDPAddr("udp", broadcastIP.String()+":"+port)
			if err != nil {
				watchingBroadcastIP = nil
				fmt.Printf("Failed to resolv bulletin board server address. %s\n", err)
				continue
			}
			fmt.Println(broadcastAddr)

			if sender != nil {
				sender.Close()
			}
			sender, err = net.DialUDP("udp", nil, broadcastAddr)
			if err != nil {
				watchingBroadcastIP = nil
				fmt.Printf("Failed to connect bulletin board server. %s\n", err)
				continue
			}
		}
		if sender == nil {
			watchingBroadcastIP = nil
			fmt.Printf("Sender is nil\n")
			continue
		}

		bytes := make([]byte, 4)
		binary.BigEndian.PutUint32(bytes, annouceType.ServerAddress)
		sender.Write(bytes)
		fmt.Println(bytes)
	}
	if sender != nil {
		sender.Close()
	}
}
