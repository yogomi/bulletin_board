package main

import (
	"../common/announce_type"
	"../common_libs/address_helper"
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

const intervalTime = 10
const udpTimeout = 3
const bufferByte = 64
var port = "8120"
var networkDeviceName = "ローカル エリア接続* 4"

func main() {
	var firstInterval = true
	var watchingBroadcastIP net.IP
	var sender *net.UDPConn
	var listener *net.UDPConn
	var endFlag = false

	go WaitSignal(&endFlag)

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

		selfIP, _, broadcastIP, err := addressHelper.GetIPv4AddressSetFromAddressList(addrs)
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
			fmt.Printf("Bulltien board broadcast address is %s\n", broadcastAddr.String())

			if sender != nil {
				sender.Close()
			}
			if listener != nil {
				listener.Close()
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

		// listenerは存在しなければここで作成
		if listener == nil {
			selfAddr, err := net.ResolveUDPAddr("udp", selfIP.String()+":"+port)
			if err != nil {
				fmt.Printf("Failed to resolv self IP address. %s\n", err)
				continue
			}
			listener, err = net.ListenUDP("udp", selfAddr)
			if err != nil {
				fmt.Printf("Failed to open self listen port. %s\n", err)
				continue
			}
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
			if listener != nil {
				listener.Close()
			}
			fmt.Printf("Failed to read message. %s\n", err)
			continue
		}

		if length != 16 {
			fmt.Println("Receive packet but not correct length for answer.")
		}
		var serverIP net.IP = net.IPv4(
			buffer[length-4],
			buffer[length-3],
			buffer[length-2],
			buffer[length-1])
		fmt.Println(serverIP)
	}
	if sender != nil {
		sender.Close()
	}
	if listener != nil {
		listener.Close()
	}
}
