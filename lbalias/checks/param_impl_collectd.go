package checks

import (
	"regexp"
	"strings"

	"gitlab.cern.ch/lb-experts/golbclient/lbalias/utils/parser"
	"gitlab.cern.ch/lb-experts/golbclient/lbalias/utils/runner"
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
)

// runCollectd : Runs the [collectdctl] for the found metric's list and populates the expression [valueList] with the values fetched from the CLI.
func runCollectd(commandPath string, metrics []string, valueList *map[string]interface{}) {
	// Run the CLI for each metric found
	for _, metric := range metrics {
		logger.Trace("first metric [%v]", metric)
		// Remove square-brackets from metric
		metric := regexp.MustCompile("[\\[\\]]").ReplaceAllString(metric, "")
		metricName := regexp.MustCompile("[a-zA-Z-_0-9]*([/][a-zA-Z_-]*)?([:][a-zA-Z-_0-9]+)?").FindAllString(metric, 1)[0]
		logger.Trace("metric [%v]", metric)
		logger.Trace("metricName [%v]", metricName)
		// Extract slice from the metric
		var value float64
		var err error
		slice := 0
		keyName := ""

		if strings.Contains(metric, ":") {
			parts := strings.Split(metric, ":")
			metric = parts[0]
			secondPart := parts[1]
			slice = int(parser.ParseInterfaceAsInteger(secondPart)) - 1
			// if it does not parse as an integer
			if slice == -2 {
				slice = 0
				keyName = secondPart
			}
		}

		logger.Debug("Running the [collectd] path [%s] cli for the metric [%s]", commandPath, metricName)
		rawOutput, err := runner.RunCommand(commandPath, true, "getval", metric)
		logger.Trace("Raw output from [collectdctl] [%v]", rawOutput)
		if err != nil {
			logger.Error("Failed to run the [collectd] cli with the error [%s]", err.Error())
			return
		}

		// No errors when running the CLI
		if keyName == "" {
			rawKeyVals := regexp.MustCompile("(?m)^[a-zA-Z-_0-9]+=([0-9]*[.]?[0-9]*[e][+-][0-9]*)").FindAllStringSubmatch(strings.TrimSpace(rawOutput), -1)

			// Abort if nothing was found
			if len(rawKeyVals) == 0 {
				logger.Error("Failed to match the value of the [collectd] [%v]", rawOutput)
				return
			}

			// Prevent panics for out-of-bounds slices
			if slice < 0 || slice >= len(rawKeyVals) {
				// Fail the whole expression if the index is out-of-bounds
				logger.Error("Accessing the element [%d/%d] of collectd, which is out of bounds", slice, len(rawKeyVals))
				return
			}

			// Get the desired slice
			value, err = parser.ParseSciNumber(rawKeyVals[slice][1], true)
		} else {
			rawKeyVals := regexp.MustCompile("(?m)^"+keyName+"=([0-9]*[.]?[0-9]*[e][+-][0-9]*)").FindAllStringSubmatch(strings.TrimSpace(rawOutput), -1)

			// Abort if nothing was found
			if len(rawKeyVals) == 0 {
				logger.Error("Failed to match the value of the [collectd] [%v] with the key [%s]", rawOutput, keyName)
				return
			}

			// We should have just one hit
			value, err = parser.ParseSciNumber(rawKeyVals[0][1], true)
		}
		if err != nil {
			logger.Error("Failed to parse the value of the [collectd] [%v] with the error [%s]", rawOutput, err.Error())
			return
		}

		// Assign the parameter key to the value fetched from the cli
		logger.Trace("%v and %v", metricName, value)
		(*valueList)[metricName] = value
		// Log
		logger.Trace("Result of the collectd command: [%v]", (*valueList)[metricName])
	}
}
