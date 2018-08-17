package checks

import (
	"bufio"
	"fmt"
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
	"gitlab.cern.ch/lb-experts/golbclient/lbalias/utils/parser"
	"os"
	"regexp"
	"strings"
)

type DaemonListening struct {
	Port int
	Protocol protocol
	IPVersion ipLevel
}


// daemonEntry : helper struct used to parse the metric line
type daemonEntry struct {
	prot protocol
	ipv  ipLevel
	port int
}

// protocol : int type used to distinguish between different transport protocol types
type protocol int

// ipLevel : int type used to distinguish between different IP levels
type ipLevel int

// Enum for the transport protocol types supported
const (
	TCPUDP protocol = iota
	TCP
	UDP
)

// Enum for the transport protocol types supported
const (
	IPV4V6 ipLevel = iota
	IPV4
	IPV6
	ANY
)

// isListening : checks if the given proc/daemon is running on the required port
func (daemon DaemonListening) scanNetwork(proc string, port int) bool {
	file, err := os.Open(proc)
	if err != nil {
		logger.Error("Error opening [%s]", proc)
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// The format of the file is 'sl  local_address rem_address   st'

	portHex := fmt.Sprintf("%04x", port)

	logger.Debug("Scanning port [%s]", portHex)

	//portOpen, _ := regexp.Compile("^ *[0-9]+: [0-9A-F]+:" + portHex + " [0-9A-F]+:[0-9A-F]+ 0A")

	portOpen, _ := regexp.Compile("(?i)(^ *[0-9]+: [0-9A-F]+:" + portHex + " [0-9A-F]+:[0-9A-F]+ 0)")
	for scanner.Scan() {
		line := scanner.Text()
		if portOpen.MatchString(line) {
			logger.Trace("Found an open port number [%s] open in [%v]", portHex, line)
			return true
		}
	}
	return false
}

func (daemon DaemonListening) Run(args ...interface{}) interface{} {
	entry := daemonEntry{}
	if daemon.Port != 0 && daemon.Protocol != 0 && daemon.IPVersion != 0 {
		// Backwards compatibility with v1.0
		entry = daemonEntry{prot :daemon.Protocol, ipv: daemon.IPVersion, port: daemon.Port}
	} else {
		// Process the daemon metric
		ent, err := daemon.processDaemonMetric(args[0].(string))
		entry = ent
		if err != nil {
			return false
		}
	}

	// Check if the given port is within bounds
	if (entry.port < 1) || (entry.port > 65535) {
		logger.Error("The specified port is out of range [1-65535]")
		return false
	}
	return daemon.isListening(entry.prot, entry.ipv, entry.port)
}

// isListening : checks if a daemon is listening on the given protocol(s) in the selected IP level and port
func (daemon DaemonListening) isListening(prot protocol, ipv ipLevel, port int) bool {
	switch prot {
	case TCP:
		switch ipv {
		case IPV4:
			if daemon.scanNetwork("/proc/net/tcp", port) {
				logger.Debug("The daemon is listening on the transport protocol type [tcp/v4] with the port [%d]", port)
				return true
			}
			logger.Debug("The daemon is not listening on the transport protocol type [tcp/v4] with the port [%d]", port)
		case IPV6:
			if daemon.scanNetwork("/proc/net/tcp6", port) {
				logger.Debug("The daemon is listening on the transport protocol type [tcp/v6] with the port [%d]", port)
				return true
			}
			logger.Debug("The daemon is not listening on the transport protocol type [tcp/v6] with the port [%d]", port)
		case IPV4V6:
			if daemon.scanNetwork("/proc/net/tcp", port) && daemon.scanNetwork("/proc/net/tcp6", port) {
				logger.Debug("The daemon is listening on the transport protocol type [tcp/all] with the port [%d]", port)
				return true
			}
			logger.Debug("The daemon is not listening on the transport protocol type [tcp/all] with the port [%d]", port)
		case ANY:
			if daemon.scanNetwork("/proc/net/tcp", port) || daemon.scanNetwork("/proc/net/tcp6", port) {
				logger.Debug("The daemon is listening on the transport protocol type [tcp] with the port [%d]", port)
				return true
			}
			logger.Debug("The daemon is not listening on the transport protocol type [tcp] with the port [%d]", port)
		}
	case UDP:
		switch ipv {
		case IPV4:
			if daemon.scanNetwork("/proc/net/udp", port) {
				logger.Debug("The daemon is listening on the transport protocol type [udp/v4] with the port [%d]", port)
				return true
			}
			logger.Debug("The daemon is not listening on the transport protocol type [udp/v4] with the port [%d]", port)
		case IPV6:
			if daemon.scanNetwork("/proc/net/udp6", port) {
				logger.Debug("The daemon is listening on the transport protocol type [udp/v6] with the port [%d]", port)
				return true
			}
			logger.Debug("The daemon is not listening on the transport protocol type [udp/v6] with the port [%d]", port)
		case IPV4V6:
			if daemon.scanNetwork("/proc/net/udp", port) && daemon.scanNetwork("/proc/net/udp6", port) {
				logger.Debug("The daemon is listening on the transport protocol type [udp/all] with the port [%d]", port)
				return true
			}
			logger.Debug("The daemon is not listening on the transport protocol type [udp/all] with the port [%d]", port)
		case ANY:
			if daemon.scanNetwork("/proc/net/udp", port) || daemon.scanNetwork("/proc/net/udp", port) {
				logger.Debug("The daemon is listening on the transport protocol type [udp] with the port [%d]", port)
				return true
			} else {
				logger.Debug("The daemon is not listening on the transport protocol type [udp] with the port [%d]", port)
			}
		}
	case TCPUDP:
		return daemon.isListening(TCP, ipv, port) && daemon.isListening(UDP, ipv, port)
	}
	return false
}

// processDaemonMetric : extracts the protocol, IP level and port from the given metric line
func (daemon DaemonListening) processDaemonMetric(line string) (de daemonEntry, err error) {
	// Log
	logger.Debug("Processing [daemon] metric entry line [%s]", line)
	found := regexp.MustCompile("(?i)((check)( )+(daemon))").Split(strings.TrimSpace(line), -1)
	if len(found) != 2 {
		// Log
		logger.Error("Incorrect syntax specified at [%s]. Please use (`check` `daemon` `<protocol><ip_level (optional)>:<port>`). Failing with code [%d]", line)
		return de, err
	}

	// Early port number syntax check
	if !strings.Contains(line, ":") {
		// Log
		logger.Error("Incorrect syntax specified at [%s]. Please use (`check` `daemon` `<protocol><ip_level (optional)>:<port>`). Failing with code [%s]", line)
		return de, err
	}

	rawMetricLine := found[1]
	//Log
	logger.Debug("Found metric line [%v]", rawMetricLine)

	// If no expression was given, fail the whole expression
	if len(strings.TrimSpace(rawMetricLine)) == 0 {
		logger.Error("Detected a (check) metric-type without any given metric. Return [false]")
		return de, err
	}

	// Discover the desired transport protocol & IP level (i.e. first set of paired squared parenthesis)
	rawProtocol := regexp.MustCompile(`\[([^\[\]]*)\]`).FindAllString(rawMetricLine, -1)[0]
	if len(strings.TrimSpace(rawProtocol)) == 2 {
		logger.Error("Incorrect syntax was detected when processing the metric line [%s]. The desired transport protocol must be specified (e.g. udp/tcp/all)")
		return de, err
	}

	/************************************************** Syntax checks ***********************************************/

	// Discover if an IP version was given
	ipVersion := regexp.MustCompile("/").Split(rawProtocol, -1)
	if len(ipVersion) > 1 {
		parsedLC := strings.ToLower(ipVersion[1])
		if parsedLC == "v4" {
			de.ipv = IPV4
		} else if parsedLC == "v6" {
			de.ipv = IPV6
		} else {
			// @TODO review syntax error failures in the application level
			logger.Error("Syntax error detected when attempting to extract the IP level from the metric line [%s]. The value [%s] is not supported. Defaulting to [all]", line, parsedLC)
			de.ipv = IPV4V6
		}
	} else {
		// Check IPv4 or IPv6 if nothing was given
		de.ipv = ANY
	}

	// Discover the type of transport protocol
	runes := []rune(rawProtocol)
	parsedLC := strings.ToLower(string(runes[1:4]))
	if parsedLC == "udp" {
		de.prot = UDP
	} else if parsedLC == "tcp" {
		de.prot = TCP
	} else {
		// @TODO review syntax error failures in the application level
		logger.Error("Syntax error detected when attempting to extract the transport protocol from the metric line [%s]. The value [%s] is not supported. Defaulting to [all]", line, parsedLC)
		de.prot = TCPUDP
	}

	// Discover the port number
	rawPortNumber := regexp.MustCompile(`:`).Split(line, -1)
	logger.Trace("Port Daemon port raw extraction [%s]", rawPortNumber)
	if len(rawPortNumber) != 2 {
		logger.Error("Syntax error detected when attempting to extract the port number from the metric line [%s]. The port number must be supplied using the ':' character at the end of your metric", line)
		return de, err
	}

	// Assign port
	de.port = int(parser.ParseInterfaceAsInteger(rawPortNumber[1]))

	return de, nil
}
