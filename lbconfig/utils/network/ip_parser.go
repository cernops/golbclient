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

	if ipAddress == "::1" {
		return "00000000000000000000000001000000", nil
	}

	ipInt := big.NewInt(0)
	if ip.To4() != nil {
		ipInt.SetBytes(ip.To4())
		ipHex := strings.ToUpper(fmt.Sprintf("%08x", reverseBytesSlice(ipInt.Bytes())))
		return ipHex, nil
	}

	ipInt.SetBytes(ip.To16())
	ipHex := strings.ToUpper(fmt.Sprintf("%032x", reverseBytesSlice(ipInt.Bytes())))
	return ipHex, nil
}

func reverseBytesSlice(s []byte) []byte {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}