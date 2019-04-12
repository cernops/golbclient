package ci

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"

	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/utils/runner"
)

// TestCICollectdCLI : checks if the alternative [collectd] used in the CI pipeline is OK
func TestCICollectdCLI(t *testing.T) {
	logger.SetLevel(logger.ERROR)
	cliPath := "/usr/bin/collectdctl"
	metricId :=  "test"
	output, err := runner.Run(cliPath,true, defaultTimeout, "getval", metricId)

	assert.Nil(t, err, "An error was detected when running the CLI [collectdctl] at [%s]. Error [%s]",
		cliPath, err)
	assert.NotEmpty(t, strings.TrimSpace(output),
		"The CLI [collectdctl] failed to return a row value for a pre-defined metric [%s]. " +
			"Failing the test...", metricId)

	logger.Trace("CLI [collectdctl] output [%s]", output)
}

func TestCollectd(t *testing.T) {
	myTests := []lbTest{
		{title: "CollectdFunctionality",
			configuration: "../test/lbclient_collectd_check_single.conf", expectedMetricValue: 5},
		{title: "ConfigurationFile",
			configuration: "../test/lbclient_collectd_check.conf", expectedMetricValue: 3},
		{title: "ConfigurationFileWithKeys",
			configuration: "../test/lbclient_collectd_check_with_keys.conf", expectedMetricValue: 7},
		{title: "FailedConfigurationFile",
			configuration: "../test/lbclient_collectd_check_fail.conf", shouldFail: true, expectedMetricValue: -15},
		{title: "FailedConfigurationFileWithWrongKey",
			configuration: "../test/lbclient_collectd_check_fail_with_wrong_key.conf", shouldFail: true, expectedMetricValue: -15},
		{title: "FailedConfigurationFileWithEmptyKey",
			configuration: "../test/lbclient_collectd_check_fail_with_empty_key.conf", shouldFail: true, expectedMetricValue: -15},
	}

	runMultipleTests(t, myTests)
}

func TestCollectdLoad(t *testing.T) {
	myTests := []lbTest{
		{title: "CollectdLoad",
			configuration: "../test/lbclient_collectd_load_single.conf", expectedMetricValue: 98},
		{title: "LoadConfigurationFile",
			configuration: "../test/lbclient_collectd_load.conf", expectedMetricValue: 72},
		{title: "LoadFailedConfigurationFile",
			configuration: "../test/lbclient_collectd_load_fail.conf", shouldFail: true, expectedMetricValue: -15},
	}

	runMultipleTests(t, myTests)
}
