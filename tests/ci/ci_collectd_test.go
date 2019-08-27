package ci

import (
	"gitlab.cern.ch/lb-experts/golbclient/tests"
	"strings"
	"testing"

	logger "github.com/sirupsen/logrus"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/utils/runner"
)

var collectdTests, collectdLoadTests []tests.LbTest

func init() {
	collectdTests = []tests.LbTest{
		{Title: "CollectdFunctionality",
			Configuration: "../test/lbclient_collectd_check_single.conf", ExpectedMetricValue: 5},
		{Title: "ConfigurationFile",
			Configuration: "../test/lbclient_collectd_check.conf", ExpectedMetricValue: 3},
		{Title: "ConfigurationFileWithKeys",
			Configuration: "../test/lbclient_collectd_check_with_keys.conf", ExpectedMetricValue: 7},
		{Title: "FailedConfigurationFile",
			Configuration: "../test/lbclient_collectd_check_fail.conf", ShouldFail: true, ExpectedMetricValue: -15},
		{Title: "FailedConfigurationFileWithWrongKey",
			Configuration: "../test/lbclient_collectd_check_fail_with_wrong_key.conf",
			ShouldFail: true, ExpectedMetricValue: -15},
		{Title: "FailedConfigurationFileWithEmptyKey",
			Configuration: "../test/lbclient_collectd_check_fail_with_empty_key.conf",
			ShouldFail: true, ExpectedMetricValue: -15},
	}

	collectdLoadTests = []tests.LbTest{
		{Title: "CollectdLoad",
			Configuration: "../test/lbclient_collectd_load_single.conf", ExpectedMetricValue: 98},
		{Title: "LoadConfigurationFile",
			Configuration: "../test/lbclient_collectd_load.conf", ExpectedMetricValue: 72},
		{Title: "LoadFailedConfigurationFile",
			Configuration: "../test/lbclient_collectd_load_fail.conf", ShouldFail: true, ExpectedMetricValue: -15},
	}
}

// TestCICollectdCLI : checks if the alternative [collectd] used in the CI pipeline is OK
func TestCICollectdCLI(t *testing.T) {
	logger.SetLevel(logger.ErrorLevel)
	output, err := runner.Run("/usr/bin/collectdctl",
		true, tests.DefaultTimeout, "getval", "test")
	if err != nil {
		logger.Errorf("An error was detected when running the CI [collectdctl]. Error [%s]", err.Error())
		t.FailNow()
	} else if len(strings.TrimSpace(output)) == 0 {
		logger.Errorf("The CI [collectdctl] failed to return a row value for a pre-defined metric")
		t.FailNow()
	}
	logger.Tracef("CI [collectdctl] output [%s]", output)
}

func TestCollectd(t *testing.T) {
	tests.RunMultipleTests(t, collectdTests)
}

func BenchmarkCollectd(b *testing.B) {
	tests.RunMultipleTests(b, collectdTests)
}

func TestCollectdLoad(t *testing.T) {
	tests.RunMultipleTests(t, collectdLoadTests)
}

func BenchmarkCollectdLoad(b *testing.B) {
	tests.RunMultipleTests(b, collectdLoadTests)
}
