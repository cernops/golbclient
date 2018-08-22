package daemon

import (
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
	"encoding/json"
	"github.com/creasty/defaults"
	"gitlab.cern.ch/lb-experts/golbclient/lbalias/utils/runner"
	"strings"
	"regexp"
	"fmt"
)

const daemonCheckCLI = "/bin/netstat"

type Listening struct {
	Port int `default:"22"`
	Protocols[] protocol `default:"[\"any\"]"`
	IPVersions[] ipVersion `default:"[\"any\"]"`
}

// daemonEntry : helper struct used to parse the metric line
type daemonEntry struct {
	protocol[] protocol
	ip[] ipVersion
	port int
}

// protocol : string type used to distinguish between different transport protocol types
type protocol string
const (
	any protocol = "any"
	tcp = "tcp"
	udp = "udp"
	all = "all"
)

// ipLevel : int type used to distinguish between different IP levels
type ipVersion string
const (
	ipv4 ipVersion = ""
	ipv6 = "6"
)

func (daemon Listening) Run(args ...interface{}) interface{} {
	entry := &daemonEntry{}
	if err := defaults.Set(entry); err != nil {
		logger.Error("An error was detected when attempting to read the default values of a daemon check entry type. Error [%s]", err.Error())
		return false
	}

	if daemon.Port != 0 {
		// Log
		logger.Trace("Legacy check type detected, running the daemon check with a default port [%d], protocol [%v] and ip version [%v]", daemon.Port, daemon.Protocols, daemon.IPVersions)
		// Backwards compatibility with v1.0
		*entry = daemonEntry{protocol: daemon.Protocols, ip: daemon.IPVersions, port: daemon.Port}
	} else {
		metric := args[0].(string)
		// Log
		logger.Trace("Processing daemon check on metric line [%s]", metric)
		// Process the daemon metric & abort if an error was detected
		if daemon.processDaemonMetric(metric, entry) != nil {
			return false
		}
	}

	// Check if the given port is within bounds
	if (entry.port < 1) || (entry.port > 65535) {
		logger.Error("The specified port is out of range [1-65535]")
		return false
	}

	// Check if there is anything listening
	return daemon.isListening(entry.protocol, entry.ip, entry.port)
}

// isListening : checks if a daemon is listening on the given protocol(s) in the selected IP level and port
func (daemon Listening) processDaemonMetric(metric string, entry *daemonEntry) error {
	metric = regexp.MustCompile("{(.*?)}").FindString(metric)
	// Unmarchal JSON into struct
	return json.Unmarshal([]byte(metric), entry)
}

// isListening : checks if a daemon is listening on the given protocol(s) in the selected IP level and port
func (daemon Listening) isListening(protocols[] protocol, ipVersions[] ipVersion, port int) bool {
	output, err := runner.RunCommand(daemonCheckCLI, true, true, "l", "u", "n", "t", "a", "p")
	if err != nil {
		logger.Error("An error was detected when attempting to run the daemon check cli. Error [%s]", err.Error())
		return false
	}

	// For each of the output lines, check protocols, ipVersions and port
	for protocol := range protocols {
		for ipVersion := ipVersions {
			
			//var prot protocol
			//if protocol == "ipv4" {
			//	prot = ipv4
			//} else {
			//	prot = ipv6
			//}
			//

			// regexp.MustCompile(fmt.Printf("(?i)(%s)([ ]+[0-9]+[ ]+[0-9]+[ ]+[0-9.]{8,})([:][%d])", )).
		}
	}


	return true
}