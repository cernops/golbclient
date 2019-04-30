package checks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/utils/filehandler"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/utils/runner"
	"regexp"
	"strconv"
	"strings"

	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
)

// port : int alias-type to represent a port number
type port = int

// protocol : string alias-type used to distinguish between different transport protocol types
type protocol = string

// ipVersion : int alias-type used to distinguish between different IP levels
type ipVersion = string

// host : string alias-type used to identify in which host the daemon should be listening on
type host = string

// DaemonListening : struct responsible for all the daemon check slices
type DaemonListening struct {
	Ports      []port
	Protocols  []protocol
	IpVersions []ipVersion
	Hosts      []host
	// Backwards compatibility
	Metric string
}

// Configuration file socks
const (
	tcp4ConfSock string = `/proc/net/tcp`
	tcp6ConfSock        = `/proc/net/tcp6`
	udp4ConfSock 		= `/proc/net/udp`
	udp6ConfSock        = `/proc/net/udp6`
)

// daemonJsonContainer : Helper struct
type daemonJSONContainer struct {
	PortRaw   interface{} `json:"port"`
	Protocol  interface{} `json:"protocol"`
	IPVersion interface{} `json:"ip"`
	Host      interface{} `json:"host"`
}


// listening_default_values_map : map that is used to set the default values for
// the ones not found in the Metric line
var defaultProtocols = []protocol{"tcp", "udp"}
var defaultIPVersions = []ipVersion{"ipv4", "ipv6"}


func (daemon DaemonListening) Run(args ...interface{}) (int, error) {
	metric := args[0].(string)

	// Log
	logger.Trace("Processing daemon check on Metric line [%s]", metric)
	// Process the daemon Metric & abort if an error was detected
	requiredLines, err := daemon.processMetricLine(metric)
	if err != nil {
		logger.Error(err.Error())
		return -1, err
	}

	// Check if there is anything listening
	return daemon.isListening(requiredLines)
}

// parseMetricLineJSON : parse a given json Metric line into the expected schema
func (daemon *DaemonListening) parseMetricLineJSON(line string) (err error) {
	if len(line) <= 2 { //TODO: does it make sense to check regex, i.e. '{   }'
		return fmt.Errorf("empty metrics line detected. Failing the check")
	}

	// Account for parsing errors
	defer func() {
		if r := recover(); r != nil {
			if re, ok := r.(error); re != nil && ok {
				err = re
			} else {
				err = fmt.Errorf("%v", r)
			}
			logger.Error("Error when decoding Metric [%s]. Error [%s]. Failing Metric...", line, err.Error())
		}
	}()

	// Attempt to parse the JSON text
	x := new(daemonJSONContainer)
	if err = json.NewDecoder(strings.NewReader(line)).Decode(x); err != nil {
		return err
	}

	// Detect duplicated keys
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
			daemon.Ports = append(daemon.Ports, port(r))
		} else if i, isFloat := p.(float64); isFloat {
			// Validate if the found port is acceptable, if not, panics - used to jump stack frame
			validatePortRange(port(i))
			daemon.Ports = append(daemon.Ports, port(i))
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
			daemon.Protocols = append(daemon.Protocols, protocol(s))
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
			daemon.IpVersions = append(daemon.IpVersions, ipVersion(s))
		} else {
			return fmt.Errorf("the `ip` value [%v] is not supported", p)
		}
	}

	// Parse :: Host
	pipelineTransform(&x.Host, &transformationContainer)
	for _, p := range *transformationContainer {
		s, isString := p.(string)
		if isString {
			daemon.Hosts = append(daemon.Hosts, host(s))
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

// processMetricLine : this function is responsible for the extraction of the JSON Metric string from the
// 	configuration line. Attempt to parse the extracted JSON or use the hardcoded Metric line if provided
//	(backwards-compatibility) process. Fetch the required amount of lines that need to be seen in the output of netstat.
func (daemon *DaemonListening) processMetricLine(metric string) (requiredLines int, err error) {
	// If no values were given to the Listening struct (required due to the backwards compatibility requirements.
	// 	For more information, see: http://configdocs.web.cern.ch/configdocs/dnslb/lbclientcodes.html
	if len(daemon.Metric) != 0 {
		metric = daemon.Metric
	}

	// The parsing is now always done
	metric = regexp.MustCompile("{(.*?)}").FindString(metric)
	// Parse json
	err = daemon.parseMetricLineJSON(metric)
	if err != nil {
		return -1, err
	}

	// Fetch the required Metric's amount
	requiredLines = daemon.getRequiredLines()

	// Apply default values
	err = daemon.applyDefaultValues()
	if err != nil {
		return -1, err
	}

	// The port Metric argument is mandatory
	if len(daemon.Ports) == 0 {
		return -1, fmt.Errorf("a port needs to be specified in a daemon check in the format `{port : <val>}`." +
			" Failing check")
	} else if len(daemon.Protocols) == 0 {
		return -1, fmt.Errorf(`failed to parse the given [protocol] entry. Only the following values are supported` +
			` ["tcp", "udp"]`)
	} else if len(daemon.IpVersions) == 0 {
		return -1, fmt.Errorf(`failed to parse the given [ip] version entry. Only the following values are supported` +
			`["ipv4", "ipv6", "4", "6"]`)
	}

	logger.Trace("Finished processing Metric file [%#v]", daemon)
	return
}

// fetchAllLocalInterfaces : Fetch all local interfaces IPs and add them to the default array of Hosts to check
func (daemon *DaemonListening) fetchAllLocalInterfaces() error {
	// Log
	logger.Trace("Fetching all IPs from the local interfaces")

	// Retrieve all interfaces on the machine by default
	output, err := runner.RunCommand(`ip addr show`, true, 0)
	if err != nil {
		logger.Error("Failed to fetch the interfaces from this machine. Error [%s]", err.Error())
		return err
	}
	outputIPs := regexp.MustCompile(`inet[0-9]?[ ][\w.:]*`).FindAllString(output, -1)
	logger.Trace("Found local addresses [%v]", outputIPs)
	for _, ip := range outputIPs {
		daemon.Hosts = append(daemon.Hosts, host(regexp.MustCompile(`inet([6])?[ ]`).Split(ip, -1)[1]))
	}

	// Add the any interface static IP pattern
	daemon.Hosts = append(daemon.Hosts, host("0.0.0.0"))
	daemon.Hosts = append(daemon.Hosts, host("::"))

	return nil
}

// isListening : checks if a daemon is listening on the given protocol(s) in the selected IP level and port
func (daemon *DaemonListening) isListening(requiredLines int) (int, error) {
	var foundLines int

	var portsFormat bytes.Buffer
	for i, p := range daemon.Ports {
		if i != 0 {
			portsFormat.WriteString("|")
		}

		portHex := strings.ToUpper(fmt.Sprintf("%04x", p))
		logger.Trace("Scanning port [%d] with HEX [%s]", p, portHex)
		portsFormat.WriteString(fmt.Sprintf("(%s)", portHex))
	}

	// Get all the Ports combination
	regex 		:= regexp.MustCompile(
		fmt.Sprintf(` *[0-9]+: [0-9A-F]+:(%s) [0-9A-F]+:[0-9A-F]+ 0A`, portsFormat.String()))

	logger.Trace("Looking with regex [%s] for open ports...", regex.String())

	// @TODO only read the required files
	tcp4Content, err := filehandler.ReadAllLinesFromFileAsString(tcp4ConfSock, " ")
	if err != nil {
		return -1, fmt.Errorf("unable to open the file [%s]. Error [%s]", tcp4ConfSock, err)
	}
	tcp6Content, err := filehandler.ReadAllLinesFromFileAsString(tcp6ConfSock, " ")
	if err != nil {
		return -1, fmt.Errorf("unable to open the file [%s]. Error [%s]", tcp6ConfSock, err)
	}
	udp4Content, err := filehandler.ReadAllLinesFromFileAsString(udp4ConfSock, " ")
	if err != nil {
		return -1, fmt.Errorf("unable to open the file [%s]. Error [%s]", udp4ConfSock, err)
	}
	udp6Content, err := filehandler.ReadAllLinesFromFileAsString(udp6ConfSock, " ")
	if err != nil {
		return -1, fmt.Errorf("unable to open the file [%s]. Error [%s]", udp6ConfSock, err)
	}

	foundLines += len(regex.FindStringSubmatch(tcp4Content))
	foundLines += len(regex.FindStringSubmatch(tcp6Content))
	foundLines += len(regex.FindStringSubmatch(udp4Content))
	foundLines += len(regex.FindStringSubmatch(udp6Content))

	if foundLines < requiredLines {
		return -1, fmt.Errorf("expected to find [%d] lines but got [%d] instead", requiredLines, foundLines)
	}

	return 1, nil
}

// applyDefaultValues : function responsible for setting the default values of the Hosts, IPVersions & Protocols.
// 	Will return an error if an issue was detected when attempting to retrieve Hosts
func (daemon *DaemonListening) applyDefaultValues() error {
	// If no Protocols were given
	if len(daemon.Protocols) == 0 {
		daemon.Protocols = defaultProtocols
	}

	// If no ip versions were given
	if len(daemon.IpVersions) == 0 {
		daemon.IpVersions = defaultIPVersions
	}

	// If no Hosts were given
	if len(daemon.Hosts) == 0 {
		// Fetch all listening interfaces
		err := daemon.fetchAllLocalInterfaces()
		if err != nil {
			return err
		}
	}

	return nil
}

// getRequiredLines :
func (daemon *DaemonListening) getRequiredLines() (requiredLines int) {
	requiredLines = 1
	countIfNotZero(len(daemon.Ports), &requiredLines)
	countIfNotZero(len(daemon.IpVersions), &requiredLines)
	countIfNotZero(len(daemon.Hosts), &requiredLines)
	countIfNotZero(len(daemon.Protocols), &requiredLines)
	return
}

// countIfNotZero : helper function for the [getRequiredLines] function
func countIfNotZero(value int, counter *int) {
	if value > 0 {
		*counter *= value
	}
}
