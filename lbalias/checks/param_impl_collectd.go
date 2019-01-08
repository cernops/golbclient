package checks

import (
	"lbalias/utils/logger"
	"lbalias/utils/parser"
	"lbalias/utils/runner"
	"regexp"
)

// runCollectd : Runs the [collectdctl] for the found metric's list and populates the expression [valueList] with the values fetched from the CLI.
func runCollectd(commandPath string, metrics []string, valueList *map[string]interface{}) {
	// Run the CLI for each metric found
	for _, metric := range metrics {
		// Remove square-brackets from metric
		metric := regexp.MustCompile("[\\[\\]]").ReplaceAllString(metric, "")
		logger.Debug("Running the [collectd] cli for the metric [%s]", metric)
		rawOutput, err := runner.RunCommand(commandPath, true, true, "getval", metric)
		if err != nil {
			logger.Error("Failed to run the [collectd] cli with the error [%s]", err.Error())
			return
		}

		// No errors when running the CLI
		value, err := parser.ParseSciNumber(regexp.MustCompile("^value=").Split(rawOutput, -1)[1], true)

		if err != nil {
			logger.Error("Failed to parse the value of the [collectd] [%v] with the error [%s]", rawOutput, err.Error())
			return
		}
		// Assign the parameter key to the value fetched from the cli
		(*valueList)[metric] = value
		// Log
		logger.Trace("Result of the collectd command: [%v]", (*valueList)[metric])
	}
}
