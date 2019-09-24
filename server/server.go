package main

import (
	"../common_libs/address_helper"
	"../common/announce_type"
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
	*endFlag = true
}

const intervalTime = 3
const bufferByte = 64
const port = "1235"
const networkDeviceName = "en0"

func main() {
	var endFlag = false
	go WaitSignal(&endFlag)

	var waitingBroadcastIP net.IP
	var listener *net.UDPConn
	var interfaceError = false

	for endFlag == false {
		if interfaceError {
			// interfaceが見つからなかった場合は、通常のintervalの3倍待つ
			time.Sleep(intervalTime * 3 * time.Second)
			interfaceError = false
		}
		networkDevice, err := net.InterfaceByName(networkDeviceName)
		if err != nil {
			waitingBroadcastIP = nil
			interfaceError = true
			fmt.Printf("Device %s is not found.\n", networkDeviceName)
			continue
		}
		addrs, err := networkDevice.Addrs()
		if err != nil {
			waitingBroadcastIP = nil
			interfaceError = true
			fmt.Printf("Failed to find address on device. %s\n", err)
			continue
		}

		broadcastIP, err := addressHelper.GetIPv4BroadcastAddressFromAddressList(addrs)
		if err != nil {
			waitingBroadcastIP = nil
			interfaceError = true
			fmt.Printf("Failed to find address on device. %s\n", err)
			continue
		}

		if waitingBroadcastIP.String() != broadcastIP.String() {
			waitingBroadcastIP = broadcastIP

			broadcastAddr, err := net.ResolveUDPAddr("udp", broadcastIP.String() + ":" + port)
			if err != nil {
				waitingBroadcastIP = nil
				interfaceError = true
				fmt.Printf("Failed to resolv bulletin board server address. %s\n", err)
				continue
			}
			fmt.Println(broadcastIP)

			if listener != nil {
				listener.Close()
			}
			listener, err = net.ListenUDP("udp", broadcastAddr)
			if err != nil {
				waitingBroadcastIP = nil
				interfaceError = true
				fmt.Printf("Failed to open server port. %s\n", err)
				continue
			}
		}

		buffer := make([]byte, bufferByte)
		listener.SetReadDeadline(time.Now().Add(intervalTime * time.Second))
		length, inboundFromAddrByte, err := listener.ReadFrom(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			if listener != nil {
				listener.Close()
			}
			waitingBroadcastIP = nil
			interfaceError = true
			fmt.Printf("Failed to read message. %s\n", err)
		}

		if length == 4 {
			var announceOrder uint32 = binary.BigEndian.Uint32(buffer[:length])
			switch announceOrder {
			case announceType.ServerAddress:
				fmt.Println("Receive order Server Address")
			default:
				fmt.Printf("Unknown announce order. %v\n", buffer[:length])
			}
			inboundFromAddr := inboundFromAddrByte.(*net.UDPAddr).String()
			fmt.Printf("Inbound %v\n", inboundFromAddr)
		}
	}

	if listener != nil {
		listener.Close()
	}
}
