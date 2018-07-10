package checks

import (
	"bufio"
	"fmt"
	"lbalias/utils/logger"
	"os"
	"regexp"
)

type DaemonListening struct {
	code int
	port int
}

func (daemon DaemonListening) Code() int {
	return daemon.code
}

func (daemon DaemonListening) daemonListen(proc string) bool {
	file, err := os.Open(proc)
	if err != nil {
		logger.Error("Error opening [%s]", proc)
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// The format of the file is 'sl  local_address rem_address   st'

	portHex := fmt.Sprintf("%04x", daemon.port)

	logger.Debug("Scanning port [%s]", portHex)
	portOpen, _ := regexp.Compile("^ *[0-9]+: [0-9A-F]+:" + portHex + " [0-9A-F]+:[0-9A-F]+ 0A")
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
	rVal := false
	if (daemon.port < 1) || (daemon.port > 65535) {
		return false
	}
	listen := []string{}
	if daemon.daemonListen("/proc/net/tcp") {
		listen = append(listen, "ipv4")
		rVal = true
	}
	if daemon.daemonListen("/proc/net/tcp6") {
		listen = append(listen, "ipv6")
		rVal = true
	}

	if logger.GetLevel() == logger.DEBUG {
		if len(listen) > 0 {
			for _, p := range listen {
				logger.Debug("Daemon [%s] on port [%d] is listening", p, daemon.port)
			}
		} else {
			logger.Debug("Daemon is not listening in port [%d]", daemon.port)
		}
	}
	return rVal
}
