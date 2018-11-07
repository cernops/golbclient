package network

import (
	"net"
)

// GetAllLocalIPAddresses : fetches all unicast-type interface addresses from the local machine
func GetAllLocalIPAddresses() ([]string, error) {
	addrs, err := net.InterfaceAddrs()
	res := make([]string, len(addrs))
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		res = append(res, addr.String())
	}

	return res, nil
}
