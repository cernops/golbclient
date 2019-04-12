package ci

import (
	"testing"

	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
)

// TestConstantLoadFunctionality : fundamental functionality test for the [load constant], output value must be = 250
func TestConstantLoadFunctionality(t *testing.T) {
	logger.SetLevel(logger.ERROR)

	runEvaluate(t,
		lbTest{title: "constant load", configuration: "../test/lbclient_constant.conf", expectedMetricValue: 250})
}
