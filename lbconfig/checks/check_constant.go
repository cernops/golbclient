package checks

import (
	"fmt"
	"strconv"
	"strings"

	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
)

type MetricConstant struct{}

func (mc MetricConstant) Run(args ...interface{}) (int, error) {
	toParseRaw := strings.Split(args[0].(string), " ")
	if len(toParseRaw) < 3 {
		return -1, fmt.Errorf("the constant metric [%v] does not have the correct syntax", args[0])
	}
	toParse := toParseRaw[2]
	logger.Debug("Attempting to parse constant metric [%s]", toParse)
	f, err := strconv.ParseFloat(toParse, 32)
	if err != nil {
		return -1, fmt.Errorf("the supplied constant is not a number")
	}
	logger.Debug("Successfully parsed the constant [%v]...", f)

	return int(f), nil
}
