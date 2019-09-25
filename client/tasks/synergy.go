package tasks

import (
	"errors"
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
	result, _ := exec.Command("synergyc", options...).Output()
	return errors.New("Synergyc result is \"" + string(result) + "\".")
}
