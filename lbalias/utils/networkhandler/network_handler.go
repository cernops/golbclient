package networkhandler

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"gitlab.cern.ch/lb-experts/golbclient/lbalias/utils/runner"
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
	"golang.org/x/sync/semaphore"
)

type portScanner struct {
	ipLookup       string
	protocolLookup string
	lock           *semaphore.Weighted
}

func NewPortScanner(ip string, protocol string) *portScanner {
	return &portScanner{
		ipLookup:       ip,
		protocolLookup: protocol,
		lock:           semaphore.NewWeighted(ulimit()),
	}
}

func (ps *portScanner) Start(start, end int, timeout time.Duration) {
	waitGroup := sync.WaitGroup{}
	defer waitGroup.Wait()

	for port := start; port <= end; port++ {
		waitGroup.Add(1)
		ps.lock.Acquire(context.TODO(), 1)

		go func(port int) {
			defer ps.lock.Release(1)
			defer waitGroup.Done()
			scanPort(ps.ipLookup, ps.protocolLookup, port, timeout)
		}(port)
	}
}

func scanPort(ip string, protocol string, port int, timeout time.Duration) {
	target := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout(protocol, target, timeout)
	defer conn.Close()
	if err != nil {
		if strings.Contains(err.Error(), "too many open files") {
			time.Sleep(timeout)
			scanPort(ip, protocol, port, timeout)
		} else {
			logger.Trace("The port [%d], on the protocol [%s], is closed.", port, protocol)
		}
		return
	}
	logger.Trace("The port [%d], on the protocol [%s], is open.", port, protocol)
}

func ulimit() (maxOpenFiles int64) {
	out, err := runner.RunCommand("ulimit", true, true, "-n")
	if err != nil {
		panic(fmt.Sprintf("An unexpected error ocurred when running [ulimit]. Error [%s]", err.Error()))
	}

	s := strings.TrimSpace(out)
	maxOpenFiles, err = strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("An unexpected error ocurred when parsing the [ulimt] output. Error [%s]", err.Error()))
	}

	return
}
