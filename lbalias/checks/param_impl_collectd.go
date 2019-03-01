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
			secondPart := strings.Split(metric, ":")[1]
			logger.Trace("secondPart [%v]", secondPart)
			slice = int(parser.ParseInterfaceAsInteger(secondPart)) - 1
			logger.Trace("slice [%v]", slice)
			if slice == -2 {
				slice = 0
				keyName = regexp.MustCompile("^[a-zA-Z-_0-9]+").FindAllString(secondPart, 1)[0]
				logger.Trace("after first MustCompile")
			}
			metric = regexp.MustCompile("[:].+").ReplaceAllString(metric, "")
			logger.Trace("after second MustCompile")
		}

		logger.Debug("Running the [collectd] path [%s] cli for the metric [%s]", commandPath, metricName)
		rawOutput, err := runner.RunCommand(commandPath, true, "getval", metric)
		logger.Trace("Raw output from [collectdctl] [%v]", rawOutput)
		if err != nil {
			logger.Error("Failed to run the [collectd] cli with the error [%s]", err.Error())
			return
		}

		// No errors when running the CLI
		rawKeys := regexp.MustCompile("(?m)^([a-zA-Z-_0-9]+)(?:[=])").FindAllString(strings.TrimSpace(rawOutput), -1)
		// Remove the '='
		rawValues := regexp.MustCompile("(?m)[0-9]*[.]?[0-9]*[e][+-][0-9]*$").FindAllString(strings.TrimSpace(rawOutput), -1)
		logger.Trace("Metrics received from [collectdctl] rawKey[%v]", rawKeys)
		logger.Trace("Metrics received from [collectdctl] rawValue[%v]", rawValues)

		// Abort if nothing was found
		if len(rawValues) == 0 {
			return
		}

		// Prevent panics for out-of-bounds slices

		if slice < 0 || slice >= len(rawValues) {
			// Fail the whole expression if the index is out-of-bounds
			logger.Error("Accessing the element [%d/%d] of collectd, which is out of bounds", slice, len(rawValues))
			return
		}

		if keyName == "" {
			// Get the desired slice
			value, err = parser.ParseSciNumber(rawValues[slice], true)
			if err != nil {
				logger.Error("Failed to parse the value of the [collectd] [%v] with the error [%s]", rawOutput, err.Error())
				return
			}
		} else {
			for index, rawK := range rawKeys {
				rk := strings.Split(rawK, "=")[0]
				if keyName == rk {
					value, err = parser.ParseSciNumber(rawValues[index], true)
					break
				}
				if err != nil {
					logger.Error("Failed to parse the value of the [collectd] [%v] with the error [%s]", rawOutput, err.Error())
					return
				}
			}
			logger.Error("Failed to match the value of the [collectd] [%v] with the key [%s]", rawOutput, keyName)

		}

		// Assign the parameter key to the value fetched from the cli
		logger.Trace("IS THIS THE BROKEN COMMAND??")
		logger.Trace("%v and %v", metricName, value)
		(*valueList)[metricName] = value
		// Log
		logger.Trace("Result of the collectd command: [%v]", (*valueList)[metricName])
	}
}
