package lbalias

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

func daemonListening(port int, proc string) bool {
	file, err := os.Open(proc)
	if err != nil {
		fmt.Println("Error openning", proc)
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// The format of the file is 'sl  local_address rem_address   st'

	portHex := fmt.Sprintf("%04x", port)

	//fmt.Println("Looking for ", portHex)
	portOpen, _ := regexp.Compile("^ *[0-9]+: [0-9A-F]+:" + portHex + " [0-9A-F]+:[0-9A-F]+ 0A")
	for scanner.Scan() {
		line := scanner.Text()
		if portOpen.MatchString(line) {
			return true
		}
	}
	return false

}
func checkDaemon(port int) func(*LBalias, string) bool {
	return func(lbalias *LBalias, line string) bool {

		if (port < 1) || (port > 65535) {
			return true
		}
		notok := true
		listen := []string{}
		if daemonListening(port, "/proc/net/tcp") {
			notok = false
			listen = append(listen, "ipv4")
		}
		if daemonListening(port, "/proc/net/tcp6") {
			notok = false
			listen = append(listen, "ipv6")
		}

		if lbalias.Debug {
			if len(listen) > 0 {
				for _, p := range listen {
					fmt.Printf("[check_daemon %s] daemon on port %d is listening\n", p, port)

				}
			} else {
				fmt.Printf("[check_daemon] daemon on port %d is not listening\n", port)
			}
		}
		return notok
	}
}
