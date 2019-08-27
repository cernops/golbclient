package ci

import (
	"gitlab.cern.ch/lb-experts/golbclient/tests"
	"testing"

	logger "github.com/sirupsen/logrus"
)

var constantTest tests.LbTest

func init() {
	constantTest = tests.LbTest{
		Title: "constant load", Configuration: "../test/lbclient_constant.conf", ExpectedMetricValue: 250}
}

// TestConstantLoadFunctionality : fundamental functionality test for the [load constant], output value must be = 250
func TestConstantLoadFunctionality(t *testing.T) {
	logger.SetLevel(logger.ErrorLevel)
	tests.RunEvaluate(t, constantTest)
}

func BenchmarkConstantLoadFunctionality(b *testing.B) {
	logger.SetLevel(logger.ErrorLevel)
	tests.RunEvaluate(b, constantTest)
}
