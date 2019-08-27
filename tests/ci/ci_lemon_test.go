package ci

import (
	"gitlab.cern.ch/lb-experts/golbclient/tests"
	"strings"
	"testing"

	logger "github.com/sirupsen/logrus"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/utils/runner"
)

var lemonTests, lemonLoadTests []tests.LbTest

func init() {
	lemonTests = []tests.LbTest{
		{Title: "CollectdFunctionality",
			Configuration: "../test/lbclient_lemon_check_single.conf", ExpectedMetricValue: 12},
		{Title: "ConfigurationFile",
			Configuration: "../test/lbclient_lemon_check.conf", ExpectedMetricValue: 8},
		{Title: "LemonFailed",
			Configuration: "../test/lbclient_lemon_check_fail.conf", ExpectedMetricValue: -12},
	}

	lemonLoadTests  = []tests.LbTest{
		{Title: "LemonLoadSingle",
			Configuration: "../test/lbclient_lemon_load_single.conf", ExpectedMetricValue: 1},
		{Title: "LemonLoad",
			Configuration: "../test/lbclient_lemon_load.conf", ExpectedMetricValue: 50},
		{Title: "LemonFailed",
			Configuration: "../test/lbclient_lemon_check_fail.conf", ExpectedMetricValue: -12},
	}
}

// TestCILemonCLI : checks if the alternative [lemon-cli] used in the CI pipeline is OK
func TestCILemonCLI(t *testing.T) {
	logger.SetLevel(logger.ErrorLevel)
	output, err := runner.Run("/usr/sbin/lemon-cli",
		true, tests.DefaultTimeout, "--script", "-m", "13163")
	if err != nil {
		t.Fatal("An error was detected when running the CI [lemon-cli]")
	} else if len(strings.TrimSpace(output)) == 0 {
		t.Fatal("The CI [lemon-cli] failed to return a row value for a pre-defined metric")
	}
	logger.Tracef("CI [lemon-cli] output [%s]", output)
}

func TestLemon(t *testing.T) {
	tests.RunMultipleTests(t, lemonTests)
}

func BenchmarkLemon(b *testing.B) {
	tests.RunMultipleTests(b, lemonLoadTests)
}

func TestLemonLoad(t *testing.T) {
	tests.RunMultipleTests(t, lemonLoadTests)
}

func BenchmarkLemonLoad(b *testing.B) {
	tests.RunMultipleTests(b, lemonLoadTests)
}