package network

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
	"golang.org/x/sync/semaphore"
)

type portScanner struct {
	portsLookup     []int
	ipLookup        string
	protocolLookup  string
	ipVersionLookup string
	lock            *semaphore.Weighted
}

func newPortScanner(ports []int, hostIP string, protocol string, ipVersion string, limit int64) *portScanner {
	return &portScanner{
		portsLookup:     ports,
		ipLookup:        hostIP,
		protocolLookup:  protocol,
		ipVersionLookup: ipVersion,
		lock:            semaphore.NewWeighted(limit),
	}
}

func (ps *portScanner) start(timeout time.Duration) {
	waitGroup := sync.WaitGroup{}
	defer waitGroup.Wait()

	for _, port := range ps.portsLookup {
		waitGroup.Add(1)
		ps.lock.Acquire(context.TODO(), 1)

		go func(port int) {
			defer ps.lock.Release(1)
			defer waitGroup.Done()
			scanPort(ps.ipLookup, ps.protocolLookup, ps.ipVersionLookup, port, timeout)
		}(port)
	}
}

func scanPort(ip string, protocol string, ipVersion string, port int, timeout time.Duration) {
	target := fmt.Sprintf("%s:%d", ip, port)
	protocolIPv := fmt.Sprintf("%s%s", protocol, ipVersion)
	conn, err := net.DialTimeout(protocolIPv, target, timeout)
	defer conn.Close()
	if err != nil {
		if strings.Contains(err.Error(), "too many open files") {
			time.Sleep(timeout)
			scanPort(ip, protocol, ipVersion, port, timeout)
		} else {
			logger.Trace("The port [%d], on the protocol [%s], is closed.", port, protocol)
		}
		return
	}
	logger.Trace("The port [%d], on the protocol [%s], is open.", port, protocol)
}
