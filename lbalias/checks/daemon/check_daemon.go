package daemon

import (
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
	"encoding/json"
	"github.com/creasty/defaults"
	"gitlab.cern.ch/lb-experts/golbclient/lbalias/utils/runner"
	"regexp"
	"strconv"
	"strings"
	"fmt"
)

const daemonCheckCLI = "/bin/netstat"

type Listening struct {
	Port []port `default:"[22]"`
	Protocol []protocol `default:"[]"`
	IPVersion []ipVersion `default:"[]"`
}

// Helper structs
type helperEntry struct {
	Port []port
	Protocol []protocol
	IPVersion []ipVersion
	Host []host
}

// port : int type to represent a port number
type port int

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

// host : string type used to identify in which host the daemon should be listening on
type host string

// Override the default `UnmarshalJSON` from the json package
func (daemon *Listening) UnmarshalJSON(b []byte) error {
	fetched := regexp.MustCompile(`(["][\w]+["][ ]*[:])([ ]*[\["]?[\w,]+([ ]*?[\w]+)?[]"]?)`).FindAllString(string(b), -1)
	for _, line := range fetched {
		p := strings.TrimSpace(strings.Split(line, ":")[1])
		if strings.HasPrefix(p, `"`){
			p = strings.Replace(p, `"`, ``, -1)
		}
		
		var helper helperEntry
		if strings.Contains(line, "port")  {
			if !strings.HasPrefix(p, "[") {
				value, err := strconv.Atoi(p)
				if err != nil {
					return err
				}
				(*daemon).Port = []port{port(value)}
			} else {
				
				err := json.Unmarshal([]byte(fmt.Sprintf("{%s}", line)), &helper)
				if err != nil {
					return err
				}
				(*daemon).Port = helper.Port
			}
		} else if strings.Contains(line, "protocol") {
			if !strings.HasPrefix(p, "[") {
				(*daemon).Protocol = []protocol{protocol(p)}
			} else {
				err := json.Unmarshal([]byte(fmt.Sprintf("{%s}", line)), &helper)
				if err != nil {
					return err
				}
				(*daemon).Port = helper.Port
			}
		} else if strings.Contains(line, "ip") {
			if !strings.HasPrefix(p, "[") {
				(*daemon).IPVersion = []ipVersion{ipVersion(p)}
			} else {
				err := json.Unmarshal([]byte(fmt.Sprintf("{%s}", line)), &helper)
				if err != nil {
					return err
				}
				(*daemon).IPVersion = helper.IPVersion
			}
		}
	}
	return nil
}



func (daemon Listening) Run(args ...interface{}) interface{} {
	if err := defaults.Set(daemon); err != nil {
		logger.Error("An error was detected when attempting to read the default values of a daemon check entry type. Error [%s]", err.Error())
		return false
	}

	if len(daemon.Port) != 0 {
		// Log
		logger.Trace("Legacy check type detected, running the daemon check with a default port [%d], protocol [%v] and ip version [%v]", daemon.Port, daemon.Protocols, daemon.IPVersions)
	} else {
		metric := args[0].(string)
		// Log
		logger.Trace("Processing daemon check on metric line [%s]", metric)
		// Process the daemon metric & abort if an error was detected
		if daemon.processDaemonMetric(metric) != nil {
			return false
		}
	}

	for port := range daemon.Port {
		// Check if the given port is within bounds
		if (port < 1) || (port > 65535) {
			logger.Error("The specified port is out of range [1-65535]")
			return false
		}
	}

	// Check if there is anything listening
	return daemon.isListening()
}

// isListening : checks if a daemon is listening on the given protocol(s) in the selected IP level and port
func (daemon Listening) processDaemonMetric(metric string) error {
	metric = regexp.MustCompile("{(.*?)}").FindString(metric)

	// Unmarshal JSON into struct
	if err := json.Unmarshal([]byte(metric), daemon); err != nil {
		logger.Error("Error when parsing json [%s]. Error [%s]", metric, err.Error())
		return err
	}
	return nil
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