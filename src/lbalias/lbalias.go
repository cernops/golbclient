package lbalias

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const ACCEPTABLE_BLOCK_RATE = 0.90
const ACCEPTABLE_INODE_RATE = 0.95

type LBalias struct {
	Name           string
	ConfigFile     string
	Debug          bool
	NoLogin        bool
	Syslog         bool
	GData          []string
	MData          []string
	CheckXsessions int
	RogerState     string
	//	Metric     string
	//	CheckSwapping bool
	//	CheckXsessions bool
	CheckMetricList []MetricEntry
	//	LoadMetricList []int
	//	ConstantList []int
}
type LBcheck struct {
	Code     int
	Function func(*LBalias, string) bool
}

var allLBchecks = map[string]LBcheck{
	"NOLOGIN":       LBcheck{1, checkNoLogin},
	"TMPFULL":       LBcheck{6, checkTmpFull},
	"SSHDAEMON":     LBcheck{7, checkDaemon(22)},
	"WEBDAEMON":     LBcheck{8, checkDaemon(80)},
	"FTPDAEMON":     LBcheck{9, checkDaemon(21)},
	"AFS":           LBcheck{10, checkAFS},
	"GRIDFTPDAEMON": LBcheck{11, checkDaemon(2811)},
	"LEMON":         LBcheck{12, checkLemon},
	"ROGER":         LBcheck{13, checkRoger},
	//The rest are a bit special. They don't return immediately
	"XSESSIONS": LBcheck{0, checkXsession}}

//These are all the current tests
//
//

// The lemon metrics are done in two steps: The first one is to add all of them to the configuration
// The second step is to call all of them in one go
func checkXsession(lbalias *LBalias, line string) bool {
	lbalias.DebugMessage("[xsessions] Checking the xsessions")

	lbalias.CheckXsessions = 1
	return false
}
//
// And here we add the methods of the class
//
//

func (lbalias LBalias) DebugMessage(s ...interface{}) {
	if lbalias.Debug {
		fmt.Println(s)
	}

}
func (lbalias LBalias) Evaluate() int {
	lbalias.DebugMessage("[lbalias] Evaluating the alias " + lbalias.Name)

	checks := []string{}
	for key, _ := range allLBchecks {
		checks = append(checks, "("+key+")")
	}
	f, err := os.Open(lbalias.ConfigFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lbalias.DebugMessage("[lbalias] Configuration file opened")

	comment, _ := regexp.Compile("^(#.*)?$")
	actions, _ := regexp.Compile("(?i)^CHECK (" + strings.Join(checks, "|") + ")")
	constant, _ := regexp.Compile("(?i)^LOAD (LEMON)|(CONSTANT) (.*)")
	for scanner.Scan() {
		line := scanner.Text()
		if comment.MatchString(line) {
			continue
		}
		actions := actions.FindStringSubmatch(line)
		if len(actions) > 0 {

			myAction := strings.ToUpper(actions[1])

			if allLBchecks[myAction].Function(&lbalias, line) {
				fmt.Println("THE CHECK OF ", myAction, "FAILED ")
				return -allLBchecks[myAction].Code
			}

			continue
		}
		if constant.MatchString(line) {
			fmt.Println("THERE IS A CONSTANT")
			continue

		}
		fmt.Println("We can't parse the configuration line", line)

	}
	if len(lbalias.CheckMetricList) > 0 {
		if lbalias.checkLemonMetric() {
			lbalias.DebugMessage("[main] Lemon metric check failed")
			return -allLBchecks["LEMON"].Code
		}
	}
	return 0
}
