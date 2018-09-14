package daemon

import (
	"encoding/json"
	"fmt"
	"github.com/creasty/defaults"
	"github.com/google/go-cmp/cmp"
	"gitlab.cern.ch/lb-experts/golbclient/lbalias/utils/benchmarker"
	"gitlab.cern.ch/lb-experts/golbclient/lbalias/utils/runner"
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const daemonCheckCLI = "/bin/netstat"

type Listening struct {
	Port      []Port      `default:"[]"`
	Protocol  []Protocol  `default:"[\"tcp\", \"udp\"]"`
	IPVersion []IPVersion `default:"[\"ipv4\", \"ipv6\"]"`
	// The host array is fetched at runtime
	Host      []Host      `default:"[]"`
}

// Helper struct
type daemonJsonContainer struct {
	PortRaw interface{} `json:"port"`
	Protocol  interface{} `json:"protocol"`
	IPVersion interface{} `json:"ip"`
	Host      interface{} `json:"host"`
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

// parseDaemonJSON : parse a given json metric line into the expected schema
func (daemon *Listening) parseDaemonJSON(line string) (err error) {
	// Account for parsing errors
	defer func() {
		if r := recover(); r != nil {
			if re, ok := r.(error); re != nil && ok {
				err = re
			}
		}
	}()

	// Attempt to parse the JSON text
	x := new(daemonJsonContainer)
	if err := json.NewDecoder(strings.NewReader(line)).Decode(x); err != nil {
		return err
	}

	// Detect duplicated keys
	benchmarker.TimeItV(validateUniqueKeys, time.Nanosecond, line)
	//validateUniqueKeys(line)

	// Reject wrong data-types
	if !validateDataTypes(x) {
		return fmt.Errorf("")
	}

	// Parse :: Port
	port0, ok := x.PortRaw.(interface{})
	if ok {
		s, isString := port0.(string)
		if isString {
			r, err := strconv.Atoi(s)
			if err != nil {
				return err
			}
			daemon.Port = append(daemon.Port , Port(r))
		}
		i, isFloat := port0.(float64)
		if isFloat {
			daemon.Port = []Port{Port(i)}
		}
	}
	port1, ok := x.PortRaw.([]interface{})
	if ok {
		for _, p := range port1 {
			s, isString := p.(string)
			if isString {
				r, err := strconv.Atoi(s)
				if err != nil {
					return err
				}
				daemon.Port = append(daemon.Port , Port(r))
			}
			i, isFloat := p.(float64)
			if isFloat {
				daemon.Port = append(daemon.Port , Port(i))
			}
		}
	}

	// Parse :: Protocol
	protocol0, ok := x.Protocol.(interface{})
	if ok {
		s, isString := protocol0.(string)
		if isString {
			daemon.Protocol = []Protocol{Protocol(s)}
		}
	}
	protocol1, ok := x.Protocol.([]interface{})
	if ok {
		for _, p := range protocol1 {
			s, isString := p.(string)
			if isString {
				daemon.Protocol = append(daemon.Protocol , Protocol(s))
			}
		}
	}

	// Parse :: IP version
	ipV0, ok := x.IPVersion.(interface{})
	if x.IPVersion != nil {
		daemon.IPVersion = daemon.IPVersion[:0]
	}
	if ok {
		s, isString := ipV0.(string)
		if isString {
			daemon.IPVersion = []IPVersion{IPVersion(s)}
		}
	}
	ipV1, ok := x.IPVersion.([]interface{})
	if ok {
		for _, p := range ipV1 {
			s, isString := p.(string)
			if isString {
				daemon.IPVersion = append(daemon.IPVersion , IPVersion(s))
			}
		}
	}

	// Parse :: Host
	host0, ok := x.Host.(interface{})
	if x.Host != nil {
		daemon.Host = daemon.Host[:0]
	}
	if ok {
		s, isString := host0.(string)
		if isString {
			daemon.Host = []Host{Host(s)}
		}
	}
	host1, ok := x.Host.([]interface{})
	if ok {
		for _, p := range host1 {
			s, isString := p.(string)
			if isString {
				daemon.Host = append(daemon.Host , Host(s))
			}
		}
	}
	return err
}

// validateUniqueKeys : checks if more than one entry for the same key was detected. If so, present a warning message to the user
func validateUniqueKeys(line interface{}) {
	foundKeys := regexp.MustCompile(`("\w+" *:)`).FindAllString(line.(string), -1)
	foundKeyMap := make(map[string]bool, len(foundKeys))
	for _, key := range foundKeys {
		key := strings.Replace(key, " ", "", -1)
		keyA := []rune(key)
		if _, exists := foundKeyMap[key]; exists {
			logger.Warn("The key [%s] was found multiple times. Note that only the last declared key-value pair will be used.", string(keyA[1:len(keyA)-2]))
		} else {
			foundKeyMap[key] = true
		}
	}
}

// validateDataTypes : type-checks that the given data-types in JSON are supported
func validateDataTypes(x *daemonJsonContainer) (b bool) {
	// Protocol
	_, wrongProtocolType := x.Protocol.(float64)
	if wrongProtocolType {
		logger.Warn("Wrong data-type given to the `protocol` key for the daemon check. Only <string> or <string_array> types are supported.")
		return
	}
	// IP version
	_, wrongIPVersionType := x.IPVersion.(float64)
	if wrongIPVersionType {
		logger.Warn("Wrong data-type given to the `ip` key for the daemon check. Only <string> or <string_array> types are supported.")
		return
	}
	// Host
	_, wrongHostType := x.Host.(float64)
	if wrongHostType {
		logger.Warn("Wrong data-type given to the `host` key for the daemon check. Only <string> or <string_array> types are supported.")
		return
	}
	return true
}

// isListening : checks if a daemon is listening on the given protocol(s) in the selected IP level and port
func (daemon *Listening) processDaemonMetric(metric string) error {
	metric = regexp.MustCompile("{(.*?)}").FindString(metric)

	// Parse json
	err := daemon.parseDaemonJSON(metric)
	if err != nil {
		logger.Error("Error when decoding metric [%s]. Error [%s]. Failing metric...", metric, err.Error())
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
	output, err := runner.RunCommand(`/usr/sbin/ifconfig`, true, true)
	if err != nil {
		logger.Error("Failed to fetch the interfaces from this machine. Error [%s]", err.Error())
		return false
	}
	outputIPs := regexp.MustCompile(`inet[0-9]?[ ][\w.:]*`).FindAllString(output, -1)
	logger.Trace("Found local addresses [%v]", outputIPs)
	for _, ip := range outputIPs {
		daemon.Host = append(daemon.Host, Host(regexp.MustCompile("inet([6])?[ ]").Split(ip, -1)[1]))
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
	res, err := runner.RunCommand(daemonCheckCLI, true, true, "-l", "-u", "-n", "-t", "-a", "-p")
	if err != nil {
		logger.Error("An error was detected when attempting to run the daemon check cli. Error [%s]", err.Error())
		return false
	}

	// Detect if the default struct values were changed
	needAny := cmp.Equal(daemon.Host, defaultCheck.Host) || cmp.Equal(daemon.IPVersion, defaultCheck.IPVersion) || cmp.Equal(daemon.Protocol, defaultCheck.Protocol)
	logger.Trace("Daemon check need any condition :: [%t]. Daemon entry [%v]", needAny)

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
					if !regexp.MustCompile(expression).MatchString(res) {
						logger.Trace(`Unable to find daemon for {"host": [%s], "protocol": [%s], "ip": [%s], "port":[%v]`, h, p, ip, pt)
						// Are all needed?
						if !needAny {
							logger.Debug("No daemon is listening on port [%d], IP version [%s] and transport protocol [%s]", pt, daemon.IPVersion[i], p)
							// Fail on first failed check
							return found
						}
					} else {
						logger.Trace(`Found daemon for {"host": [%s], "protocol": [%s], "ip": [%s], "port":[%v]}`, h, p, ip, pt)
						found = true
					}
				}
			}
		}
	}
	// All has passed
	return found
}
