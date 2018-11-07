package network

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"gitlab.cern.ch/lb-experts/golbclient/lbalias/utils/runner"
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

// NewConcurrentPortScanner : create a new instance of the type *concurrentPortScanner based on the params @see Params
//	given
func NewConcurrentPortScanner(params Params) *concurrentPortScanner {
	cpsInstance := new(concurrentPortScanner)
	psInstances := make([]portScanner, len(params.Hosts))
	ulimit := ulimit() / int64(len(params.Hosts))

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

func (cps *concurrentPortScanner) Start(timeout time.Duration) {
	for _, instance := range *cps.instances {
		instance.start(timeout)
	}
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
