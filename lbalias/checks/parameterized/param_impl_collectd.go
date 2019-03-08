package param

import (
	"fmt"
	"regexp"
	"strings"

	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
	"gitlab.cern.ch/lb-experts/golbclient/lbalias/utils/parser"
	"gitlab.cern.ch/lb-experts/golbclient/lbalias/utils/runner"
)

type CollectdImpl struct {
	CommandPath string
}

func (ci CollectdImpl) Name() string {
	return "collectd"
}

// Run : Runs the [collectdctl] for the found metric's list and populates the expression [valueList] with the values fetched from the CLI.
func (ci CollectdImpl) Run(metrics []string, valueList *map[string]interface{}) error {
	if len(ci.CommandPath) == 0 {
		ci.CommandPath = "/usr/bin/collectdctl"
	}

	// Run the CLI for each metric found
	for _, metric := range metrics {
		logger.Trace("Looking for the collectd metric [%v]", metric)
		// Remove square-brackets from metric
		metric := regexp.MustCompile("[\\[\\]]").ReplaceAllString(metric, "")
		metricName := regexp.MustCompile("[a-zA-Z-_0-9]*([/][a-zA-Z_-]*)?([:][a-zA-Z-_0-9]+)?").FindAllString(metric, 1)[0]
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
				logger.Trace("The anchor is not an integer [%v]", secondPart)
				if secondPart == "" {
					return fmt.Errorf("empty anchor in the metric [%v:]", metric)
				} else {
					slice = 0
					keyName = secondPart
				}
			}
		}

		logger.Debug("Running the [collectd] path [%s] cli for the metric [%s]", ci.CommandPath, metricName)
		rawOutput, err := runner.RunCommand(ci.CommandPath, true, "getval", metric)
		logger.Trace("Raw output from [collectdctl] [%v]", rawOutput)
		if err != nil {
			return fmt.Errorf("failed to run the [collectd] cli with the error [%s]", err.Error())
		}

		// No errors when running the CLI
		if keyName == "" {
			rawKeyVals := regexp.MustCompile("(?m)^[a-zA-Z-_0-9]+=([0-9]*[.]?[0-9]*[e][+-][0-9]*)").FindAllStringSubmatch(strings.TrimSpace(rawOutput), -1)

			// Abort if nothing was found
			if len(rawKeyVals) == 0 {
				return fmt.Errorf("failed to match the value of the [collectd] [%v]", rawOutput)
			}

			// Prevent panics for out-of-bounds slices
			if slice < 0 || slice >= len(rawKeyVals) {
				// Fail the whole expression if the index is out-of-bounds
				return fmt.Errorf("accessing the element [%d/%d] of collectd, which is out of bounds", slice, len(rawKeyVals))
			}

			// Get the desired slice
			value, err = parser.ParseSciNumber(rawKeyVals[slice][1], true)
		} else {
			rawKeyVals := regexp.MustCompile(fmt.Sprintf("(?m)^%s=([0-9]*[.]?[0-9]*[e][+-][0-9]*)", keyName)).FindAllStringSubmatch(strings.TrimSpace(rawOutput), -1)

			// Abort if nothing was found
			if len(rawKeyVals) == 0 {
				return fmt.Errorf("failed to match the value of the [collectd] [%v] with the key [%s]", rawOutput, keyName)
			}

			// We should have just one hit
			value, err = parser.ParseSciNumber(rawKeyVals[0][1], true)
		}
		if err != nil {
			return fmt.Errorf("failed to parse the value of the [collectd] [%v] with the error [%s]", rawOutput, err.Error())
		}

		// Assign the parameter key to the value fetched from the cli
		logger.Trace("%v and %v", metricName, value)
		(*valueList)[metricName] = value
		// Log
		logger.Trace("Result of the collectd command: [%v]", (*valueList)[metricName])
	}
	return nil
}
