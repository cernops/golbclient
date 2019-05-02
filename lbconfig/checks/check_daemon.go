package checks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/utils/filehandler"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/utils/network"
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
	// Internal
	requiresTcp, requiresUdp, requiresIPv4, requiresIPv6 bool
	// Metric syntax
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
	err := daemon.processMetricLine(metric)
	if err != nil {
		logger.Error(err.Error())
		return -1, err
	}

	// Check if there is anything listening
	return daemon.isListening()
}

// parseMetricLineJSON : parse a given json Metric line into the expected schema
func (daemon *DaemonListening) parseMetricLineJSON(line string) (err error) {
	if len(line) <= 2 {
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
			s = strings.TrimSpace(s)
			if p != "tcp" && p != "udp" {
				panic(fmt.Sprintf("The specified protocol [%s] is not supported", p))
			} else {
				if p == "tcp" {
					daemon.requiresTcp = true
				} else {
					daemon.requiresUdp = true
				}
			}
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
				daemon.requiresIPv4 = true
			} else if s == "ipv6" || s == "6" {
				s = "6"
				daemon.requiresIPv6 = true
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
func (daemon *DaemonListening) processMetricLine(metric string) error {
	// If no values were given to the Listening struct (required due to the backwards compatibility requirements.
	// 	For more information, see: http://configdocs.web.cern.ch/configdocs/dnslb/lbclientcodes.html
	if len(daemon.Metric) != 0 {
		metric = daemon.Metric
	}

	// The parsing is now always done
	metric = regexp.MustCompile("{(.*?)}").FindString(metric)
	// Parse json
	err := daemon.parseMetricLineJSON(metric)
	if err != nil {
		return err
	}

	// Apply default values
	err = daemon.applyDefaultValues()
	if err != nil {
		return err
	}

	// The port Metric argument is mandatory
	if len(daemon.Ports) == 0 {
		return fmt.Errorf("a port needs to be specified in a daemon check in the format `{port : <val>}`")
	}

	logger.Trace("Finished processing Metric file [%#v]", daemon)
	return nil
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
func (daemon *DaemonListening) isListening() (int, error) {
	portsFormat 		:= daemon.getPortsRegexFormat()
	hostsFormat, err 	:= daemon.getHostRegexFormat()
	if err != nil {
		return -1, err
	}

	// Get all the Ports combination
	regex 		:= regexp.MustCompile(
		fmt.Sprintf(`[0-9]+: (%s):(%s)`, hostsFormat, portsFormat))
	logger.Trace("Looking with regex [%s] for open ports...", regex.String())


	var foundLines int

	// TCP & IPv4
	err = sumAndMatchIfRequired(daemon.requiresIPv4 && daemon.requiresTcp, tcp4ConfSock, regex, &foundLines)
	if err != nil {
		return -1, err
	}

	// TCP & IPv6
	err = sumAndMatchIfRequired(daemon.requiresIPv6 && daemon.requiresTcp, tcp6ConfSock, regex, &foundLines)
	if err != nil {
		return -1, err
	}

	// UDP & IPv4
	err = sumAndMatchIfRequired(daemon.requiresIPv4 && daemon.requiresUdp, udp4ConfSock, regex, &foundLines)
	if err != nil {
		return -1, err
	}

	// UDP & IPv6
	err = sumAndMatchIfRequired(daemon.requiresIPv6 && daemon.requiresUdp, udp6ConfSock, regex, &foundLines)
	if err != nil {
		return -1, err
	}

	if foundLines < 1 {
		return -1, fmt.Errorf("expected to find at least 1 matching line for the daemon check [%#v]", daemon)
	}

	return 1, nil
}

// applyDefaultValues : function responsible for setting the default values of the Hosts, IPVersions & Protocols.
// 	Will return an error if an issue was detected when attempting to retrieve Hosts
func (daemon *DaemonListening) applyDefaultValues() error {
	// If no Protocols were given
	if len(daemon.Protocols) == 0 {
		daemon.Protocols = defaultProtocols
		daemon.requiresTcp = true
		daemon.requiresUdp = true
	}

	// If no ip versions were given
	if len(daemon.IpVersions) == 0 {
		daemon.IpVersions = defaultIPVersions
		daemon.requiresIPv4 = true
		daemon.requiresIPv6 = true
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

// getHostRegexFormat : Helper function that creates a regex-ready string from all the found [daemon.Hosts] entries
func (daemon *DaemonListening) getHostRegexFormat() (string, error) {
	// Get all hosts in a regex-ready format
	var hostsFormat bytes.Buffer
	for i, h := range daemon.Hosts {
		if i != 0 {
			hostsFormat.WriteString("|")
		}
		hostHex, err := network.GetPackedReprFromIP(h)
		if err != nil {
			return "", err
		}

		logger.Trace("Scanning host [%s] with HEX [%s]", h, hostHex)
		hostsFormat.WriteString(fmt.Sprintf("(%s)", hostHex))
	}

	return hostsFormat.String(), nil
}

// getPortsRegexFormat : Helper function that creates a regex-ready string from all the found [daemon.Ports] entries
func (daemon *DaemonListening) getPortsRegexFormat() string {
	// Get all ports in a regex-ready format
	var portsFormat bytes.Buffer
	for i, p := range daemon.Ports {
		if i != 0 {
			portsFormat.WriteString("|")
		}

		portHex := strings.ToUpper(fmt.Sprintf("%04x", p))
		logger.Trace("Scanning port [%d] with HEX [%s]", p, portHex)
		portsFormat.WriteString(fmt.Sprintf("(%s)", portHex))
	}
	return portsFormat.String()
}

// sumAndMatchIfRequired : Helper function that only executed if the given condition evaluates to [true]. If so,
// the file contents of @arg sockPath will be read and matched against the given @arg regex. The amount of matches
// will then be added to the given counter
func sumAndMatchIfRequired(cond bool, sockPath string, regex *regexp.Regexp, counter *int) error {
	if cond {
		logger.Trace("Looking for regex in sock file [%s]...", sockPath)
		
		fileContent, err := filehandler.ReadAllLinesFromFileAsString(sockPath, " ")
		if err != nil {
			return fmt.Errorf("unable to open the file [%s]. Error [%s]", sockPath, err)
		}
		foundLines := len(regex.FindStringSubmatch(fileContent))
		*counter += foundLines
		logger.Debug("Found [%d] matching lines in sock file...", foundLines, sockPath)
	}
	return nil
}