package ci

import (
	"testing"

	logger "github.com/sirupsen/logrus"
)

// TestConstantLoadFunctionality : fundamental functionality test for the [load constant], output value must be = 250
func TestConstantLoadFunctionality(t *testing.T) {
	logger.SetLevel(logger.ErrorLevel)

	runEvaluate(t,
		lbTest{title: "constant load", configuration: "../test/lbclient_constant.conf", expectedMetricValue: 250})
}
