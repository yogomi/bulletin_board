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
	*endFlag = true
}

const intervalTime = 10
const udpTimeout = 3
const bufferByte = 64

var serverPort = "8120"
var clientPort = "8121"

const networkDeviceName = "en0"

func main() {
	var endFlag = false
	go WaitSignal(&endFlag)

	var waitingBroadcastIP net.IP
	var listener *net.UDPConn
	var interfaceError = false

	for endFlag == false {
		if interfaceError {
			time.Sleep(intervalTime * time.Second)
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

		selfIP, _, broadcastIP, err := addressHelper.GetIPv4AddressSetFromAddressList(addrs)
		if err != nil {
			waitingBroadcastIP = nil
			interfaceError = true
			fmt.Printf("Failed to find address on device. %s\n", err)
			continue
		}

		if waitingBroadcastIP.String() != broadcastIP.String() {
			waitingBroadcastIP = broadcastIP

			broadcastAddr, err := net.ResolveUDPAddr("udp", broadcastIP.String()+":"+serverPort)
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
		listener.SetReadDeadline(time.Now().Add(udpTimeout * time.Second))
		length, inboundFromUDPAddr, err := listener.ReadFrom(buffer)
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
			continue
		}

		if length == 4 {
			var announceOrder uint32 = binary.BigEndian.Uint32(buffer[:length])
			switch announceOrder {
			case announceType.ServerAddress:
				fmt.Println("Receive order Server Address")
				{

					clientTargetAddr, err := net.ResolveUDPAddr(
						"udp",
						inboundFromUDPAddr.(*net.UDPAddr).IP.String()+":"+clientPort)
					fmt.Println(clientTargetAddr)
					sender, err := net.DialUDP("udp", nil, clientTargetAddr)
					if err != nil {
						fmt.Println("Failed to connect client")
						break
					}
					defer sender.Close()
					sendedLength, err := sender.Write(selfIP)
					if err != nil {
						fmt.Println("Failed to send Server Address to client")
						break
					}
					fmt.Printf("%d length is sended.", sendedLength)
				}
			default:
				fmt.Printf("Unknown announce order. %v\n", buffer[:length])
			}
			inboundFromAddr := inboundFromUDPAddr.String()
			fmt.Printf("Inbound %v\n", inboundFromAddr)
		}
	}

	if listener != nil {
		listener.Close()
	}
}
