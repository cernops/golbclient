package checks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/creasty/defaults"
	"github.com/google/go-cmp/cmp"
	"gitlab.cern.ch/lb-experts/golbclient/lbalias/utils/benchmarker"
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
	"gitlab.cern.ch/lb-experts/golbclient/lbalias/utils/runner"
	"reflect"
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

// protocol : string type used to distinguish between different transport protocol types
type Protocol string

// ipLevel : int type used to distinguish between different IP levels
type IPVersion string

// host : string type used to identify in which host the daemon should be listening on
type Host string

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


	// Check if there is anything listening
	return daemon.isListening()
}

// parseDaemonJSON : parse a given json metric line into the expected schema
func (daemon *Listening) parseDaemonJSON(line string) (err error) {
	if len(line) == 0 {
		logger.Trace("Skipping empty metric line...")
		return err
	}

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
	benchmarker.TimeItV(time.Nanosecond, validateUniqueKeys, line)
	//validateUniqueKeys(line)

	// Container variable
	var transformationContainer []interface{}

	// Parse :: Port
	if portEntry, ok := x.PortRaw.([]interface{}); ok {
		transformationContainer = portEntry
	} else if portEntry, ok := x.PortRaw.(interface{}); ok {
		transformationContainer = []interface{}{portEntry}
	}

	for _, p := range transformationContainer {
		if s, isString := p.(string); isString {
			r, err := strconv.Atoi(s)
			if err != nil {
				return err
			}
			daemon.Port = append(daemon.Port , Port(r))
		}
		
		if i, isFloat := p.(float64); isFloat {
			daemon.Port = append(daemon.Port , Port(i))
		}
	}

	// Parse :: Protocol
	if x.Protocol != nil {
		daemon.Protocol = daemon.Protocol[:0]
	}
	if protocolEntry, ok := x.Protocol.([]interface{}); ok {
		transformationContainer = protocolEntry
	} else if protocolEntry, ok := x.Protocol.(interface{}); ok {
		transformationContainer = []interface{}{protocolEntry}
	}
	for _, p := range transformationContainer {
		s, isString := p.(string)
		if isString {
			daemon.Protocol = append(daemon.Protocol , Protocol(s))
		}
	}

	// Clear default entries
	if x.IPVersion != nil {
		daemon.IPVersion = daemon.IPVersion[:0]
	}
	// Parse :: IP version
	if ipVersionEntry, ok := x.IPVersion.([]interface{}); ok {
		transformationContainer = ipVersionEntry
	} else if ipVersionEntry, ok := x.IPVersion.(interface{}); ok {
		transformationContainer = []interface{}{ipVersionEntry}
	}
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

			daemon.IPVersion = append(daemon.IPVersion , IPVersion(s))
		}
	}


	if x.Host != nil {
		daemon.Host = daemon.Host[:0]
	}
	// Parse :: Host
	if hostEntry, ok := x.Host.([]interface{}); ok {
		transformationContainer = hostEntry
	} else if hostEntry, ok := x.Host.(interface{}); ok {
		transformationContainer = []interface{}{hostEntry}
	}
	for _, p := range transformationContainer {
		s, isString := p.(string)
		if isString {
			daemon.Host = append(daemon.Host , Host(s))
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

// isListening : checks if a daemon is listening on the given protocol(s) in the selected IP level and port
func (daemon *Listening) processDaemonMetric(metric string) error {
	metric = regexp.MustCompile("{(.*?)}").FindString(metric)

	// Parse json
	err := daemon.parseDaemonJSON(metric)
	if err != nil {
		logger.Error("Error when decoding metric [%s]. Error [%s]. Failing metric...", metric, err.Error())
		return err
	}

	for _, port := range daemon.Port {
		// Check if the given port is within bounds
		if (port < 1) || (port > 65535) {
			logger.Error("The specified port [%d] is out of range [1-65535]", port)
			return fmt.Errorf("ignored")
		}
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
	} else if len(daemon.Protocol) == 0 {
		logger.Error(`Failed to parse the given [protocol] entry. Only the following values are supported ["tcp", "udp"]`)
		return false
	} else if len(daemon.IPVersion) == 0 {
		logger.Error(`Failed to parse the given [ip] version entry. Only the following values are supported ["ipv4", "ipv6", "4", "6"]`)
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
	logger.Trace("Daemon check need any condition :: [%t]. Daemon entry [%v]", needAny, *daemon)

	/*
		V2

	 */

	ports := strings.Trim(strings.Replace(fmt.Sprint(daemon.Port), " ", "|", -1), "[]")
	protocols := interfaceJoin(daemon.Protocol, "|")
	hosts := interfaceJoin(daemon.Host, "|")
	expression := fmt.Sprintf(`(?i)(%s(6)?)([ ]+[0-9]+[ ]+[0-9]+[ ]+(%s))([:](%s))(.*)(LISTEN|[ ]+)`, protocols, hosts, ports)

	filteredRes := regexp.MustCompile(expression).FindAllString(res, -1)

	if (needAny && len(filteredRes) >= 1) || (!needAny && len(filteredRes) == len(daemon.Port) * len(daemon.Protocol) * len(daemon.Host) * len(daemon.IPVersion)){
		logger.Trace("Found daemon listening, matching lines [%d], [any] condition [%t] and expression [%s], with entry [%v]", len(filteredRes), needAny, expression, *daemon)
		return true
	} else {
		logger.Trace("Unable to find daemon listening, [any] condition [%t], with entry [%v]", *daemon)
		return false
	}

	/*
	   V2

	*/


	/*

	// If the any condition is required, then merge all expression into a single one
	if needAny {
		ports := strings.Trim(strings.Replace(fmt.Sprint(daemon.Port), " ", "|", -1), "[]")
		protocols := interfaceJoin(daemon.Protocol, "|")
		hosts := interfaceJoin(daemon.Host, "|")

		// IP version is not needed because the expression will find all (e.g. tcp/tcp6)
		expression := fmt.Sprintf(`(?i)(%s(6)?)([ ]+[0-9]+[ ]+[0-9]+[ ]+(%s))([:](%s))(.*)(LISTEN)`, protocols, hosts, ports)
		if regexp.MustCompile(expression).MatchString(res) {
			logger.Trace(`Found daemon for {"host": [%s], "protocol": [%s], "port":[%v]}`, hosts, protocols, ports)
			return true
		}

		logger.Trace(`Unable to find daemon for {"host": [%s], "protocol": [%s], "ip": [%s], "port":[%v], with expression [%s]`, daemon.Host, daemon.Protocol, daemon.IPVersion, daemon.Port, expression)
		return false
	}

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
						logger.Debug("No daemon is listening on port [%d], IP version [%s] and transport protocol [%s]", pt, daemon.IPVersion[i], p)
						// Fail on first failed check
						return false
					} else {
						logger.Trace(`Found daemon for {"host": [%s], "protocol": [%s], "ip": [%s], "port":[%v]}`, h, p, ip, pt)
					}
				}
			}
		}
	}
	// All has passed
	return true
	*/
}
