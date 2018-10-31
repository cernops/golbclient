package lbalias

import (
	"fmt"
	"gitlab.cern.ch/lb-experts/golbclient/lbalias/checks"
	"gitlab.cern.ch/lb-experts/golbclient/lbalias/utils/filehandler"
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type LBalias struct {
	Name           string
	ConfigFile     string
	//NoLogin        bool
	Syslog         bool
	GData          []string
	MData          []string
	//CheckXsessions int
	//RogerState     string
	Metric         int
	//CheckMetricList []MetricEntry
	//LoadMetricList  []MetricEntry
	Constant        float32
	//CheckAttributes map[string]bool
	ChecksDone      map[string]bool
}

type ExpressionCode struct {
	code int
	cli CLI
}

// @TODO: add values to the wiki page: http://configdocs.web.cern.ch/configdocs/dnslb/lbclientcodes.html
var allLBExpressions = map[string] ExpressionCode{
	"NOLOGIN":       {code: 1, cli: checks.NoLogin{}},
	"TMPFULL":       {code: 6, cli: checks.TmpFull{}},
	"SSHDAEMON":     {code: 7, cli: checks.Listening{Port: []int{22}, Protocol: []string{"tcp"}, IPVersion: []string{"ipv4"}}},
	"WEBDAEMON":     {code: 8, cli: checks.Listening{Port: []int{80}, Protocol: []string{"tcp"}, IPVersion: []string{"ipv4"}}},
	"FTPDAEMON":     {code: 9, cli: checks.Listening{Port: []int{21}, Protocol: []string{"tcp"}, IPVersion: []string{"ipv4"}}},
	"GRIDFTPDAEMON": {code: 11, cli: checks.Listening{Port: []int{2811}, Protocol: []string{"tcp"}, IPVersion: []string{"ipv4"}}},
	"DAEMON":		 {code: 7, cli: checks.Listening{}},
	"AFS":           {code: 10, cli: checks.AFS{}},
	"LEMON":         {code: 12, cli: checks.ParamCheck{Command: "lemon"}},
	"LEMONLOAD":     {code: 12, cli: checks.ParamCheck{Command: "lemon"}},
	"ROGER":         {code: 13, cli: checks.RogerState{}},
	"COMMAND":       {code: 14, cli: checks.Command{}},
	"COLLECTD":      {code: 15, cli: checks.ParamCheck{Command: "collectd"}},
	"COLLECTDLOAD":  {code: 15, cli: checks.ParamCheck{Command: "collectd"}},
	"XSESSIONS": {code: 6, cli: checks.CheckAttribute{}},
	"SWAPPING":  {code: 6, cli: checks.CheckAttribute{}},
}

// Evaluate : Evaluates a [lbalias] entry
func (lbalias *LBalias) Evaluate() error {
	logger.Debug("Evaluating the alias [%s]", lbalias.Name)

	// Create a string array containing all the checks to be performed
	var checks []string
	for key := range allLBExpressions {
		checks = append(checks, fmt.Sprintf("(%s)", key))
	}

	// Attempt to read the configuration file
	logger.Debug("Attempting to read the configuration file [%s]", lbalias.ConfigFile)
	lines, err := filehandler.ReadAllLinesFromFile(lbalias.ConfigFile)
	if err != nil {
		logger.Error("Fatal error when attempting to open the alias configuration file [%s]", err.Error())
		return err		
	}

	// Read the configuration file using the scanner API
	logger.Debug("Successfully opened the alias configuration file")

	// Detect all comments
	comment, _ := regexp.Compile("^(#.*)?$")
	// Detect all checks to be made
	actions, _ := regexp.Compile("(?i)^CHECK (" + strings.Join(checks, "|") + ")")
	// Detect all loads to be made
	constant, _ := regexp.Compile("(?i)^LOAD ((LEMON)|(COLLECTD)|(CONSTANT))( )*(.*)")

	// Read the configuration file line-by-line
	for _, line := range lines {
		if comment.MatchString(line) {
			continue
		}
		checks := actions.FindStringSubmatch(line)
		if len(checks) > 0 {
			/********************************** CHECKS **********************************/
			myAction := strings.ToUpper(checks[1])
			if b, ok := allLBExpressions[myAction].cli.Run(line, lbalias.Name).(bool); !b || !ok {
				lbalias.Metric = -allLBExpressions[myAction].code
				logger.Error("The check of [%s] failed. Aborting with the code [%d]", myAction, lbalias.Metric)
				return nil
			}
			lbalias.ChecksDone[checks[1]] = true
			continue
		}
		loads := constant.FindStringSubmatch(line)
		if len(loads) > 0 {
			var result int
			/********************************** LOADS **********************************/
			cliName := strings.ToUpper(loads[1])
			if cliName == "LEMON" || cliName == "COLLECTD" {
				result = int(allLBExpressions[cliName].cli.Run(line, lbalias.Name).(int32))
				if result == -1 {
					logger.Error("[%s] metric returned a negative number [%d]", cliName, result)
					lbalias.Metric = -allLBExpressions[cliName].code
					return nil
				}
			} else {
				if lbalias.addConstant(loads[4]) {
					lbalias.Metric = -20
					return nil
				}
			}
			// Added metric value to the total in case of no problems
			lbalias.Metric += result
			continue
		}

		// If none of the regexs were found, then it is assumed that there is a user-made mistake in the configuration file
		logger.Error("Unable to parse the configuration file line [%s]", line)

	}
	// Log
	logger.Trace("Final metric value [%d]", lbalias.Metric)

	if lbalias.Metric == 0 {
		logger.Debug("No metric value was found. Defaulting to the generic load calculation")
		lbalias.Metric = lbalias.defaultLoad()
	}
	return nil
}
func (lbalias *LBalias) addConstant(exp string) bool {
	logger.Debug("Adding Constant [%s]", exp)
	// @TODO: Replace with the parser.ParseInterfaceAsFloat (reflection?)
	f, err := strconv.ParseFloat(exp, 32)
	if err != nil {
		logger.Error("Error parsing the floating point number from the value [%s]", exp)
		return true
	}

	logger.Debug("\tValue = [%d]", f)
	lbalias.Constant += float32(f)
	return false
}

func (lbalias *LBalias) defaultLoad() int {
	swap := lbalias.swapFree()
	logger.Debug("Result of swap formula = %f", swap)
	cpuload := lbalias.cpuLoad()
	logger.Debug("Result of cpu formula = %f", cpuload)
	swaping := float32(0)
	if lbalias.ChecksDone["swapping"] {
		logger.Debug("Result of swapoing formula = %f", swaping)
	}

	f_sm, nb_processes, users := lbalias.sessionManager()
	if lbalias.ChecksDone["xsessions"]  {
		logger.Debug("Result of X sessions formula = %f", f_sm)
	} else {
		f_sm = float32(0)
	}

	logger.Debug("Number of processes = %d", int(nb_processes))
	logger.Debug("Number of users logged in = %d ", int(users))

	myLoad := (((swap + users/25.) / 2.) + (2. * swaping) + (3. * cpuload) + (2. * f_sm)) / 6.

	//((swap + users / 25.) / 2.) + (2. * swaping * self.check_swaping) + (3. * cpuload) + (2. * f_sm * self.check_xsessions)) / 6.
	logger.Debug("LOAD = %f, swap = %.3f, users = %.0f, swaping = %.3f, cpuload = %.3f, f_sm = %.3f", myLoad, swap, users, swaping, cpuload, f_sm)
	return int(myLoad * 1000)

}

func (lbalias *LBalias) swapFree() float32 {
	lines, err := filehandler.ReadAllLinesFromFile("/proc/meminfo")
	if err != nil {
		logger.Error("Error opening the file [%s]. Error [%s]", "/proc/meminfo", err.Error())
		return -2
	}
	memoryMap := map[string]int{}
	//fmt.Println("Looking for ", portHex)
	memory, _ := regexp.Compile("^((MemTotal)|(MemFree)|(SwapTotal)|(SwapFree)|(CommitLimit)|(Committed_AS)): +([0-9]+)")
	for _, line := range lines {
		match := memory.FindStringSubmatch(line)
		if len(match) > 0 {
			memoryMap[match[1]], _ = strconv.Atoi(match[8])
		}
	}
	logger.Debug("Mem:  %d %d\nCommit:  %d %d\nSwap: %d %d",
		memoryMap["MemTotal"], memoryMap["MemFree"], memoryMap["CommitLimit"],
		memoryMap["Committed_AS"], memoryMap["SwapTotal"], memoryMap["SwapFree"])

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
	line, err := filehandler.ReadFirstLineFromFile("/proc/loadavg")
	if err != nil {
		logger.Error("Error opening the file [%s]. Error [%s]", "/proc/loadavg", err.Error())
		return -2
	}
	cpu := strings.Split(line, " ")
	cpuFloat, _ := strconv.ParseFloat(cpu[0], 32)
	return float32(cpuFloat / 10.)
}

func (lbalias *LBalias) sessionManager() (float32, float32, float32) {

	out, err := exec.Command("/bin/ps", "auxw").Output()

	if err != nil {
		logger.Error("Error while executing the command [%s]. Error [%s]", "ps", err.Error())
		return -10, -10, -10
	}

	// Let's parse the output, and collect the number of processes
	f_sm, nb_processes := 0.0, -1.0
	users := map[string]bool{}
	// There are 3 processes per gnome sesion, and 4 for the fvm
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

