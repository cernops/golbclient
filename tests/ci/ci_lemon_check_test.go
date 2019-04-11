package ci

import (
	"strings"
	"testing"

	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/utils/runner"
)

// TestCILemonCLI : checks if the alternative [lemon-cli] used in the CI pipeline is OK
func TestCILemonCLI(t *testing.T) {
	logger.SetLevel(logger.ERROR)
	output, err := runner.Run("/usr/sbin/lemon-cli",
		true, defaultTimeout, "--script", "-m", "13163")
	if err != nil {
		logger.Error("An error was detected when running the CI [lemon-cli]")
		t.FailNow()
	} else if len(strings.TrimSpace(output)) == 0 {
		logger.Error("The CI [lemon-cli] failed to return a row value for a pre-defined metric")
		t.FailNow()
	}
	logger.Trace("CI [lemon-cli] output [%s]", output)
}

func TestLemon(t *testing.T) {
	myTests := []lbTest{
		lbTest{title: "CollectdFunctionality", configuration: "../test/lbclient_lemon_check_single.conf", expectedMetricValue: 12},
		lbTest{title: "ConfigurationFile", configuration: "../test/lbclient_lemon_check.conf", expectedMetricValue: 8},
		lbTest{title: "LemonFailed", configuration: "../test/lbclient_lemon_check_fail.conf", expectedMetricValue: -12},
	}
	runMultipleTests(t, myTests)
}
