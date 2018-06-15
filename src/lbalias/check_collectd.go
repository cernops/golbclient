package lbalias

import (
	"lbalias/utils/logger"
	"lbalias/utils/parser"
	"lbalias/utils/runner"
	"regexp"
	"strings"

	"github.com/Knetic/govaluate"
)

type Collectd struct {
	code int
}

func (c Collectd) Code() int {
	return c.code
}

// COLLECTD_CLI : Path to the [collectdctl] cli
const COLLECTD_CLI = "/usr/bin/collectdctl"

func (c Collectd) Run(args ...interface{}) interface{} {
	var rVal interface{}
	line := args[0].(string)
	isCheck := strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), "check")
	isLoad := strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), "load")

	// Log
	logger.LOG(logger.DEBUG, false, "Adding collectd metric [%s]", line)

	// Support unintentional errors => e.g., [loadcheck collectd], panics if the regex cannot be compiled
	found := regexp.MustCompile("(?i)((check|load)( )+collectd)").Split(strings.TrimSpace(line), -1)

	// Found the correct syntax
	if len(found) == 2 && (isCheck || isLoad) {
		rawExpression := found[1]

		// If no expression was given
		if len(strings.TrimSpace(rawExpression)) == 0 {
			rVal = secureFail(isCheck, isLoad)
			logger.LOG(logger.DEBUG, false, "Detected a (check|load) without metrics. Returning [%v]", rVal)
			return secureFail(isCheck, isLoad)
		}

		// Discover all the metrics found in the expression
		metrics := regexp.MustCompile(`\[([^\[\]]*)\]`).FindAllString(found[1], -1)
		parameters := make(map[string]interface{}, len(metrics))

		// Run [collectdctl] with a list of all the metrics found and return a key/value map
		c.runCollectdCommand(metrics, &parameters)

		// Parse the expression
		expression, err := govaluate.NewEvaluableExpression(rawExpression)

		if err != nil {
			logger.LOG(logger.ERROR, false, "Error when evaluating expression [%s]", err)
			return secureFail(isCheck, isLoad)
		}

		// Evaluate the expression
		result, err2 := expression.Evaluate(parameters)
		if err2 != nil {
			rVal := secureFail(isCheck, isLoad)
			logger.LOG(logger.ERROR, false, "Error when evaluating the parameters of the expression [%s]. Returning [%v]", err2, rVal)
			return rVal
		}

		// Debug
		logger.LOG(logger.DEBUG, false, "Expression returned result [%v]", result)
		if isCheck {
			/******************************** CHECK ************************************/
			rVal = parser.ParseInterfaceAsBool(result)
			logger.LOG(logger.TRACE, false, "Detected a [check] type metric, returning as boolean [%t]", rVal)
			return rVal
		} else if isLoad {
			/********************************* LOAD ************************************/
			rVal = parser.ParseInterfaceAsInteger(result)
			logger.LOG(logger.TRACE, false, "Detected a [load] type metric, returning as integer [%d]", rVal)
			return rVal
		} else {
			logger.LOG(logger.ERROR, false, "Failed to parse the result of the collectd expression [%v]", result)
			return false
		}
	}
	logger.LOG(logger.ERROR, false, "Incorrect syntax specified at [%s]. Please use (`load` or `check` `<cli>`)", line)
	return secureFail(isCheck, isLoad)
}

// secureFail : returns an interface depending on the presence of the [check] type metric
func secureFail(isCheck bool, isLoad bool) interface{} {
	if isCheck {
		return false
	} else if isLoad {
		return int64(-1)
	} else {
		return nil
	}
}

// runCollectdCommand: runs the [collectdctl] cli with the specified parameters
func (c Collectd) runCollectdCommand(metrics []string, valueList *map[string]interface{}) {
	// Get the hostname of the running machine
	logger.LOG(logger.DEBUG, false, "Attempting to resolve the hostname of the machine")
	hostname, err := runner.RunCommand("hostname", true, true)
	if err != nil {
		logger.LOG(logger.DEBUG, false, "Failed to resolve the hostname of the machine with the error [%s]", err.Error())
		return
	}
	for _, metric := range metrics {
		// Remove square-brackets from metric
		metric := regexp.MustCompile("[\\[\\]]").ReplaceAllString(metric, "")

		logger.LOG(logger.DEBUG, false, "Running the collectd cli for the metric [%s%s]", hostname, metric)
		rawOutput, err := runner.RunCommand(COLLECTD_CLI, true, true, "getval", hostname+metric)
		if err != nil {
			logger.LOG(logger.DEBUG, false, "Failed to run the collectdctl cli with the error [%s]", err.Error())
			return
		}

		value, err := parser.ParseSciNumber(regexp.MustCompile("^value=").Split(rawOutput, -1)[1], true)
		if err != nil {
			logger.LOG(logger.ERROR, false, "Failed to parse the value of the collectdctl cli [%v] with the error [%s]", rawOutput, err.Error())
			return
		}

		// Assign the parameter key to the value fetched from collectdctl
		(*valueList)[metric] = value
		// Log
		logger.LOG(logger.TRACE, false, "Result of the collectd command: [%v]", (*valueList)[metric])
	}
}

/*
func (lbalias *LBalias) checkCollectdMetric() bool {
	return false
}

func (lbalias *LBalias) evaluateLoadCollectd() int {
	return -1
}
*/
