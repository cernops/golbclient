package param

import (
	"fmt"
	logger "github.com/sirupsen/logrus"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/utils/parser"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/utils/runner"
	"regexp"
	"strings"
)

type LemonImpl struct {
	CommandPath string
}

// sliceEntry : Helper struct for the management of sliced metric entries
type sliceEntry struct {
	name  string
	slice int32
	mname string
}

func (li LemonImpl) Name() string {
	return "lemon"
}

// Run : Runs the [lemon-cli] for the found metric's list and populates the expression [valueList] with the values fetched from the CLI.
func (li LemonImpl) Run(contextLogger *logger.Entry, metrics []string, valueList *map[string]interface{}) error {
	if len(li.CommandPath) == 0 {
		li.CommandPath = "/usr/sbin/lemon-cli"
	}

	// Join metrics in one string
	metric := strings.Join(metrics, " ")
	// Remove square-brackets from metric
	metric = regexp.MustCompile("[\\[\\]]").ReplaceAllString(metric, "")
	// Create the slices map
	contextLogger.Tracef("Metric [%s]", metric)
	slicesMap := map[int]sliceEntry{}
	slices := regexp.MustCompile("[0-9]{2,}[:][0-9]").FindAllString(metric, -1)
	contextLogger.Tracef("Found slices [%v]", slices)

	for i, slice := range slices {
		slice = regexp.MustCompile("[ ]").ReplaceAllString(slice, "")
		contextLogger.Tracef("Processing slice [%s]", slice)
		ps := strings.Split(slice, ":")
		slicesMap[i] = sliceEntry{fmt.Sprintf("[%s]", ps[0]), parser.ParseInterfaceAsInteger(ps[1]), slice}
	}
	// Log
	contextLogger.Debugf("Slices map [%v]", slicesMap)
	// Remove the slice from the metric
	metric = regexp.MustCompile("[:][0-9]+").ReplaceAllString(metric, "")
	// Run the CLI with all the metrics found
	contextLogger.Debugf("Running the [lemon] cli path [%s] for the metrics [%s]", li.CommandPath, metric)
	// Add the [lemon-cli] arguments

	output, err := runner.Run(li.CommandPath, true, 0, "--script", "-m", metric)
	if err != nil {
		return fmt.Errorf("failed to run the [lemon] cli with the error [%s]", err.Error())
	}

	// Abort if nothing is returned
	if len(output) == 0 {
		return fmt.Errorf("the lemon command returned an empty output")
	}

	// Create a map of the output
	sepOutput := strings.Split(output, "\n")
	outputMap := make(map[string][]string, len(sepOutput))
	for _, line := range sepOutput {
		ps := strings.Split(line, " ")
		outputMap[fmt.Sprintf("[%s]", ps[1])] = ps[3:]
	}

	contextLogger.Tracef("Output map [%v]", outputMap)

	// Assign sliced metrics
	for _, slicedMetric := range slicesMap {
		si := int(slicedMetric.slice) - 1
		if si < 0 || si >= len(outputMap[slicedMetric.name]) {
			// Fail the whole expression if the index is out-of-bounds
			return fmt.Errorf("the lemon slice is out of bounds")
		}
		contextLogger.Tracef("Assigning sliced metric [%s] to value [%s]", slicedMetric.mname, outputMap[slicedMetric.name][si])
		(*valueList)[slicedMetric.mname] = parser.ParseInterfaceAsFloat(outputMap[slicedMetric.name][si])
		contextLogger.Tracef("Value list [%v]", *valueList)
	}

	// Assign non-slices metrics
	for _, mname := range metrics {
		if !strings.Contains(mname, ":") && len(outputMap[mname]) != 0 {
			contextLogger.Tracef("Assigning non-slices metric [%s]", mname)
			runes := []rune(mname)
			normName := string(runes[1 : len(runes)-1])
			(*valueList)[normName] = parser.ParseInterfaceAsFloat(outputMap[mname][0])
			contextLogger.Tracef("Value list [%v]", *valueList)
		}
	}

	// Log
	contextLogger.Tracef("Value map [%v]", *valueList)
	return nil
}
