package lbconfig

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/checks"
	param "gitlab.cern.ch/lb-experts/golbclient/lbconfig/checks/parameterized"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/mapping"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/utils/filehandler"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/utils/timer"
)

// ExpressionCode : return value for the CLI calls
type ExpressionCode struct {
	code int
	cli  CLI
}

// @TODO: add values to the wiki page: http://configdocs.web.cern.ch/configdocs/dnslb/lbclientcodes.html
var allLBExpressions = map[string]ExpressionCode{
	"NOLOGIN":       {code: 1, cli: checks.NoLogin{}},
	"TMPFULL":       {code: 6, cli: checks.TmpFull{}},
	"SSHDAEMON":     {code: 7, cli: checks.DaemonListening{Port: 22}},
	"WEBDAEMON":     {code: 8, cli: checks.DaemonListening{Port: 80}},
	"FTPDAEMON":     {code: 9, cli: checks.DaemonListening{Port: 21}},
	"AFS":           {code: 10, cli: checks.AFS{}},
	"GRIDFTPDAEMON": {code: 11, cli: checks.DaemonListening{Port: 2811}},
	"LEMON":         {code: 12, cli: checks.ParamCheck{Type: param.LemonImpl{}}},
	"LEMONLOAD":     {code: 12, cli: checks.ParamCheck{Type: param.LemonImpl{}}},
	"ROGER":         {code: 13, cli: checks.RogerState{}},
	"COMMAND":       {code: 14, cli: checks.Command{}},
	"COLLECTD":      {code: 15, cli: checks.ParamCheck{Type: param.CollectdImpl{}}},
	"COLLECTDLOAD":  {code: 15, cli: checks.ParamCheck{Type: param.CollectdImpl{}}},
	"XSESSIONS":     {code: 6, cli: checks.CheckAttribute{}},
	"SWAPPING":      {code: 6, cli: checks.CheckAttribute{}},
}

/*
	And here we add the methods of the class
*/

// Evaluate : Evaluates a [lbalias] entry
func Evaluate(cm *mapping.ConfigurationMapping, timeout time.Duration, checkConfig bool) error {
	logger.Debug("Evaluating the configuration file [%s] for aliases [%v]", cm.ConfigFilePath, cm.AliasNames)

	// Create a string array containing all the checksToExecute to be performed
	var checksToExecute []string
	for key := range allLBExpressions {
		checksToExecute = append(checksToExecute, fmt.Sprintf("(%s)", key))
	}

	// Attempt to read the configuration file

	lines, err := filehandler.ReadAllLinesFromFile(cm.ConfigFilePath)
	if err != nil {
		logger.Error("Fatal error when attempting to open the alias configuration file [%s]", err.Error())
		return err
	}
	// Read the configuration file using the scanner API
	logger.Debug("Successfully opened the alias configuration file [%v]", cm.ConfigFilePath)

	// Detect all comments
	comment, _ := regexp.Compile("^[ \t]*(#.*)?$")
	// Detect all checksToExecute to be made
	actions, _ := regexp.Compile("(?i)^CHECK (" + strings.Join(checksToExecute, "|") + ")")
	// Detect all loads to be made
	constant, _ := regexp.Compile("(?i)^LOAD ((LEMON)|(COLLECTD)|(CONSTANT))( )*(.*)")

	// Read the configuration file line-by-line
	for _, line := range lines {
		if comment.MatchString(line) {
			continue
		}
		runningChecks := actions.FindStringSubmatch(line)
		if len(runningChecks) > 0 {
			/********************************** CHECKS **********************************/
			myAction := strings.ToUpper(runningChecks[1])
			negRet := -allLBExpressions[myAction].code
			logger.Info("Doing ", myAction)
			ret, err := timer.ExecuteWithTimeoutR(timeout, allLBExpressions[myAction].cli.Run, line, cm.AliasNames, cm.Default)
			logger.Info("Do we have an error, ", ret, err)
			if err != nil {
				cm.MetricValue = negRet
				return err
			}
			if ret.([]interface{})[0] == false && !checkConfig {
				cm.MetricValue = negRet
				return nil
			}
			continue
		}
		loads := constant.FindStringSubmatch(line)
		if len(loads) > 0 {
			/********************************** LOADS **********************************/
			cliName := strings.ToUpper(loads[1])
			negRet := -allLBExpressions[cliName].code
			if cliName == "LEMON" || cliName == "COLLECTD" {
				ret, err := timer.ExecuteWithTimeoutR(timeout, allLBExpressions[cliName].cli.Run, line)
				if err != nil {
					cm.MetricValue = negRet
					return fmt.Errorf("metric [%s] returned a negative number [%v]", cliName, ret)
				}
				if len(ret.([]interface{})) != 0 && !checkConfig {
					cm.MetricValue = negRet
					return nil
				}

				retVal := ret.([]interface{})[0].(int32)
				cm.MetricValue += int(retVal)
			} else {
				constant := strings.TrimSpace(regexp.MustCompile("(?i)(load constant)").Split(loads[0], -1)[1])
				if !cm.AddConstant(constant) {
					return fmt.Errorf("failed to load the constant value [%v]", constant)
				}
			}
			// Added metric value to the total in case of no problems
			continue
		}
		// If none of the regexs were found, then it is assumed that there is a user-made mistake in the configuration file
		logger.Error("Unable to parse the configuration file line [%s]", line)
		cm.MetricValue = -1
		return fmt.Errorf("Unable to parse the configuration file line [%s]", line)
	}
	if checkConfig {
		logger.Info("The configuration file is correct")
		return nil
	}

	if cm.MetricValue == 0 {
		logger.Info("No metric value was found. Defaulting to the generic load calculation")
		cm.MetricValue = defaultLoad()
	}

	// Log
	logger.Trace("Final metric value [%d]", cm.MetricValue)

	return nil
}

func defaultLoad() int {
	swap := swapFree()
	logger.Debug("Result of swap formula = %f", swap)
	cpuLoad := cpuLoad()
	logger.Debug("Result of cpu formula = %f", cpuLoad)
	swapping := float32(0)
	fSm, nbProcesses, users := sessionManager()
	logger.Debug("Number of processes = %d", int(nbProcesses))
	logger.Debug("Number of users logged in = %d ", int(users))
	myLoad := (((swap + users/25.) / 2.) + (2. * swapping) + (3. * cpuLoad) + (2. * fSm)) / 6.
	logger.Debug("LOAD = %f, swap = %.3f, users = %.0f, swapping = %.3f, "+
		"cpuLoad = %.3f, f_sm = %.3f", myLoad, swap, users, swapping, cpuLoad, fSm)
	return int(myLoad * 1000)

}

func swapFree() float32 {
	lines, err := filehandler.ReadAllLinesFromFile("/proc/meminfo")
	if err != nil {
		logger.Error("Error opening the file [%s]. Error [%s]", "/proc/meminfo", err.Error())
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
	logger.Debug("Mem:  %d %d\nCommit:  %d %d\nSwap: %d %d",
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
		logger.Error("Error opening the file [%s]. Error [%s]", "/proc/loadavg", err.Error())
		return -2
	}
	cpu := strings.Split(line, " ")
	cpuFloat, _ := strconv.ParseFloat(cpu[0], 32)
	return float32(cpuFloat / 10.)
}

func sessionManager() (float32, float32, float32) {
	out, err := exec.Command("/bin/ps", "auxw").Output()
	if err != nil {
		logger.Error("Error while executing the command [%s]. Error [%s]", "ps", err.Error())
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
