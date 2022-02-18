// +build linux darwin

package lbconfig

import (
	"fmt"
	logger "github.com/sirupsen/logrus"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/checks"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/checks/parameterized"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/mapping"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/utils/filehandler"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/utils/timer"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ExpressionCode : return value for the CLI calls
type ExpressionCode struct {
	code int
	cli  CLI
}

// @TODO: add values to the wiki page: http://configdocs.web.cern.ch/configdocs/dnslb/lbclientcodes.html
var allLBExpressions = map[string]ExpressionCode{
	"NOLOGIN": {code: 1, cli: checks.NoLogin{}},
	"TMPFULL": {code: 6, cli: checks.TmpFull{}},
	"SSHDAEMON": {code: 7, cli: checks.DaemonListening{Metric: `{"port": 22, 	"protocol": "tcp", "ip":["ipv4", "ipv6"]}`}},
	"WEBDAEMON": {code: 8, cli: checks.DaemonListening{Metric: `{"port": 80, 	"protocol": "tcp", "ip":["ipv4", "ipv6"]}`}},
	"FTPDAEMON": {code: 9, cli: checks.DaemonListening{Metric: `{"port": 21, 	"protocol": "tcp", "ip":["ipv4", "ipv6"]}`}},
	"AFS":             {code: 10, cli: checks.AFS{}},
	"GRIDFTPDAEMON":   {code: 11, cli: checks.DaemonListening{Metric: `{"port": 2811, "protocol": "tcp", "ip":["ipv4", "ipv6"]}`}},
	"LEMON":           {code: 12, cli: checks.ParamCheck{Impl: param.LemonImpl{}}},
	"LEMONLOAD":       {code: 12, cli: checks.ParamCheck{Impl: param.LemonImpl{}}},
	"ROGER":           {code: 13, cli: checks.RogerState{}},
	"COMMAND":         {code: 14, cli: checks.Command{}},
	"COLLECTD":        {code: 15, cli: checks.ParamCheck{Impl: param.CollectdImpl{}}},
	"COLLECTDLOAD":    {code: 15, cli: checks.ParamCheck{Impl: param.CollectdImpl{}}},
	"COLLECTD_ALARMS": {code: 15, cli: checks.ParamCheck{Impl: param.CollectdAlarmImpl{}, Type: "alarm"}},
	"CONSTANT":        {code: 16, cli: checks.MetricConstant{}},
	"DAEMON":          {code: 17, cli: checks.DaemonListening{}},
	"EOS":             {code: 18, cli: checks.EOS{}},
	"REBOOT":          {code: 19, cli: checks.Reboot{}},
	"XSESSIONS":       {code: 6, cli: checks.CheckAttribute{}},
	"SWAPPING":        {code: 6, cli: checks.CheckAttribute{}},
	"SWAPING":         {code: 6, cli: checks.CheckAttribute{}},
}

/*
	And here we add the methods of the class
*/

// Evaluate : Evaluates a [lbalias] entry
func Evaluate(cm *mapping.ConfigurationMapping, timeout time.Duration, checkConfig bool) error {
	contextLogger := logger.WithFields(logger.Fields{
		"EVALUATION":   "LOADING",
		"CFG_PATH":     cm.ConfigFilePath,
		"MAX_TIMEOUT":  timeout.String(),
		"CHECK_CONFIG": strconv.FormatBool(checkConfig),
	})

	contextLogger.Debug("Started the evaluation of the the configuration file...")

	// Create a string array containing all the checksToExecute to be performed
	var checksToExecute []string
	for key := range allLBExpressions {
		if key != "CONSTANT" {
			checksToExecute = append(checksToExecute, fmt.Sprintf("(%s)", key))
		}
	}

	// Attempt to read the configuration file

	lines, err := filehandler.ReadAllLinesFromFile(cm.ConfigFilePath)
	if err != nil {
		contextLogger.Errorf("Fatal error when attempting to open the alias configuration file [%s]", err.Error())
		return err
	}
	// Read the configuration file using the scanner API
	contextLogger.Debugf("Successfully opened the alias configuration file [%v]", cm.ConfigFilePath)

	// Detect all comments
	comment := regexp.MustCompile("^[ \t]*(#.*)?$")
	// Detect all actions (checks or loads) to be made
	checksFormat := "^[ ]*CHECK (" + strings.Join(checksToExecute, "|") + ")"
	loadsFormat := "^[ ]*LOAD ((LEMON)|(COLLECTD)|(CONSTANT))( )*(.*)"
	actions := regexp.MustCompile(fmt.Sprintf(`(?i)((%s)|(%s))`, checksFormat, loadsFormat))

	// Read the configuration file line-by-line
	for _, line := range lines {
		if comment.MatchString(line) {
			continue
		}
		foundActions := actions.FindStringSubmatch(line)
		if len(foundActions) > 0 {
			/********************************** ACTIONS **********************************/
			myAction := strings.ToUpper(strings.Split(line, " ")[1])
			if _, ok := allLBExpressions[myAction]; !ok {
				return fmt.Errorf("the given action (check or load) metric [%s] is not supported", myAction)
			}
			isLoad := regexp.MustCompile(`(?i)^LOAD`).MatchString(line)
			negRet := -allLBExpressions[myAction].code
			ret, err := timer.ExecuteWithTimeoutRInt(timeout, allLBExpressions[myAction].cli.Run,
				contextLogger.WithFields(logger.Fields{
					"CLI":        myAction,
					"EVALUATION": "ONGOING",
				}), line, cm.AliasNames, cm.Default)

			if err != nil {
				cm.MetricValue = negRet
				return err
			}
			if ret < 0 && !checkConfig {
				cm.MetricValue = negRet
				return nil
			}
			if isLoad {
				cm.MetricValue += ret
			}

		} else {
			// If none of the regexps were found, then it is assumed that there is a user-made mistake in the configuration file
			cm.MetricValue = -1
			return fmt.Errorf("unable to parse the configuration metric line [%s]. Stopping execution with "+
				"code [%d]", line, cm.MetricValue)
		}
	}

	if cm.MetricValue == 0 {
		contextLogger.Infof("No metric value was found. Defaulting to the generic load calculation")
		cm.MetricValue = defaultLoad()
	}

	// Log
	contextLogger.WithField("EVALUATION", "FINISHED").Tracef("Final metric value [%d]", cm.MetricValue)

	return nil
}

func defaultLoad() int {
	swap := swapFree()
	logger.Debugf("Result of swap formula = %f", swap)
	cpuLoad := cpuLoad()
	logger.Debugf("Result of cpu formula = %f", cpuLoad)
	swapping := float32(0)
	fSm, nbProcesses, users := sessionManager()
	logger.Debugf("Number of processes = %d", int(nbProcesses))
	logger.Debugf("Number of users logged in = %d ", int(users))
	myLoad := (((swap + users/25.) / 2.) + (2. * swapping) + (3. * cpuLoad) + (2. * fSm)) / 6.
	logger.Debugf("LOAD = %f, swap = %.3f, users = %.0f, swapping = %.3f, "+
		"cpuLoad = %.3f, f_sm = %.3f", myLoad, swap, users, swapping, cpuLoad, fSm)
	return int(myLoad * 1000)

}

func swapFree() float32 {
	lines, err := filehandler.ReadAllLinesFromFile("/proc/meminfo")
	if err != nil {
		logger.Errorf("Error opening the file [%s]. Error [%s]", "/proc/meminfo", err.Error())
		return -2
	}
	memoryMap := map[string]int{}
	memory := regexp.MustCompile("^((MemTotal)|(MemFree)|(SwapTotal)|(SwapFree)|(CommitLimit)|(Committed_AS)): +([0-9]+)")
	for _, line := range lines {
		match := memory.FindStringSubmatch(line)
		if len(match) > 0 {
			memoryMap[match[1]], _ = strconv.Atoi(match[8])
		}
	}
	logger.Debugf("Mem:  %d %d\nCommit:  %d %d\nSwap: %d %d",
		memoryMap["MemTotal"], memoryMap["MemFree"], memoryMap["CommitLimit"],
		memoryMap["Committed_AS"], memoryMap["SwapTotal"], memoryMap["SwapFree"])

	if memoryMap["SwapTotal"] == 0 {
		memoryMap["SwapTotal"], memoryMap["SwapFree"] = memoryMap["MemTotal"], memoryMap["MemFree"]
	}
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

func cpuLoad() float32 {
	line, err := filehandler.ReadFirstLineFromFile("/proc/loadavg")
	if err != nil {
		logger.Errorf("Error opening the file [%s]. Error [%s]", "/proc/loadavg", err.Error())
		return -2
	}
	cpu := strings.Split(line, " ")
	cpuFloat, _ := strconv.ParseFloat(cpu[0], 32)
	return float32(cpuFloat / 10.)
}

func sessionManager() (float32, float32, float32) {
	out, err := exec.Command("/bin/ps", "auxw").Output()
	if err != nil {
		logger.Errorf("Error while executing the command [%s]. Error [%s]", "ps", err.Error())
		return -10, -10, -10
	}

	// Let's parse the output, and collect the number of processes
	fSm, nbProcesses := 0.0, -1.0
	users := map[string]bool{}
	// There are 3 processes per gnome sesion, and 4 for the fvm
	gnome := regexp.MustCompile("^([^ ]+ +){10}[^ ]*((gnome-session)|(kdesktop))")
	fvm := regexp.MustCompile("^([^ ]+ +){10}[^ ]*fvwm")
	user := regexp.MustCompile("^([^ ]+)")

	for _, line := range strings.Split(string(out), "\n") {
		nbProcesses++
		if gnome.MatchString(line) {
			fSm += 1 / 3.
		}
		if fvm.MatchString(line) {
			fSm += 1 / 4.
		}
		a := user.FindStringSubmatch(line)
		if len(a) > 0 {
			users[a[1]] = true
		}

	}
	return float32(fSm), float32(nbProcesses), float32(len(users) - 1)
}
