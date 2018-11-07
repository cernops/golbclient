package checks

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"gitlab.cern.ch/lb-experts/golbclient/lbalias/utils/runner"
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
)

const daemonCheckCLI = `/bin/netstat -luntp`

// port : int alias-type to represent a port number
type port = int

// protocol : string alias-type used to distinguish between different transport protocol types
type protocol = string

// ipVersion : int alias-type used to distinguish between different IP levels
type ipVersion = string

// host : string alias-type used to identify in which host the daemon should be listening on
type host = string

// Listening : struct responsible for all the daemon check slices
type Listening struct {
	ports      []port
	protocols  []protocol
	ipVersions []ipVersion
	hosts      []host
	// Backwards compatibility
	Metric string
}

// daemonJsonContainer : Helper struct
type daemonJSONContainer struct {
	PortRaw   interface{} `json:"port"`
	Protocol  interface{} `json:"protocol"`
	IPVersion interface{} `json:"ip"`
	Host      interface{} `json:"host"`
}

// listening_default_values_map : map that is used to set the default values for
// the ones not found in the metric line
var defaultProtocols = []protocol{"tcp", "udp"}
var defaultIPVersions = []ipVersion{"ipv4", "ipv6"}

// Run : expects that metric line (in string format) to be given in the first position of the
//	arguments
func (daemon Listening) Run(args ...interface{}) interface{} {
	metric := args[0].(string)

	// Log
	logger.Trace("Processing daemon check on metric line [%s]", metric)
	// Process the daemon metric & abort if an error was detected
	requiredLines, err := daemon.processDaemonMetric(metric)
	if err != nil {
		logger.Error(err.Error())
		return false
	}

	// Check if there is anything listening
	return daemon.isListening(requiredLines)
}

// parseDaemonJSON : parse a given json metric line into the expected schema
func (daemon *Listening) parseDaemonJSON(line string) (err error) {
	if len(line) <= 2 { //TODO: does it make sense to check regex, i.e. '{   }'
		msg := "Empty metrics line detected. Aborting execution..."
		logger.Error(msg)
		return fmt.Errorf(msg)
	}

	// Account for parsing errors
	defer func() {
		if r := recover(); r != nil {
			if re, ok := r.(error); re != nil && ok {
				err = re
			} else {
				err = fmt.Errorf("%v", r)
			}
			logger.Error("Error when decoding metric [%s]. Error [%s]. Failing metric...", line, err.Error())
		}
	}()

	// Attempt to parse the JSON text
	x := new(daemonJSONContainer)
	if err = json.NewDecoder(strings.NewReader(line)).Decode(x); err != nil {
		return err
	}

	// Detect duplicated keys
	//benchmarker.TimeItV(time.Nanosecond, validateUniqueKeys, line)
	validateUniqueKeys(line)

	// Container variable
	transformationContainer := new([]interface{})

	// Parse :: Port
	pipelineTransform(&x.PortRaw, &transformationContainer)
	for _, p := range *transformationContainer {
		if s, isString := p.(string); isString {
			r, err := strconv.Atoi(s)
			if err != nil {
				return err
			}
			// Validate if the found port is acceptable, if not, panics - used to jump stack frame
			validatePortRange(r)
			daemon.ports = append(daemon.ports, port(r))
		} else if i, isFloat := p.(float64); isFloat {
			// Validate if the found port is acceptable, if not, panics - used to jump stack frame
			validatePortRange(port(i))
			daemon.ports = append(daemon.ports, port(i))
		} else {
			return fmt.Errorf("the `port` value [%v] is not supported", p)
		}
	}

	// Parse :: Protocol
	pipelineTransform(&x.Protocol, &transformationContainer)
	for _, p := range *transformationContainer {
		s, isString := p.(string)
		if isString {
			validateProtocol(protocol(s))
			daemon.protocols = append(daemon.protocols, protocol(s))
		} else {
			return fmt.Errorf("the `protocol` value [%v] is not supported", p)
		}
	}

	// Parse :: IP version
	pipelineTransform(&x.IPVersion, &transformationContainer)
	for _, p := range *transformationContainer {
		s, isString := p.(string)
		if isString {
			if s == "ipv4" || s == "4" {
				s = ""
			} else if s == "ipv6" || s == "6" {
				s = "6"
			} else {
				return fmt.Errorf("the `ip` value [%s] is not supported", s)
			}
			daemon.ipVersions = append(daemon.ipVersions, ipVersion(s))
		} else {
			return fmt.Errorf("the `ip` value [%v] is not supported", p)
		}
	}

	// Parse :: Host
	pipelineTransform(&x.Host, &transformationContainer)
	for _, p := range *transformationContainer {
		s, isString := p.(string)
		if isString {
			daemon.hosts = append(daemon.hosts, host(s))
		} else {
			return fmt.Errorf("the `host` value [%v] is not supported", p)
		}
	}
	return err
}

// validatePortRange : validates that the given port is within the accepted range
func validatePortRange(port port) {
	if (port < 1) || (port > 65535) {
		panic(fmt.Sprintf("The specified port [%d] is out of range [1-65535]", port))
	}
}

// validateProtocol : validates that the given protocol is an accepted value
func validateProtocol(p protocol) {
	if p != "tcp" && p != "udp" {
		panic(fmt.Sprintf("The specified protocol [%s] is not supported", p))
	}
}

// pipelineTransform : helper function that populates a container based on a given interface.
// If the interface is of slice interface type, then the container takes its value.
// Else if the interface is of non slice interface type, then the container is created with the
//  interface value as its first element. This helps to abstract the handling of different
//  schemas.
func pipelineTransform(arg *interface{}, container **[]interface{}) {
	switch value := (*arg).(type) {
	case []interface{}:
		*container = &value
	case interface{}:
		*container = &[]interface{}{value}
	default:
		**container = nil
	}
}

// validateUniqueKeys : checks if more than one entry for the same key was detected. If so, present a warning message
// 	to the user
func validateUniqueKeys(line interface{}) {
	foundKeys := regexp.MustCompile(`("\w+" *:)`).FindAllString(line.(string), -1)
	foundKeyMap := make(map[string]bool, len(foundKeys))
	for _, key := range foundKeys {
		key := strings.Replace(key, " ", "", -1)
		keyA := []rune(key)
		if _, exists := foundKeyMap[key]; exists {
			logger.Warn("The key [%s] was found multiple times. Note that only the last declared key-value pair"+
				"will be used.", string(keyA[1:len(keyA)-2]))
		} else {
			foundKeyMap[key] = true
		}
	}
}

// processDaemonMetric : this function is responsible for the extraction of the JSON metric string from the
// 	configuration line. Attempt to parse the extracted JSON or use the hardcoded metric line if provided
//	(backwards-compatibility) process. Fetch the required amount of lines that need to be seen in the output of netstat.
// [@TODO: find new name(?)]
func (daemon *Listening) processDaemonMetric(metric string) (requiredLines int, err error) {
	// If no values were given to the Listening struct (required due to the backwards compatibility requirements.
	// 	For more information, see: http://configdocs.web.cern.ch/configdocs/dnslb/lbclientcodes.html
	if len(daemon.Metric) != 0 {
		metric = daemon.Metric
	}

	// The parsing is now always done
	metric = regexp.MustCompile("{(.*?)}").FindString(metric)
	// Parse json
	err = daemon.parseDaemonJSON(metric)
	if err != nil {
		return -1, err
	}

	// Fetch the required metric's amount
	requiredLines = daemon.getRequiredLines()

	// Apply default values
	err = daemon.applyDefaultValues()
	if err != nil {
		return -1, err
	}

	// The port metric argument is mandatory
	if len(daemon.ports) == 0 {
		return -1, fmt.Errorf("A port needs to be specified in a daemon check in the format `{port : <val>}`." +
			" Aborting check...")
	} else if len(daemon.protocols) == 0 {
		return -1, fmt.Errorf(`Failed to parse the given [protocol] entry. Only the following values are supported` +
			` ["tcp", "udp"]`)
	} else if len(daemon.ipVersions) == 0 {
		return -1, fmt.Errorf(`Failed to parse the given [ip] version entry. Only the following values are supported` +
			`["ipv4", "ipv6", "4", "6"]`)
	}

	logger.Trace("Finished processing metric file [%v]", daemon)
	return
}

// fetchAllLocalInterfaces : Fetch all local interfaces IPs and add them to the default array of hosts to check
func (daemon *Listening) fetchAllLocalInterfaces() error {
	// Log
	logger.Trace("Fetching all IPs from the local interfaces")

	// Retrieve all interfaces on the machine by default
	output, err := runner.RunCommand(`ifconfig`, true, true)
	if err != nil {
		logger.Error("Failed to fetch the interfaces from this machine. Error [%s]", err.Error())
		return err
	}
	outputIPs := regexp.MustCompile(`inet[0-9]?[ ][\w.:]*`).FindAllString(output, -1)
	logger.Trace("Found local addresses [%v]", outputIPs)
	for _, ip := range outputIPs {
		daemon.hosts = append(daemon.hosts, host(regexp.MustCompile("inet([6])?[ ]").Split(ip, -1)[1]))
	}

	// Add the any interface static IP pattern
	daemon.hosts = append(daemon.hosts, host("0.0.0.0"))
	daemon.hosts = append(daemon.hosts, host("::"))

	return nil
}

// isListening : checks if a daemon is listening on the given protocol(s) in the selected IP level and port
func (daemon *Listening) isListening(requiredLines int) bool {
	// Run the cli
	//TODO : pass netstat through awk,sort, uniq....
	// DO WE STILL NEED THE applyDefaults??
	res, err := runner.RunDirectCommand(daemonCheckCLI, true, true)
	if err != nil {
		logger.Error("An error was detected when attempting to run the daemon check cli. Error [%s]", err.Error())
		return false
	}

	if len(res) >= requiredLines {
		logger.Trace("Found daemon listening, expected [%d] and got [%d]", requiredLines, len(res))
		return true
	}
	logger.Trace("Unable to find daemon listening, expected [%d] lines but only got [%d], with entry [%v]",
		requiredLines, len(res), *daemon)
	return false
}

// applyDefaultValues : function responsible for setting the default values of the Hosts, IPVersions & Protocols.
// 	Will return an error if an issue was detected when attempting to retrieve Hosts
func (daemon *Listening) applyDefaultValues() error {
	// If no protocols were given
	if len(daemon.protocols) == 0 {
		daemon.protocols = defaultProtocols
	}

	// If no ip versions were given
	if len(daemon.ipVersions) == 0 {
		daemon.ipVersions = defaultIPVersions
	}

	// If no hosts were given
	if len(daemon.hosts) == 0 {
		// Fetch all listening interfaces
		err := daemon.fetchAllLocalInterfaces()
		if err != nil {
			return err
		}
	}

	return nil
}

// getRequiredLines :
func (daemon *Listening) getRequiredLines() (requiredLines int) {
	requiredLines = 1
	countIfNotZero(len(daemon.ports), &requiredLines)
	countIfNotZero(len(daemon.ipVersions), &requiredLines)
	countIfNotZero(len(daemon.hosts), &requiredLines)
	countIfNotZero(len(daemon.protocols), &requiredLines)
	return
}

// countIfNotZero : helper function for the [getRequiredLines] function
func countIfNotZero(value int, counter *int) {
	if value > 0 {
		*counter *= value
	}
}
