package checks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/utils/network"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
)

// DaemonListening : struct responsible for all the daemon check slices
type DaemonListening struct {
	// Internal
	requiresTCP, requiresUDP, requiresIPV4, requiresIPV6 bool
	// Metric syntax
	Ports []int
	Hosts []string
	// Backwards compatibility
	Metric string
}

// Configuration file socks
const (
	tcp4ConfSock string = `/proc/net/tcp`
	tcp6ConfSock        = `/proc/net/tcp6`
	udp4ConfSock        = `/proc/net/udp`
	udp6ConfSock        = `/proc/net/udp6`
)

// daemonJsonContainer : Helper struct
type daemonJSONContainer struct {
	PortRaw   interface{} `json:"port"`
	Protocol  interface{} `json:"protocol"`
	IPVersion interface{} `json:"ip"`
	Host      interface{} `json:"host"`
}

// condFilePair : Helper struct to achieve code-reuse
type condFilePair struct {
	cond bool
	filepath string
}

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
			err = validatePortRange(r)
			if err != nil {
				return err
			}

			daemon.Ports = append(daemon.Ports, r)
		} else if i, isFloat := p.(float64); isFloat {
			err := validatePortRange(int(i))
			if err != nil {
				return err
			}
			daemon.Ports = append(daemon.Ports, int(i))
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
				return fmt.Errorf("the specified protocol [%s] is not supported", p)
			} else {
				if p == "tcp" {
					daemon.requiresTCP = true
				} else {
					daemon.requiresUDP = true
				}
			}
		} else {
			return fmt.Errorf("the `protocol` value [%v] is not supported", p)
		}
	}

	// Protocol :: If no Protocols were given
	if !daemon.requiresTCP && !daemon.requiresUDP {
		daemon.requiresTCP = true
		daemon.requiresUDP = true
	}

	// Parse :: IP version
	pipelineTransform(&x.IPVersion, &transformationContainer)
	for _, p := range *transformationContainer {
		s, isString := p.(string)
		if isString {
			if s == "ipv4" || s == "4" {
				s = ""
				daemon.requiresIPV4 = true
			} else if s == "ipv6" || s == "6" {
				s = "6"
				daemon.requiresIPV6 = true
			} else {
				return fmt.Errorf("the `ip` value [%s] is not supported", s)
			}
		} else {
			return fmt.Errorf("the `ip` value [%v] is not supported", p)
		}
	}

	// IP version :: If no ip versions were given
	if !daemon.requiresIPV4 && !daemon.requiresIPV6 {
		daemon.requiresIPV4 = true
		daemon.requiresIPV6 = true
	}

	// Parse :: Host
	pipelineTransform(&x.Host, &transformationContainer)
	for _, p := range *transformationContainer {
		s, isString := p.(string)
		if isString {
			daemon.Hosts = append(daemon.Hosts, s)
		} else {
			return fmt.Errorf("the `host` value [%v] is not supported", p)
		}
	}
	return err
}

// validatePortRange : validates that the given port is within the accepted range
func validatePortRange(port int) error {
	if (port < 1) || (port > 65535) {
		return fmt.Errorf("the specified port [%d] is out of range [1-65535]", port)
	}
	return nil
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

	// The port Metric argument is mandatory
	if len(daemon.Ports) == 0 {
		return fmt.Errorf("a port needs to be specified in a daemon check in the format `{port : <val>}`")
	}

	logger.Trace("Finished processing Metric file [%#v]", daemon)
	return nil
}

// isListening : checks if a daemon is listening on the given protocol(s) in the selected IP level and port
func (daemon *DaemonListening) isListening() (int, error) {
	portsFormat := daemon.getPortsRegexFormat()
	hostsFormat, err := daemon.getHostRegexFormat()
	if err != nil {
		return -1, err
	}

	// Get all the Ports combination
	regex := regexp.MustCompile(
		fmt.Sprintf(`[0-9]+: (%s):(%s)`, hostsFormat, portsFormat))
	logger.Trace("Looking with regex [%s] for open ports...", regex.String())

	// Conditions & file-lookup map
	condFilePairList := []condFilePair{
		{daemon.requiresIPV4 && daemon.requiresTCP, tcp4ConfSock}, // TCP & IPv4
		{daemon.requiresIPV6 && daemon.requiresTCP, tcp6ConfSock}, // TCP & IPv6
		{daemon.requiresIPV4 && daemon.requiresUDP, udp4ConfSock}, // UDP & IPv4
		{daemon.requiresIPV6 && daemon.requiresUDP, udp6ConfSock}, // UDP & IPv6
	}

	for _, cfp := range condFilePairList {
		foundLines, err := matchIfRequired(cfp.cond, cfp.filepath, regex)
		if err != nil {
			return -1, err
		}

		if foundLines >= 1 {
			logger.Trace("Found the required ports [%s] listening on [%s]", daemon.Ports, cfp.filepath)
			return 1, nil
		}
	}

	return -1, fmt.Errorf("failed to find the required open ports [%v]", daemon.Ports)
}

// getHostRegexFormat : Helper function that creates a regex-ready string from all the found [daemon.Hosts] entries
func (daemon *DaemonListening) getHostRegexFormat() (string, error) {
	// Return wildcard if no hosts were specified by the user
	if len(daemon.Hosts) == 0 {
		return ".*", nil
	}
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

// matchIfRequired : Helper function that only executed if the given condition evaluates to [true]. If so,
// the file contents of @arg sockPath will be read and matched against the given @arg regex. The amount of matches
// will then be added to the given counter
func matchIfRequired(cond bool, sockPath string, regex *regexp.Regexp) (foundLines int, err error) {
	if cond {
		logger.Trace("Looking for regex in sock file [%s]...", sockPath)

		fileContent, err := ioutil.ReadFile(sockPath)
		if err != nil {
			return foundLines, fmt.Errorf("unable to open the file [%s]. Error [%s]", sockPath, err)
		}
		foundLines = len(regex.FindStringSubmatch(string(fileContent)))
		logger.Debug("Found [%d] matching lines in sock file [%s]...", foundLines, sockPath)
	}
	return
}
