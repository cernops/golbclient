package checks

import (
	"gitlab.cern.ch/lb-experts/golbclient/lbalias/utils/parser"
	"gitlab.cern.ch/lb-experts/golbclient/lbalias/utils/runner"
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
	"regexp"
	"strings"
)

// runCollectd : Runs the [collectdctl] for the found metric's list and populates the expression [valueList] with the values fetched from the CLI.
func runCollectd(commandPath string, metrics []string, valueList *map[string]interface{}) {
	// Run the CLI for each metric found
	for _, metric := range metrics {
		// Remove square-brackets from metric
		metric := regexp.MustCompile("[\\[\\]]").ReplaceAllString(metric, "")
		metricName := regexp.MustCompile("[a-zA-Z]*([/][a-zA-Z_-]*)?([:][0-9])?").FindAllString(metric, 1)[0]

		// Extract slice from the metric
		slice := 0
		if strings.Contains(metric, ":") {
			slice = int(parser.ParseInterfaceAsInteger(strings.Split(metric, ":")[1]))
			metric = regexp.MustCompile("[:][0-9]+").ReplaceAllString(metric, "")
		}

		logger.Debug("Running the [collectd] path [%s] cli for the metric [%s]", commandPath, metricName)
		rawOutput, err := runner.RunCommand(commandPath, true, true, "getval", metric)
		logger.Trace("Raw output from [collectdctl] [%v]", rawOutput)
		if err != nil {
			logger.Error("Failed to run the [collectd] cli with the error [%s]", err.Error())
			return
		}

		// No errors when running the CLI
		rawValue := regexp.MustCompile("[0-9]*[.]?[0-9]*[e][+-][0-9]*").FindAllString(strings.TrimSpace(rawOutput), -1)
		logger.Trace("Metrics received from [collectdctl] [%v]", rawValue)

		// Abort if nothing was found
		if len(rawValue) == 0 {
			return
		}

		// Prevent panics for out-of-bounds slices
		slice = slice % len(rawValue)

		// Get the desired slice
		value, err := parser.ParseSciNumber(rawValue[slice], true)

		if err != nil {
			logger.Error("Failed to parse the value of the [collectd] [%v] with the error [%s]", rawOutput, err.Error())
			return
		}
		// Assign the parameter key to the value fetched from the cli
		(*valueList)[metricName] = value
		// Log
		logger.Trace("Result of the collectd command: [%v]", (*valueList)[metricName])
	}
}
