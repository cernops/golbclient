package checks

import (
	"fmt"
	"lbalias/utils/logger"
	"lbalias/utils/parser"
	"lbalias/utils/runner"
	"regexp"
	"strings"
)

// sliceEntry : Helper struct for the management of sliced metric entries
type sliceEntry struct {
	name  string
	slice int32
	mname string
}

// runLemon : Runs the [lemon-cli] for the found metric's list and populates the expression [valueList] with the values fetched from the CLI.
func runLemon(commandPath string, metrics []string, valueList *map[string]interface{}) {
	// Join metrics in one string
	metric := strings.Join(metrics, " ")
	// Remove square-brackets from metric
	metric = regexp.MustCompile("[\\[\\]]").ReplaceAllString(metric, "")
	// Create the slices map
	slicesMap := map[int]sliceEntry{}
	slices := regexp.MustCompile("[0-9]+[:][0-9]+").FindAllString(metric, -1)
	for i, slice := range slices {
		ps := strings.Split(slice, ":")
		slicesMap[i] = sliceEntry{fmt.Sprintf("[%s]", ps[0]), parser.ParseInterfaceAsInteger(ps[1]), slice}
	}
	// Log
	logger.Debug("Slices map [%v]", slicesMap)
	// Remove the slice from the metric
	metric = regexp.MustCompile("[:][0-9]+").ReplaceAllString(metric, "")
	// Run the CLI with all the metrics found
	logger.Debug("Running the [lemon] cli for the metrics [%s]", metric)
	// Add the [lemon-cli] arguments
	output, err := runner.RunCommand(commandPath, true, true, "--script", "-m", metric)
	if err != nil {
		logger.Error("Failed to run the [lemon] cli with the error [%s]", err.Error())
		// Fail the whole expression
		return
	}

	// Log
	logger.Trace("OUTPUT FROM LEMON [%v]", output)

	// Create a map of the output
	sepOutput := strings.Split(output, "\n")
	outputMap := make(map[string][]string, len(sepOutput))
	for _, line := range sepOutput {
		ps := strings.Split(line, " ")
		outputMap[fmt.Sprintf("[%s]", ps[1])] = ps[3:]
	}

	// Assign sliced metrics
	for _, slicedMetric := range slicesMap {
		si := int(slicedMetric.slice)
		if si < 0 || si >= len(outputMap[slicedMetric.name]) {
			// Fail the whole expression if the index is out-of-bounds
			return
		}
		(*valueList)[slicedMetric.mname] = parser.ParseInterfaceAsFloat(outputMap[slicedMetric.name][slicedMetric.slice])
	}

	// Assign non-slices metrics
	for _, mname := range metrics {
		if !strings.Contains(mname, ":") {
			runes := []rune(mname)
			normName := string(runes[1 : len(runes)-1])
			(*valueList)[normName] = parser.ParseInterfaceAsFloat(outputMap[mname][0])
		}
	}

	// Log
	logger.Trace("Value map [%v]", (*valueList))
}
