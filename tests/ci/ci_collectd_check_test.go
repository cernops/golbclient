package ci

import (
	"strings"
	"testing"

	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/utils/runner"
)

// TestCICollectdCLI : checks if the alternative [collectd] used in the CI pipeline is OK
func TestCICollectdCLI(t *testing.T) {
	logger.SetLevel(logger.ERROR)
	output, err := runner.Run("/usr/bin/collectdctl",
		true, defaultTimeout, "getval", "test")
	if err != nil {
		logger.Error("An error was detected when running the CI [collectdctl]")
		t.FailNow()
	} else if len(strings.TrimSpace(output)) == 0 {
		logger.Error("The CI [collectdctl] failed to return a row value for a pre-defined metric")
		t.FailNow()
	}
	logger.Trace("CI [collectdctl] output [%s]", output)
}

func TestCollectd(t *testing.T) {
	myTests := []lbTest{
		lbTest{title: "CollectdFunctionality", configuration: "../test/lbclient_collectd_check_single.conf", expectedMetricValue: 5},
		lbTest{title: "ConfigurationFile", configuration: "../test/lbclient_collectd_check.conf", expectedMetricValue: 3},
		lbTest{title: "ConfigurationFileWithKeys", configuration: "../test/lbclient_collectd_check_with_keys.conf", expectedMetricValue: 7},
		lbTest{title: "FailedConfigurationFile", configuration: "../test/lbclient_collectd_check_fail.conf", shouldFail: true, expectedMetricValue: -15},
		lbTest{title: "FailedConfigurationFileWithWrongKey", configuration: "../test/lbclient_collectd_check_fail_with_wrong_key.conf", shouldFail: true, expectedMetricValue: -15},
		lbTest{title: "FailedConfigurationFileWithEmptyKey", configuration: "../test/lbclient_collectd_check_fail_with_empty_key.conf", shouldFail: true, expectedMetricValue: -15},
	}

	runMultipleTests(t, myTests)
}
