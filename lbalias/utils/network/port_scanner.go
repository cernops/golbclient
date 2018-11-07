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

func (ps *portScanner) Start(timeout time.Duration) {
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
	// Process protocol
	if len(protocol) != 0 {
		launchScan(ip, protocol, ipVersion, port, timeout)
	} else {
		waitGroup := sync.WaitGroup{}
		waitGroup.Add(2)
		go func() {
			defer waitGroup.Done()
			launchScan(ip, "tcp", ipVersion, port, timeout)
		}()
		go func() {
			defer waitGroup.Done()
			launchScan(ip, "udp", ipVersion, port, timeout)
		}()
		waitGroup.Wait()
	}
}

func launchScan(ip string, protocol string, ipVersion string, port int, timeout time.Duration) {
	// Process IPVersion
	if ipVersion == "ipv4" {
		ipVersion = "4"
	} else if ipVersion == "ipv6" {
		ipVersion = "6"
	}

	target := fmt.Sprintf("%s:%d", ip, port)
	protocolIPv := fmt.Sprintf("%s%s", protocol, ipVersion)

	conn, err := net.DialTimeout(protocolIPv, target, timeout)
	if err != nil {
		if strings.Contains(err.Error(), "too many open files") {
			time.Sleep(timeout)
			scanPort(ip, protocol, ipVersion, port, timeout)
		} else {
			logger.Info("The port [%d], on the protocol [%s] :: target [%s], is closed.", port, protocolIPv, target)
		}
		return
	}
	conn.Close()
	logger.Info("The port [%d], on the protocol [%s] :: target [%s], is open.", port, protocolIPv, target)
}
