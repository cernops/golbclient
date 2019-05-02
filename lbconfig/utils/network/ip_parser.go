package network

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
	"math/big"
	"net"
)

func ip4toInt(IPv4Address net.IP) int64 {
	IPv4Int := big.NewInt(0)
	IPv4Int.SetBytes(IPv4Address.To4())
	return IPv4Int.Int64()
}

func ip6toInt(IPv6Address net.IP) int64 {
	IPv6Int := big.NewInt(0)
	IPv6Int.SetBytes(IPv6Address.To16())
	return IPv6Int.Int64()
}

func pack32BinaryIP6(ip6Address string) (string, error) {
	ipv6Decimal := ip6toInt(net.ParseIP(ip6Address))

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, uint32(ipv6Decimal))

	if err != nil {
		return "", fmt.Errorf("unable to create a packed binary representation of the given string [%s]" +
			" to IPv6. Error [%s]",	ip6Address, err.Error())
	}

	result := fmt.Sprintf("%x", buf.Bytes())
	return result, nil
}

func pack32BinaryIP4(ip4Address string) (string, error) {
	ipv4Decimal := ip4toInt(net.ParseIP(ip4Address))

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, uint32(ipv4Decimal))

	if err != nil {
		return "", fmt.Errorf("unable to create a packed binary representation of the given string [%s]" +
			" to IPv4. Error [%s]",	ip4Address, err.Error())
	}

	result := fmt.Sprintf("%x", buf.Bytes())
	return result, nil
}

func GetPackedReprFromIP(ipAddress string) (string, error) {
	hexString, err := pack32BinaryIP4(ipAddress)
	if err != nil {
		logger.Debug(err.Error())
		hexString, err = pack32BinaryIP6(ipAddress)
		if err != nil {
			return "", err
		}
	}
	return hexString, nil
}