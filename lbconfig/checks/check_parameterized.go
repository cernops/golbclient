package checks

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Knetic/govaluate"
	logger "github.com/sirupsen/logrus"
	param "gitlab.cern.ch/lb-experts/golbclient/lbconfig/checks/parameterized"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/utils/parser"
)

type ParamCheckType = string
const (
	ALARM ParamCheckType = "alarm"
)


type ParamCheck struct {
	Impl param.Parameterized
	Type ParamCheckType
}

func (g ParamCheck) isAlarm() bool {
	return strings.TrimSpace(strings.ToLower(g.Type)) == ALARM
}

func (g ParamCheck) Run(contextLogger *logger.Entry, args ...interface{}) (int, error) {
	var rVal interface{}
	line := args[0].(string)
	isCheck := strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), "check")

	// Abort if no Impl was given during the instancing of the ParamCheck struct
	if g.Impl == nil {
		return -1, fmt.Errorf("expected a Impl of check to be given, please see the contract [Parameterized]")
	}

	isLoad := strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), "load")
	// Log
	contextLogger.Debugf("Adding [%s] metric [%s]", g.Impl.Name(), line)

	// Support unintentional errors => e.g., [loadcheck collectd], panics if the regex cannot be compiled
	found := regexp.MustCompile("(?i)(((check)( )+(collectd_alarms))|(check|load)( )+(collectd|lemon))").Split(strings.TrimSpace(line), -1)

	// Found the correct syntax
	if len(found) != 2 || (!isCheck && !isLoad) {
		return -1, fmt.Errorf("incorrect syntax specified at [%s]. Please use (`load` or `check` `<cli>`)", line)
	}



	rawExpression := found[1]
	// Log
	contextLogger.Tracef("Found expression [%s]", rawExpression)

	// If no expression was given, fail the whole expression
	if len(strings.TrimSpace(rawExpression)) == 0 {
		return -1, fmt.Errorf("detected a (check|load) without metrics. Returning [%v]", rVal)
	}

	var metrics []string
	if !g.isAlarm() {
		// Backwards compatible (remove unnecessary underscores from the expression)
		g.compatibilityProcess(contextLogger, &rawExpression)

		// Discover all the metrics found in the expression
		metrics = regexp.MustCompile(`\[([^\[\]]*)]`).FindAllString(rawExpression, -1)
	} else {
		// Extract everything between curly brackets (JSON) and send it down to the impl
		metrics = []string{regexp.MustCompile(`\[.*]`).FindString(rawExpression)}
	}

	contextLogger.Tracef("Found metrics [%v], len [%d]", metrics, len(metrics))
	parameters := make(map[string]interface{}, len(metrics))

	// Run command with a list of all the metrics found and return a key/value map
	err := g.Impl.Run(contextLogger.WithField("TYPE", strings.ToUpper(g.Impl.Name())), metrics, &parameters)
	if err != nil {
		return -1, err
	}

	// Return before evaluating an expression
	if g.isAlarm() {
		return 1, nil
	}

	// Parse the expression
	expression, err := govaluate.NewEvaluableExpression(rawExpression)

	if err != nil {
		return -1, err
	}

	// Evaluate the expression
	result, err2 := expression.Evaluate(parameters)
	intResult := int(parser.ParseInterfaceAsInteger(result))
	if err2 != nil {
		return -1, err2
	}

	// Debug
	contextLogger.Debugf("Expression returned result [%+v]", result)
	return intResult, nil
}

// compatibilityProcess : Processes the metric line so that all the metrics found (_metric) are ported to the new format ([metric])
func (g ParamCheck) compatibilityProcess(contextLogger *logger.Entry, metric *string) {
	contextLogger.Tracef("Processing metric [%s]", *metric)

	*metric = regexp.MustCompile(`([]0-9][ ]*)[=]([ ]*[0-9\[])`).ReplaceAllString(*metric, "$1==$2")

	// Trim all spaces
	*metric = strings.Replace(*metric, " ", "", -1)

	// Underscore backwards compatibility with lemon (sliced metrics)
	toProcess := regexp.MustCompile("_[0-9]+([:][0-9]+)?").FindAllStringIndex(*metric, -1)
	for c, arrI := range toProcess {
		*metric = fmt.Sprintf("%s[%s]%s", (*metric)[:arrI[0]+c], (*metric)[arrI[0]+c+1:arrI[1]+c], (*metric)[arrI[1]+c:])
	}

	logger.Tracef("Processed metric [%s]", *metric)
}
