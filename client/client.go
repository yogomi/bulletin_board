package main

import (
	"../common/announce_type"
	"../common_libs/address_helper"
	"./tasks"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"os/signal"
	"time"
)

// WaitSignal は外部からのシグナルを受け取る関数。
func WaitSignal(endFlag *bool) {
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	fmt.Println("Execute ending process.")
	*endFlag = true
}

const intervalTime = 5
const udpTimeout = 3
const bufferByte = 64
var serverPort = "8120"
var clientPort = "8121"
const primaryNetworkDeviceIndex = 4
const secondaryNetworkDeviceIndex = 12

func main() {
	var firstInterval = true
	var sender *net.UDPConn
	var listener *net.UDPConn
	var endFlag = false
	var primaryMode = true

	go WaitSignal(&endFlag)

	for endFlag == false {
		networkDeviceIndex := func(primaryMode bool) int {
			if primaryMode {
				return primaryNetworkDeviceIndex
			} else {
				return secondaryNetworkDeviceIndex
			}
		} (primaryMode)
		primaryMode = !primaryMode
		if firstInterval {
			firstInterval = false
		} else {
			time.Sleep(intervalTime * time.Second)
		}
		networkDevice, err := net.InterfaceByIndex(networkDeviceIndex)
		if err != nil {
			fmt.Printf("Device index %d is not found.\n", networkDeviceIndex)
			continue
		}
		addrs, err := networkDevice.Addrs()
		if err != nil {
			fmt.Printf("Failed to find address on device. %s\n", err)
			continue
		}

		selfIP, _, broadcastIP, err := addressHelper.GetIPv4AddressSetFromAddressList(addrs)
		if err != nil {
			fmt.Printf("Failed to find address on device. %s\n", err)
			continue
		}
		broadcastAddr, err := net.ResolveUDPAddr("udp", broadcastIP.String()+":"+serverPort)
		if err != nil {
			fmt.Printf("Failed to resolv bulletin board server address. %s\n", err)
			continue
		}
		fmt.Printf("Bulltien board broadcast address is %s\n", broadcastAddr.String())

		if sender != nil {
			sender.Close()
		}
		sender, err = net.DialUDP("udp", nil, broadcastAddr)
		if err != nil {
			fmt.Printf("Failed to connect bulletin board server. %s\n", err)
			continue
		}

		selfAddr, err := net.ResolveUDPAddr("udp", selfIP.String()+":"+clientPort)
		fmt.Printf("Create listen port %s\n", selfAddr)
		if err != nil {
			fmt.Printf("Failed to resolv self IP address. %s\n", err)
			continue
		}
		if listener != nil {
			listener.Close()
		}
		listener, err = net.ListenUDP("udp", selfAddr)
		if err != nil {
			fmt.Printf("Failed to open self listen port. %s\n", err)
			continue
		}

		order := make([]byte, 4)
		binary.BigEndian.PutUint32(order, announceType.ServerAddress)
		sender.Write(order)

		buffer := make([]byte, bufferByte)
		listener.SetReadDeadline(time.Now().Add(udpTimeout * time.Second))
		length, err := listener.Read(buffer)

		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				fmt.Println("Receive answer timeout.")
				continue
			}
			fmt.Printf("Failed to read message. %s\n", err)
			continue
		}

		if length != 16 {
			fmt.Println("Receive packet but not correct length for answer.")
			continue
		}
		var serverIP net.IP = net.IPv4(
			buffer[length-4],
			buffer[length-3],
			buffer[length-2],
			buffer[length-1])
		err = tasks.ConnectSynergy(serverIP)
		fmt.Println(err)
	}
	if sender != nil {
		sender.Close()
	}
	if listener != nil {
		listener.Close()
	}
}
