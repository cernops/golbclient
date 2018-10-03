package lbalias

import (
	"bufio"
	"fmt"

	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

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
	Metric         int
	//	CheckSwapping bool
	//	CheckXsessions bool
	CheckMetricList []MetricEntry
	LoadMetricList  []MetricEntry
	Constant        float32
	CheckAttributes map[string]bool
}

func (self LBalias) String() string {
	return fmt.Sprintf("Alias name %s with load of %f and loadmetrics %v", self.Name, self.Constant, self.LoadMetricList)
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
	"LEMON":         LBcheck{12, checkLemon("check")},
	"ROGER":         LBcheck{13, checkRoger},
	"COMMAND":       LBcheck{14, checkCommand},
	//The rest are a bit special. They don't return immediately
	"XSESSIONS": LBcheck{0, checkAttribute("xsession")},
	"SWAPPING":  LBcheck{0, checkAttribute("swapping")},
	"LEMONLOAD": LBcheck{17, checkLemon("load")}}

//These are all the current tests
//
//

// The lemon metrics are done in two steps: The first one is to add all of them to the configuration
// The second step is to call all of them in one go
func checkAttribute(name string) func(*LBalias, string) bool {

	return func(lbalias *LBalias, line string) bool {
		if lbalias.CheckAttributes == nil {
			lbalias.CheckAttributes = map[string]bool{}
		}
		lbalias.DebugMessage("[check_attribute] Checking the attribute ", name)
		lbalias.CheckAttributes[name] = true

		return false
	}
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
func (lbalias *LBalias) Evaluate() {
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
	constant, _ := regexp.Compile("(?i)^LOAD ((LEMON)|(CONSTANT)) (.*)")
	for scanner.Scan() {
		line := scanner.Text()
		if comment.MatchString(line) {
			continue
		}
		actions := actions.FindStringSubmatch(line)
		if len(actions) > 0 {

			myAction := strings.ToUpper(actions[1])

			if allLBchecks[myAction].Function(lbalias, line) {
				lbalias.DebugMessage("THE CHECK OF ", myAction, "FAILED ")
				lbalias.Metric = -allLBchecks[myAction].Code
				return
			}

			continue
		}
		constants :=
			constant.FindStringSubmatch(line)
		if len(constants) > 0 {
			//fmt.Println(constants[1])
			if strings.ToUpper(constants[1]) == "LEMON" {
				if allLBchecks["LEMONLOAD"].Function(lbalias, line) {
					fmt.Println("Error adding the lemon metric for the load")
					lbalias.Metric = -allLBchecks["LEMONLOAD"].Code
					return
				}
			} else {
				if lbalias.addConstant(constants[4]) {
					lbalias.Metric = -16
					return
				}
			}
			continue

		}
		fmt.Println("We can't parse the configuration line", line)

	}
	if len(lbalias.CheckMetricList) > 0 {
		if lbalias.checkLemonMetric() {
			lbalias.DebugMessage("[main] Lemon metric check failed")
			lbalias.Metric = -allLBchecks["LEMON"].Code
			return
		}
	}
	lbalias.Metric = int(lbalias.Constant)

	if len(lbalias.LoadMetricList) > 0 {
		lemon_load := lbalias.evaluateLoadLemon()
		if lemon_load < 0 {
			fmt.Println("Lemon load returned negative!")
			lbalias.Metric = -allLBchecks["LEMONLOAD"].Code
			return
		}
		lbalias.Metric += lemon_load
	}
	if lbalias.Metric == 0 {
		lbalias.DebugMessage("Default method to calculate the load")
		lbalias.Metric = lbalias.defaultLoad()
	}
}
func (lbalias *LBalias) addConstant(exp string) bool {
	lbalias.DebugMessage("[add_constant] Adding Constant ", exp)
	f, err := strconv.ParseFloat(exp, 32)
	if err != nil {
		fmt.Println("Error parsing the number from ", exp)
		return true
	}
	fmt.Println("[add_constant] value=", f)
	lbalias.Constant += float32(f)
	return false
}

func (lbalias *LBalias) defaultLoad() int {

	swap := lbalias.swapFree()

	lbalias.DebugMessage(fmt.Sprintf("[main] result of swap formula = %f", swap))
	cpuload := lbalias.cpuLoad()
	lbalias.DebugMessage(fmt.Sprintf("[main] result of cpu formula = %f", cpuload))

	swaping := float32(0)
	if lbalias.CheckAttributes["swapping"] {
		//   swaping = stat_swaping()
		lbalias.DebugMessage(fmt.Sprintf("[main] result of swaping formula = %f", swaping))
	}

	f_sm, nb_processes, _ := lbalias.sessionManager()
	users := lbalias.countUsers()
	if lbalias.CheckAttributes["xsessions"] {
		lbalias.DebugMessage(fmt.Sprintf("[main] result of X sessions formula = %f", f_sm))
	} else {
		f_sm = float32(0)
	}

	lbalias.DebugMessage(fmt.Sprintf("[main] number of processes: %d", int(nb_processes)))

	lbalias.DebugMessage(fmt.Sprintf("[main] number of users logged in: %d", int(users)))

	myLoad := (((swap + users/25.) / 2.) + (2. * swaping) + (3. * cpuload) + (2. * f_sm)) / 6.

	//((swap + users / 25.) / 2.) + (2. * swaping * self.check_swaping) + (3. * cpuload) + (2. * f_sm * self.check_xsessions)) / 6.
	lbalias.DebugMessage(fmt.Sprintf("[main] LOAD = %f, swap = %.3f, users = %.0f, swaping = %.3f, cpuload = %.3f, f_sm = %.3f", myLoad, swap, users, swaping, cpuload, f_sm))
	return int(myLoad * 1000)

}

func (lbalias *LBalias) swapFree() float32 {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		fmt.Println("Error openning", "/proc/meminfo")
		return -2
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	memoryMap := map[string]int{}
	//fmt.Println("Looking for ", portHex)
	memory, _ := regexp.Compile("^((MemTotal)|(MemFree)|(SwapTotal)|(SwapFree)|(CommitLimit)|(Committed_AS)): +([0-9]+)")
	for scanner.Scan() {
		line := scanner.Text()
		match := memory.FindStringSubmatch(line)
		if len(match) > 0 {

			memoryMap[match[1]], _ = strconv.Atoi(match[8])
		}
	}
	lbalias.DebugMessage(fmt.Sprintf(
		"Mem:  %d %d\nCommit:  %d %d\nSwap: %d %d",
		memoryMap["MemTotal"], memoryMap["MemFree"], memoryMap["CommitLimit"],
		memoryMap["Committed_AS"], memoryMap["SwapTotal"], memoryMap["SwapFree"]))

	if memoryMap["SwapTotal"] == 0 {
		memoryMap["SwapTotal"], memoryMap["SwapFree"] = memoryMap["MemTotal"], memoryMap["MemFree"]
	}
	// recalculate swap numbers in Gbytes
	memoryMap["SwapTotal"] = memoryMap["SwapTotal"] / (1024 * 1024)
	memoryMap["SwapFree"] = memoryMap["SwapFree"] / (1024 * 1024)

	if (100*memoryMap["SwapFree"] < 75*memoryMap["SwapTotal"]) ||
		(100*memoryMap["Committed_AS"] > (75 * memoryMap["CommitLimit"])) {
		return 5
	}
	if memoryMap["SwapTotal"] == 0 {
		return 0
	}
	return (21 - (20. * float32(memoryMap["SwapFree"]) / float32(memoryMap["SwapTotal"]))) / 6.
}

func (lbalias *LBalias) cpuLoad() float32 {
	file, err := os.Open("/proc/loadavg")
	if err != nil {
		fmt.Println("Error openning", "/proc/loadavg")
		return -2
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	line := scanner.Text()
	cpu := strings.Split(line, " ")

	cpuFloat, _ := strconv.ParseFloat(cpu[0], 32)
	return float32(cpuFloat / 10.)
} /*


def stat_swaping():
    P1 = get_pagecounters()
    time.sleep(2)
    P2 = get_pagecounters()
    P=abs((P2-P1)/2)
    if (P>5000):
        m=1.
    else:
        m=P
        m=m/5000.
    if debug:
        print m
    return(m)*/

func (lbalias *LBalias) sessionManager() (float32, float32, float32) {

	out, err := exec.Command("/bin/ps", "auxw").Output()

	if err != nil {
		fmt.Println("Error executing the ps command!", err)
		return -10, -10, -10
	}

	//Let's parse the output, and collect the number of processes
	f_sm, nb_processes := 0.0, -1.0
	users := map[string]bool{}
	//There are 3 processes per gnome sesion, and 4 for the fvm
	gnome, _ := regexp.Compile("^([^ ]+ +){10}[^ ]*((gnome-session)|(kdesktop))")
	fvm, _ := regexp.Compile("^([^ ]+ +){10}[^ ]*fvwm")
	user, _ := regexp.Compile("^([^ ]+)")

	for _, line := range strings.Split(string(out), "\n") {
		nb_processes++
		if gnome.MatchString(line) {
			f_sm += 1 / 3.
		}
		if fvm.MatchString(line) {
			f_sm += 1 / 4.
		}
		a := user.FindStringSubmatch(line)
		if len(a) > 0 {
			users[a[1]] = true
		}

	}
	return float32(f_sm), float32(nb_processes), float32(len(users) - 1)
}

func (lbalias *LBalias) countUsers() float32 {
	utmpSlice, err := readAllUtmpEntries(LoginProcess, UserProcess)
	if err != nil {
		fmt.Println("Error reading the utmp file!", err)
		return -10
	} else {
		return float32(len(utmpSlice) - 1)
	}
}
