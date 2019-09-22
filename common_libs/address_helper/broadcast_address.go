package addressHelper

import (
	"encoding/binary"
	"errors"
	"net"
)

// GetIPv4BroadcastAddressFromAddressList はnet.Addrのリストを取得し、その中から
// IPv4アドレスを確認したらそのブロードキャストアドレスを作成して返します。
func GetIPv4BroadcastAddressFromAddressList(addressList []net.Addr) (net.IP, error) {
	for _, addr := range addressList {
		selfIP, networkIP, _ := net.ParseCIDR(addr.String())
		if selfIP.To4() != nil {
			broadcastIPBuff := make(net.IP, len(networkIP.IP.To4()))
			binary.BigEndian.PutUint32(
				broadcastIPBuff,
				binary.BigEndian.Uint32(
					networkIP.IP.To4())|^binary.BigEndian.Uint32(net.IP(networkIP.Mask).To4()))
			return broadcastIPBuff, nil
		}
	}
	return nil, errors.New("Cannot find IPv4 address")
}
