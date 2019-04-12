package ci

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"

	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/utils/runner"
)

// TestCILemonCLI : checks if the alternative [lemon-cli] used in the CI pipeline is OK
func TestCILemonCLI(t *testing.T) {
	logger.SetLevel(logger.ERROR)
	cliPath := "/usr/sbin/lemon-cli"
	metricId := "13163"
	output, err := runner.Run(cliPath,true, defaultTimeout, "--script", "-m", metricId)

	assert.Nil(t, err,
		"An error was detected when running the CLI [lemon-cli] at [%s]. Error [%s]",
			cliPath, err)

	assert.NotEmpty(t, strings.TrimSpace(output),
		"The CLI [lemon-cli] failed to return a row value for a pre-defined metric [%s]. " +
			"Failing the test...", metricId)

	logger.Trace("CLI [lemon-cli] output [%s]", output)
}

func TestLemon(t *testing.T) {
	myTests := []lbTest{
		{title: "CollectdFunctionality",
			configuration: "../test/lbclient_lemon_check_single.conf", expectedMetricValue: 12},
		{title: "ConfigurationFile",
			configuration: "../test/lbclient_lemon_check.conf", expectedMetricValue: 8},
		{title: "LemonFailed",
			configuration: "../test/lbclient_lemon_check_fail.conf", expectedMetricValue: -12},
	}
	runMultipleTests(t, myTests)
}

func TestLemonLoad(t *testing.T) {
	myTests := []lbTest{
		{title: "LemonLoadSingle",
			configuration: "../test/lbclient_lemon_load_single.conf", expectedMetricValue: 1},
		{title: "LemonLoad",
			configuration: "../test/lbclient_lemon_load.conf", expectedMetricValue: 50},
		{title: "LemonFailed",
			configuration: "../test/lbclient_lemon_check_fail.conf", expectedMetricValue: -12}}

	runMultipleTests(t, myTests)
}
