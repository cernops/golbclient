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
	var myTests [6]lbTest
	myTests[0] = lbTest{title: "CollectdFunctionality", configuration: "../test/lbclient_collectd_check_single.conf", shouldWork: true, metricValue: 5}
	myTests[1] = lbTest{title: "ConfigurationFile", configuration: "../test/lbclient_collectd_check.conf", shouldWork: true, metricValue: 3}
	myTests[2] = lbTest{title: "ConfigurationFileWithKeys", configuration: "../test/lbclient_collectd_check_with_keys.conf", shouldWork: true, metricValue: 7}
	myTests[3] = lbTest{title: "FailedConfigurationFile", configuration: "../test/lbclient_collectd_check_fail.conf", shouldWork: true, metricValue: -15}
	myTests[4] = lbTest{title: "FailedConfigurationFileWithWrongKey", configuration: "../test/lbclient_collectd_check_fail_with_wrong_key.conf", shouldWork: true, metricValue: -15}
	myTests[5] = lbTest{title: "FailedConfigurationFileWithEmptyKey", configuration: "../test/lbclient_collectd_check_fail_with_empty_key.conf", shouldWork: true, metricValue: -15}

	//	runMultipleTests(t, false, myTests[:])
}
