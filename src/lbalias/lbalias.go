package lbalias

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"syscall"
)

const ACCEPTABLE_BLOCK_RATE = 0.90
const ACCEPTABLE_INODE_RATE = 0.95

type LBalias struct {
	Name       string
	ConfigFile string
	Debug      bool
	NoLogin    bool
	Syslog     bool
	GData      []string
	MData      []string
	//	Metric     string
	//	CheckSwapping bool
	//	CheckXsessions bool
	//	CheckMetricList []int
	//	LoadMetricList []int
	//	ConstantList []int
}
type LBcheck struct {
	Code     int
	Function func(LBalias) bool
}

func daemonListening(port int, proc string) bool {
	/* try:
	       f_in=open(procfile,"r")
	   except:
	       return False
	   else:
	       header = f_in.readline()
	       for line in f_in:
	           words = line.split()
	           (localaddr, localport)   = words[1].split(':')
	           (remoteaddr, remoteport) = words[2].split(':')
	           status = words[3]
	           localport = int(localport,16)
	           status    = int(status,16)
	           listening = int('0x0a',16)
	           if (localport == port) and (status == listening):
	               return True
	   finally:
	       try:
	           f_in.close()
	       except NameError:
	           pass*/
	return false

}
func checkDaemon(port int) func(LBalias) bool {
	return func(lbalias LBalias) bool {

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
					fmt.Printf("[check_daemon %s] daemon on port %d is not listening\n", p, port)

				}
			} else {
				fmt.Printf("[check_daemon] daemon on port %d is not listening\n", port)
			}
		}
		return notok
	}
}

func checkTmpFull(lbalias LBalias) bool {
	var stat syscall.Statfs_t

	syscall.Statfs("/tmp", &stat)
	blockLevel := 1 - (float64(stat.Bavail) / float64(stat.Blocks))
	inodeLevel := 1 - (float64(stat.Ffree) / float64(stat.Files))
	if lbalias.Debug {
		fmt.Printf("[check_tmpfull] blocks occupancy: %.2f%% inodes occupancy: %.2f%%\n", blockLevel*100, inodeLevel*100)
	}
	return ((blockLevel > ACCEPTABLE_BLOCK_RATE) || (inodeLevel > ACCEPTABLE_INODE_RATE))
}

func checkAFS(lbalias LBalias) bool {
	return true
}
func checkNoLogin(lbalias LBalias) bool {
	if lbalias.NoLogin {
		return false
	}
	nologin := [2]string{"/etc/nologin", "/etc/iss.nologin"}

	if lbalias.Name != "" {
		nologin[1] += "." + lbalias.Name
	}
	fmt.Println("Checking if the files exist", nologin)
	for _, file := range nologin {
		fmt.Println("Checking if file exists", file)
		_, err := os.Stat(file)

		if err == nil {
			if lbalias.Debug {
				fmt.Printf("[check_nologin] %s present\n", file)
			}

			return true
		}
	}
	if lbalias.Debug {
		fmt.Printf("[check_nologin] users allowed to log in\n")

	}
	return false

}

var allLBchecks = map[string]LBcheck{
	"NOLOGIN":   LBcheck{1, checkNoLogin},
	"AFS":       LBcheck{10, checkAFS},
	"WEBDAEMON": LBcheck{8, checkDaemon(80)},
        "SSHDAEMON": LBcheck{7, checkDaemon(22)},
        "FTPDAEMON": LBcheck{9, checkDaemon(21)},
        "GRIDFTPDAEMON": LBcheck{11, checkDaemon(2811)},
	"TMPFULL":   LBcheck{6, checkTmpFull}}

func (lbalias LBalias) Evaluate() int {
	fmt.Println("Evaluating the alias", lbalias)

	checks := []string{}
	for key, _ := range allLBchecks {
		checks = append(checks, "("+key+")")
	}
	//	fmt.Println(checks)
	f, err := os.Open(lbalias.ConfigFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	fmt.Println("Configuration file opened")
	comment, _ := regexp.Compile("^(#.*)?$")
	actions, _ := regexp.Compile("(?i)^CHECK (" + strings.Join(checks, "|") + ")")
	constant, _ := regexp.Compile("(?i)^LOAD (LEMON)|(CONSTANT) (.*)")
	for scanner.Scan() {
		line := scanner.Text()
		if comment.MatchString(line) {
			//fmt.Println("THERE IS A COMMENT")
			continue
		}
		actions := actions.FindStringSubmatch(line)
		if len(actions) > 0 {
			//fmt.Println(actions)
			//fmt.Println(actions[1])
			//fmt.Println(actions[1])

			myAction := strings.ToUpper(actions[1])
			//			fmt.Println("THERE IS AN ACTION", myAction)
			//if  allLBchecks[myAction].Function != nil {
			// fmt.Println("CHECKING THE FUNCTION")
			if allLBchecks[myAction].Function(lbalias) {
				fmt.Println("THE CHECK OF ", myAction, "FAILED ")
				return -allLBchecks[myAction].Code
			}
			continue
		}
		//}
		if constant.MatchString(line) {
			fmt.Println("THERE IS A CONSTANT")
			continue

		}
		fmt.Println("We can't parse the configuration line", line)

	}
	return 0
}
