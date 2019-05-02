package network

import (
	"fmt"
	"math/big"
	"net"
	"strings"
)

func GetPackedReprFromIP(ipAddress string) (string, error) {
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return "", fmt.Errorf("the given string [%s] is not a valid IP address", ipAddress)
	}

	ipv4 := false
	if ip.To4() != nil {
		ipv4 = true
	}

	ipInt := big.NewInt(0)
	if ipv4 {
		ipInt.SetBytes(ip.To4())
		ipHex := strings.ToUpper(fmt.Sprintf("%08x", reverseBytesSlice(ipInt.Bytes())))
		return ipHex, nil
	}

	ipInt.SetBytes(ip.To16())
	ipHex := strings.ToUpper(fmt.Sprintf("%032x", reverseBytesSlice(ipInt.Bytes())))
	return ipHex, nil
}

func reverseBytesSlice(slice []byte) []byte {
	if len(slice) == 0 {
		return slice
	}
	return append(reverseBytesSlice(slice[1:]), slice[0])
}