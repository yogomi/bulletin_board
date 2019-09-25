package tasks

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
)

// ConnectSynergy は指定されたアドレスへsynergyのConnectを行う。
func ConnectSynergy(synergyServer net.IP) error {
	var options = []string{"-1"}
	if runtime.GOOS != "windows" {
		options = append(options, "-f")
	}
	options = append(options, synergyServer.String())

	fmt.Println("Connect to synergy serve %s", synergyServer)
	err := exec.Command("synergyc", options...).Run()
	return err
}
