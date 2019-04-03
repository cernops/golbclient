package checks

import (
	"fmt"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/checks/parameterized"
	"regexp"
	"strings"

	"github.com/Knetic/govaluate"
	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
)

type ParamCheckType = string

type ParamCheck struct {
	Type param.Parameterized
}

func (g ParamCheck) Run(args ...interface{}) (interface{}, error) {
	var rVal interface{}
	line := args[0].(string)
	isCheck := strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), "check")

	// Abort if no Type was given during the instancing of the ParamCheck struct
	if g.Type == nil {
		return -1, fmt.Errorf("expected a Type of check to be given, please see the contract [Parameterized]")
	}

	isLoad := strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), "load")
	// Log
	logger.Debug("Adding [%s] metric [%s]", g.Type.Name(), line)

	// Support unintentional errors => e.g., [loadcheck collectd], panics if the regex cannot be compiled
	found := regexp.MustCompile("(?i)((check|load)( )+(collectd|lemon))").Split(strings.TrimSpace(line), -1)

	// Found the correct syntax
	if len(found) != 2 || (!isCheck && !isLoad) {
		return -1, fmt.Errorf("incorrect syntax specified at [%s]. Please use (`load` or `check` `<cli>`)", line)
	}

	rawExpression := found[1]
	// Log
	logger.Trace("Found expression [%s]", rawExpression)

	// If no expression was given, fail the whole expression
	if len(strings.TrimSpace(rawExpression)) == 0 {
		return -1, fmt.Errorf("detected a (check|load) without metrics. Returning [%v]", rVal)
	}

	// Backwards compatible (remove unnecessary underscores from the expression)
	g.compatibilityProcess(&rawExpression)

	// Discover all the metrics found in the expression
	metrics := regexp.MustCompile(`\[([^\[\]]*)]`).FindAllString(rawExpression, -1)
	logger.Trace("Found metrics [%v], len [%d]", metrics, len(metrics))
	parameters := make(map[string]interface{}, len(metrics))

	// Run command with a list of all the metrics found and return a key/value map
	err := g.Type.Run(metrics, &parameters)
	if err != nil {
		return -1, err
	}

	// Parse the expression
	expression, err := govaluate.NewEvaluableExpression(rawExpression)

	if err != nil {
		return -1, err
	}

	// Evaluate the expression
	result, err2 := expression.Evaluate(parameters)
	if err2 != nil {
		return -1, err2
	}

	// Debug
	logger.Debug("Expression returned result [%+v]", result)
	return result, nil
}

// compatibilityProcess : Processes the metric line so that all the metrics found (_metric) are ported to the new format ([metric])
func (g ParamCheck) compatibilityProcess(metric *string) {
	logger.Trace("Processing metric [%s]", *metric)

	*metric = regexp.MustCompile(`([]0-9][ ]*)[=]([ ]*[0-9\[])`).ReplaceAllString(*metric, "$1==$2")

	// Trim all spaces
	*metric = strings.Replace(*metric, " ", "", -1)

	// Underscore backwards compatibility with lemon (sliced metrics)
	toProcess := regexp.MustCompile("_[0-9]+([:][0-9]+)?").FindAllStringIndex(*metric, -1)
	for c, arrI := range toProcess {
		*metric = fmt.Sprintf("%s[%s]%s", (*metric)[:arrI[0]+c], (*metric)[arrI[0]+c+1:arrI[1]+c], (*metric)[arrI[1]+c:])
	}

	logger.Trace("Processed metric [%s]", *metric)
}
