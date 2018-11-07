package network

import (
	"fmt"
	"sync"
	"syscall"
	"time"
)

type concurrentPortScanner struct {
	instances *[]portScanner
}

// Params : struct responsible for supplying a named-params-like functionality to the creation of a new
// 	concurrentPortScanner instance
type Params struct {
	Hosts      []string
	Protocols  []string
	IPVersions []string
	Ports      []int
}

func (p *Params) applyDefaultValues() {
	enforceNonEmpty(&p.Hosts)
	enforceNonEmpty(&p.Protocols)
	enforceNonEmpty(&p.IPVersions)
}

func enforceNonEmpty(sarr *[]string) {
	if len(*sarr) == 0 {
		*sarr = []string{""}
	}
}

// NewConcurrentPortScanner : create a new instance of the type *ConcurrentPortScanner based on the params @see Params
//	given
func NewConcurrentPortScanner(params Params) *concurrentPortScanner {
	params.applyDefaultValues()

	cpsInstance := new(concurrentPortScanner)
	psInstances := make([]portScanner, len(params.Hosts))
	ulimit := (ulimit() / int64(len(params.Hosts)) * int64(len(params.Protocols)) *
		int64(len(params.IPVersions))) - int64(len(params.Hosts))

	for _, host := range params.Hosts {
		for _, protocol := range params.Protocols {
			for _, ipVersion := range params.IPVersions {
				psInstances = append(
					psInstances,
					*newPortScanner(params.Ports, host, protocol, ipVersion, ulimit))
			}
		}
	}

	cpsInstance.instances = &psInstances
	return cpsInstance
}

// Run : runs the ConcurrentPortScanner against the specified params @see newConcurrentPortScanner
func (cps *concurrentPortScanner) Run(timeout time.Duration) {
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(len(*cps.instances))

	for _, instance := range *cps.instances {
		go func(i portScanner) {
			defer waitGroup.Done()
			i.Start(timeout)
		}(instance)
	}

	waitGroup.Wait()
}

func ulimit() int64 {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		panic(fmt.Sprintf("An unexpected error ocurred when running [ulimit]. Error [%s]", err.Error()))
	}

	return int64(rLimit.Cur)
}
