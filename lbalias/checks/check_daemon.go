package checks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gitlab.cern.ch/lb-experts/golbclient/lbalias/utils/benchmarker"
	"gitlab.cern.ch/lb-experts/golbclient/lbalias/utils/runner"
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const daemonCheckCLI = "/bin/netstat -luntap"

type Listening struct {
	Ports      []Port
	Protocols  []Protocol
	IPVersions []IPVersion
	// The host array is fetched at runtime
	Hosts      []Host
}



// Helper struct
type daemonJsonContainer struct {
	PortRaw interface{} `json:"port"`
	Protocol  interface{} `json:"protocol"`
	IPVersion interface{} `json:"ip"`
	Host      interface{} `json:"host"`
}

// listening_default_values_map : map that is used to set the default values for
// the ones not found in the metric line
var defaultProtocols = []Protocol{"tcp", "udp"}
var defaultIPVersions = []string{"ipv4", "ipv6"}

// port : int alias-type to represent a port number
type Port = int

// protocol : string alias-type used to distinguish between different transport protocol types
type Protocol = string

// ipLevel : int alias-type used to distinguish between different IP levels
type IPVersion = string

// host : string alias-type used to identify in which host the daemon should be listening on
type Host = string

// interfaceJoin : joins all the given structs in a string separated with the chosen delimiter.
func interfaceJoin(iface interface{}, delim string) (_ string) {
	var res bytes.Buffer

		val := reflect.ValueOf(iface)
		if val.Kind() == reflect.Slice {
			for i := 0; i < val.Len(); i++ {
				if i == val.Len() - 1 {
					delim = ""
				}
				res.WriteString(fmt.Sprintf("%s%s", val.Index(i).Interface(), delim))
			}
		} else {
			logger.Error("Unable to join the given interface [%v]", iface)
			return
		}
	return res.String()
}

func (daemon Listening) Run(args ...interface{}) interface{} {
	metric := args[0].(string)

	// Log
	logger.Trace("Processing daemon check on metric line [%s]", metric)
	// Process the daemon metric & abort if an error was detected
	requiredLines, err := daemon.processDaemonMetric(metric)
	if err != nil {
		return false
	}

	// Check if there is anything listening
	return daemon.isListening(requiredLines)
}

// parseDaemonJSON : parse a given json metric line into the expected schema
func (daemon *Listening) parseDaemonJSON(line string) (err error) {
	if len(line) == 0 {
		logger.Warn("Skipping empty metric line...")
		return err
	}

	// Account for parsing errors
	defer func() {
		if r := recover(); r != nil {
			if re, ok := r.(error); re != nil && ok {
				err = re
			} else {
				err = fmt.Errorf("unexpected error when decoding metric [%s]", line)
			}
			logger.Error("Error when decoding metric [%s]. Error [%s]. Failing metric...", line, err.Error())
		}
	}()

	// Attempt to parse the JSON text
	x := new(daemonJsonContainer)
	if err := json.NewDecoder(strings.NewReader(line)).Decode(x); err != nil {
		return err
	}

	// Detect duplicated keys
	benchmarker.TimeItV(time.Nanosecond, validateUniqueKeys, line)
	//validateUniqueKeys(line)

	// Container variable
	var transformationContainer []interface{}

	// Parse :: Port
	pipelineTransform(&x.PortRaw, &transformationContainer)
	for _, p := range transformationContainer {
		if s, isString := p.(string); isString {
			r, err := strconv.Atoi(s)
			if err != nil {
				return err
			}
			// Validate if the found port is acceptable, if not, panics - used to jump stack frame
			validatePortRange(r)
			daemon.Ports = append(daemon.Ports , Port(r))
		}
		
		if i, isFloat := p.(float64); isFloat {
			// Validate if the found port is acceptable, if not, panics - used to jump stack frame
			validatePortRange(Port(i))
			daemon.Ports = append(daemon.Ports , Port(i))
		}
	}

	// Parse :: Protocol
	pipelineTransform(&x.Protocol, &transformationContainer)
	for _, p := range transformationContainer {
		s, isString := p.(string)
		if isString {
			daemon.Protocols = append(daemon.Protocols , Protocol(s))
		}
	}

	// Parse :: IP version
	pipelineTransform(&x.IPVersion, &transformationContainer)
	for _, p := range transformationContainer {
		s, isString := p.(string)
		if isString {
			if s == "ipv4" || s == "4" {
				s = ""
			} else if s == "ipv6" || s == "6" {
				s = "6"
			} else {
				continue
			}

			daemon.IPVersions = append(daemon.IPVersions, IPVersion(s))
		}
	}

	// Parse :: Host
	pipelineTransform(&x.Host, &transformationContainer)
	for _, p := range transformationContainer {
		s, isString := p.(string)
		if isString {
			daemon.Hosts = append(daemon.Hosts, Host(s))
		}
	}

	return err
}

// validatePortRange : validates that the given port is within the accepted range
func validatePortRange(port int) {
	if (port < 1) || (port > 65535) {
		panic(fmt.Sprintf("The specified port [%d] is out of range [1-65535]", port))
	}
}

// pipelineTransform : helper function that populates a container based on a given interface.
// If the interface is of slice interface type, then the container takes its value.
// Else if the interface is of non slice interface type, then the container is created with the
//  interface value as its first element. This helps to abstract the handling of different
//  schemas.
func pipelineTransform(arg *interface{}, container *[]interface{}){
	if valueArray, ok := (*arg).([]interface{}); ok {
		*container = valueArray
	} else if valueEntry, ok := (*arg).(interface{}); ok {
		container = &[]interface{}{valueEntry}
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

// isListening : checks if a daemon is listening on the given protocol(s) in the selected IP level and port
func (daemon *Listening) processDaemonMetric(metric string) (int, error) {
	metric = regexp.MustCompile("{(.*?)}").FindString(metric)

	var err error
	// If no values were given to the Listening struct (required due to the backwards compatibility requirements.
	// 	For more information, see: http://configdocs.web.cern.ch/configdocs/dnslb/lbclientcodes.html
	if len(daemon.IPVersions) == 0 && len(daemon.Hosts) == 0 &&	len(daemon.Ports) == 0 && len(daemon.Protocols) == 0 {
		// Parse json
		err := daemon.parseDaemonJSON(metric)
		if err != nil {
			return -1, err
		}
	}

	// Fetch the required metric's amount
	requiredLines := daemon.getRequiredLines()

	// Apply default values
	err = daemon.applyDefaultValues()
	if err != nil {
		return -1, err
	}

	logger.Trace("Finished processing metric file [%v]", daemon)
	return requiredLines, nil
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
		daemon.Hosts = append(daemon.Hosts, Host(regexp.MustCompile("inet([6])?[ ]").Split(ip, -1)[1]))
	}

	// Add the any interface static IP pattern
	daemon.Hosts = append(daemon.Hosts, Host("0.0.0.0"))

	return nil
}

// isListening : checks if a daemon is listening on the given protocol(s) in the selected IP level and port
func (daemon *Listening) isListening(requiredLines int) bool {
	// The port metric argument is mandatory
	if len(daemon.Ports) == 0 {
		logger.Error("A port needs to be specified in a daemon check in the format `{port : <val>}`."+
			"Aborting check...")
		return false
	} else if len(daemon.Protocols) == 0 {
		logger.Error(`Failed to parse the given [protocol] entry. Only the following values are supported`+
			`["tcp", "udp"]`)
		return false
	} else if len(daemon.IPVersions) == 0 {
		logger.Error(`Failed to parse the given [ip] version entry. Only the following values are supported`+
			`["ipv4", "ipv6", "4", "6"]`)
		return false
	}

	// Run the cli
	res, err := runner.RunDirectCommand(daemonCheckCLI, true, true)
	if err != nil {
		logger.Error("An error was detected when attempting to run the daemon check cli. Error [%s]", err.Error())
		return false
	}

	// Prepare the regex expression
	ports := strings.Trim(strings.Replace(fmt.Sprint(daemon.Ports), " ", "|", -1), "[]")
	protocols := interfaceJoin(daemon.Protocols, "|")
	hosts := interfaceJoin(daemon.Hosts, "|")
	expression := fmt.Sprintf(`(?i)(%s(6)?)([ ]+[0-9]+[ ]+[0-9]+[ ]+(%s))([:](%s))(.*)(LISTEN|[ ]+)`,
		protocols, hosts, ports)

	filteredRes := regexp.MustCompile(expression).FindAllString(res, -1)

	if len(filteredRes) >= requiredLines {
		logger.Trace("Found daemon listening, matching lines [%d]," +
			" expression [%s], with entry [%v]", len(filteredRes), expression, *daemon)
		return true
	} else {
		logger.Trace("Unable to find daemon listening, expected [%d]" +
			" lines but only got [%d], with entry [%v]", requiredLines, len(filteredRes), *daemon)
		return false
	}
}

// applyDefaultValues : function responsible for setting the default values of the Hosts, IPVersions & Protocols.
// 	Will return an error if an issue was detected when attempting to retrieve Hosts
func (daemon *Listening) applyDefaultValues() error {
	// If no protocols were given
	if len(daemon.Protocols) == 0 {
		daemon.Protocols = defaultProtocols
	}

	// If no ip versions were given
	if len(daemon.IPVersions) == 0 {
		daemon.IPVersions = defaultIPVersions
	}

	// If no hosts were given
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
func (daemon *Listening) getRequiredLines() (requiredLines int) {
	countIfNotZero(len(daemon.Ports), &requiredLines)
	countIfNotZero(len(daemon.IPVersions), &requiredLines)
	countIfNotZero(len(daemon.Hosts), &requiredLines)
	countIfNotZero(len(daemon.Protocols), &requiredLines)
	return
}

// countIfNotZero : helper function for the [getRequiredLines] function
func countIfNotZero(value int, counter *int){
	if value > 0 {
		if *counter == 0 {
			*counter += value
		} else {
			*counter *= value
		}
	}
}
