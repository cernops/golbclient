package checks

import (
	"fmt"
	"lbalias/utils/logger"
	"lbalias/utils/parser"
	"regexp"
	"strings"

	"github.com/Knetic/govaluate"
)

type ParamCheck struct {
	code    int
	command string
}

func (g ParamCheck) Code() int {
	return g.code
}

func (g ParamCheck) Run(args ...interface{}) interface{} {
	var rVal interface{}
	line := args[0].(string)
	isCheck := strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), "check")
	isLoad := strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), "load")

	// Log
	logger.Debug("Adding [%s] metric [%s]", g.command, line)

	// Support unintentional errors => e.g., [loadcheck collectd], panics if the regex cannot be compiled
	found := regexp.MustCompile("(?i)((check|load)( )+(collectd|lemon))").Split(strings.TrimSpace(line), -1)

	// Found the correct syntax
	if len(found) != 2 || (!isCheck && !isLoad) {
		logger.Error("Incorrect syntax specified at [%s]. Please use (`load` or `check` `<cli>`)", line)
		return preventPanic(isCheck, isLoad)
	}
	rawExpression := found[1]
	// Log
	logger.Trace("Found expression [%s]", rawExpression)

	// If no expression was given, fail the whole expression
	if len(strings.TrimSpace(rawExpression)) == 0 {
		rVal = preventPanic(isCheck, isLoad)
		logger.Debug("Detected a (check|load) without metrics. Returning [%v]", rVal)
		return rVal
	}

	// Backwards compatible (remove unecessary underscores from the expression)
	g.compatibilityProcess(&rawExpression)

	// Discover all the metrics found in the expression
	metrics := regexp.MustCompile(`\[([^\[\]]*)\]`).FindAllString(rawExpression, -1)
	parameters := make(map[string]interface{}, len(metrics))

	// Run command with a list of all the metrics found and return a key/value map
	g.runCommand(metrics, &parameters)

	// Parse the expression
	expression, err := govaluate.NewEvaluableExpression(rawExpression)

	if err != nil {
		logger.Error("Error when evaluating expression [%s]", err)
		return preventPanic(isCheck, isLoad)
	}

	// Evaluate the expression
	result, err2 := expression.Evaluate(parameters)
	if err2 != nil {
		rVal := preventPanic(isCheck, isLoad)
		logger.Error("Error when evaluating the parameters of the expression [%s]. Error [%s]. Returning [%v]", rawExpression, err2, rVal)
		return rVal
	}

	// Debug
	logger.Debug("Expression returned result [%v]", result)
	if isCheck {
		/******************************** CHECK ************************************/
		rVal = parser.ParseInterfaceAsBool(result)
		logger.Trace("Detected a [check] type metric, returning as boolean [%t]", rVal)
		return rVal
	} else if isLoad {
		/********************************* LOAD ************************************/
		rVal = parser.ParseInterfaceAsInteger(result)
		logger.Trace("Detected a [load] type metric, returning as integer [%d]", rVal)
		return rVal
	} else {
		logger.Error("Failed to parse the result of the collectd expression [%v]", result)
		return false
	}
}

// secureFail : returns an interface depending on the presence of the [check] type metric
func preventPanic(isCheck bool, isLoad bool) interface{} {
	if isCheck {
		return false
	} else if isLoad {
		return int32(-1)
	} else {
		return nil
	}
}

// getCliPath : returns the path to the desired cli
func getCliPath(cli string) (_ string) {
	switch strings.ToLower(cli) {
	case "lemon":
		return "/usr/sbin/lemon-cli"
	case "collectd":
		return "/usr/bin/collectdctl"
	default:
		logger.Error("The Generic check does not support the cli type [%s]", cli)
		return
	}
}

// runCommand : @TODO support both [collectdctl-OK] & [lemon-cli-@TODO]
func (g ParamCheck) runCommand(metrics []string, valueList *map[string]interface{}) {
	// Run CLI
	lwCMD := strings.ToLower(g.command)
	cmdPath := getCliPath(lwCMD)
	switch lwCMD {
	case "collectd":
		runCollectd(cmdPath, metrics, valueList)
	case "lemon":
		runLemon(cmdPath, metrics, valueList)
	}
}

// compatibilityProcess : Processes the metric line so that all the metrics found (_metric) are ported to the new format ([metric])
func (g ParamCheck) compatibilityProcess(metric *string) {
	logger.LOGC(logger.TRACE, "Processing metric [%s]", *metric)

	// Trim all spaces
	*metric = strings.Replace(*metric, " ", "", -1)

	// Underscore backwards compatibility with lemon
	toProcess := regexp.MustCompile("_[0-9]+([:][0-9]+)?").FindAllStringIndex(*metric, -1)
	for c, arrI := range toProcess {
		*metric = fmt.Sprintf("%s[%s]%s", (*metric)[:arrI[0]+c], (*metric)[arrI[0]+c+1:arrI[1]+c], (*metric)[arrI[1]+c:])
	}

	logger.Trace("New  metric [%s]", *metric)
}
