package daemon

import (
	"encoding/json"
	"fmt"
	"github.com/creasty/defaults"
	"github.com/google/go-cmp/cmp"
	"gitlab.cern.ch/lb-experts/golbclient/lbalias/utils/runner"
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
	"regexp"
	"strconv"
	"strings"
)

const daemonCheckCLI = "/bin/netstat"

type Listening struct {
	Port      []Port      `default:"[]"`
	Protocol  []Protocol  `default:"[\"tcp\", \"udp\"]"`
	IPVersion []IPVersion `default:"[\"ipv4\", \"ipv6\"]"`
	Host      []Host      `default:"[]"`
}

// Helper structs
type helperEntry struct {
	Port      []Port
	Protocol  []Protocol
	IPVersion []IPVersion
	Host      []Host
}

// defaultCheck : reusable container variable for the default struct
var defaultCheck *Listening

// port : int type to represent a port number
type Port int

// protocol : string type used to distinguish between different transport protocol types
type Protocol string

// ipLevel : int type used to distinguish between different IP levels
type IPVersion string

const (
	ipv4 IPVersion = ""
	ipv6           = "6"
)

// host : string type used to identify in which host the daemon should be listening on
type Host string

const (
	localhost Host = "127.0.0.1"
)

func (daemon Listening) Run(args ...interface{}) interface{} {
	if err := defaults.Set(&daemon); err != nil {
		logger.Error("An error was detected when attempting to read the default values of a daemon check entry type. Error [%s]", err.Error())
		return false
	}

	// Assign default variable
	defaultCheck = &daemon

	// Fetch all listening interfaces
	res := daemon.fetchAllLocalInterfaces()
	if !res {
		return false
	}

	// Log
	logger.Debug("Loaded default daemon values [%v]", daemon)

	metric := args[0].(string)
	// Log
	logger.Trace("Processing daemon check on metric line [%s]", metric)
	// Process the daemon metric & abort if an error was detected
	if daemon.processDaemonMetric(metric) != nil {
		return false
	}

	for _, port := range daemon.Port {
		// Check if the given port is within bounds
		if (port < 1) || (port > 65535) {
			logger.Error("The specified port [%d] is out of range [1-65535]", port)
			return false
		}
	}

	// Check if there is anything listening
	return daemon.isListening()
}

// Override the default `UnmarshalJSON` from the json package
func (daemon *Listening) UnmarshalJSON(b []byte) error {
	var resultingErr error
	defer func() {
		if r := recover(); r != nil {
			resultingErr = r.(error)
		}
	}()

	fetched := regexp.MustCompile(`(["][\w]+["][ ]*[:])([ ]*[\["]?[\w,]+([ ]*?[\w.]+)?[]"]?)`).FindAllString(string(b), -1)
	for _, line := range fetched {
		p := strings.TrimSpace(strings.Split(line, ":")[1])
		if strings.HasPrefix(p, `"`) {
			p = strings.Replace(p, `"`, ``, -1)
		}

		var helper helperEntry
		if strings.Contains(line, "port") {
			if !strings.HasPrefix(p, "[") {
				logger.Trace("Extracting [Port] single-value to array-value in [%s]...", p)
				value, err := strconv.Atoi(strings.Split(p, ",")[0])
				if err != nil {
					return err
				}
				daemon.Port = []Port{Port(value)}
			} else {
				logger.Trace("Extracting [Port] array-value in [%s]...", p)
				err := json.Unmarshal([]byte(fmt.Sprintf("{%s}", line)), &helper)
				if err != nil {
					return err
				}
				daemon.Port = helper.Port
			}
		} else if strings.Contains(line, "protocol") {
			if !strings.HasPrefix(p, "[") {
				logger.Trace("Extracting [Protocol] single-value to array-value in [%s]...", p)
				daemon.Protocol = []Protocol{Protocol(p)}
			} else {
				logger.Trace("Extracting [Protocol] array-value in [%s]...", p)
				err := json.Unmarshal([]byte(fmt.Sprintf("{%s}", line)), &helper)
				if err != nil {
					return err
				}
				daemon.Port = helper.Port
			}
		} else if strings.Contains(line, "ip") {
			if !strings.HasPrefix(p, "[") {
				logger.Trace("Extracting [IPVersion] single-value to array-value in [%s]...", p)
				daemon.IPVersion = []IPVersion{IPVersion(p)}
			} else {
				logger.Trace("Extracting [IPVersion] array-value in [%s]...", p)
				err := json.Unmarshal([]byte(fmt.Sprintf("{%s}", line)), &helper)
				if err != nil {
					return err
				}
				daemon.IPVersion = helper.IPVersion
			}
		} else if strings.Contains(line, "host") {
			if !strings.HasPrefix(p, "[") {
				logger.Trace("Extracting [Host] single-value to array-value in [%s]...", p)
				daemon.Host = []Host{Host(p)}
			} else {
				logger.Trace("Extracting [Host] array-value in [%s]...", p)
				err := json.Unmarshal([]byte(fmt.Sprintf("{%s}", line)), &helper)
				if err != nil {
					return err
				}
				daemon.Host = helper.Host
			}
		}

	}
	return resultingErr
}

// isListening : checks if a daemon is listening on the given protocol(s) in the selected IP level and port
func (daemon *Listening) processDaemonMetric(metric string) error {
	metric = regexp.MustCompile("{(.*?)}").FindString(metric)

	// Unmarshal JSON into struct
	if err := json.Unmarshal([]byte(metric), &daemon); err != nil {
		logger.Error("Error when parsing json [%s]. Error [%s]", metric, err.Error())
		return err
	}

	logger.Trace("Finished processing metric file [%v]", daemon)
	return nil
}

// fetchAllLocalInterfaces : Fetch all local interfaces IPs and add them to the default array of hosts to check
func (daemon *Listening) fetchAllLocalInterfaces() bool {
	// Log
	logger.Trace("Fetching all IPs from the local interfaces")

	// Retrieve all interfaces on the machine by default
	output, err := runner.RunCommand(`ifconfig`, true, true)
	if err != nil {
		logger.Error("Failed to fetch the interfaces from this machine. Error [%s]", err.Error())
		return false
	}
	outputIPs := regexp.MustCompile(`inet[0-9]?[ ][\w.:]*`).FindAllString(output, -1)
	logger.Trace("Found local addresses [%v]", outputIPs)
	for _, ip := range outputIPs {
		daemon.Host = append(daemon.Host, Host(strings.Split(ip, " addr:")[1]))
	}

	// Add the any interface static IP pattern
	daemon.Host = append(daemon.Host, Host("0.0.0.0"))

	return true
}

// isListening : checks if a daemon is listening on the given protocol(s) in the selected IP level and port
func (daemon *Listening) isListening() bool {
	// The port metric argument is mandatory
	if len(daemon.Port) == 0 {
		logger.Error("A port needs to be specified in a daemon check in the format `{port : <val>}`. Aborting check...")
		return false
	}

	// Run the cli
	output, err := runner.RunCommand(daemonCheckCLI, true, true, "-l", "-u", "-n", "-t", "-a", "-p")
	if err != nil {
		logger.Error("An error was detected when attempting to run the daemon check cli. Error [%s]", err.Error())
		return false
	}

	// Detect if the default struct values were changed
	needAny := cmp.Equal(daemon.Host, defaultCheck.Host) || cmp.Equal(daemon.IPVersion, defaultCheck.IPVersion) || cmp.Equal(daemon.Protocol, defaultCheck.Protocol)
	logger.Trace("Daemon check need any condition :: [%t]", needAny)

	found := false

	// For each of the output lines, check protocols, ipVersions and port
	for _, h := range daemon.Host {
		logger.Trace("Checking for host [%s]", h)
		if h == "localhost" {
			h = localhost
		}
		for _, p := range daemon.Protocol {
			for _, pt := range daemon.Port {
				for i, ip := range daemon.IPVersion {
					if ip == "ipv4" {
						ip = ipv4
					} else {
						ip = ipv6
					}

					expression := fmt.Sprintf(`(?i)(%s%s)([ ]+[0-9]+[ ]+[0-9]+[ ]+(%s))([:](%d))(.*)(LISTEN)`, p, ip, h, pt)
					logger.Trace("Checking if daemon is listening with expression [%s]", expression)
					if !regexp.MustCompile(expression).MatchString(output) {
						// Are all needed?
						if !needAny {
							logger.Debug("No daemon is listening on port [%d], IP version [%s] and transport protocol [%s]", pt, daemon.IPVersion[i], p)
							// Fail on first failed check
							return found
						}
					} else {
						found = true
					}
				}
			}
		}
	}
	// All has passed
	return found
}
