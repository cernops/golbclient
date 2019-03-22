package ci

import (
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/mapping"
	"strings"
	"testing"

	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/utils/runner"
)

// TestCILemonCLI : checks if the alternative [lemon-cli] used in the CI pipeline is OK
func TestCILemonCLI(t *testing.T) {
	logger.SetLevel(logger.ERROR)
	output, err := runner.Run("/usr/sbin/lemon-cli",
		true , defaultTimeout, "--script", "-m", "13163")
	if err != nil {
		logger.Error("An error was detected when running the CI [lemon-cli]")
		t.FailNow()
	} else if len(strings.TrimSpace(output)) == 0 {
		logger.Error("The CI [lemon-cli] failed to return a row value for a pre-defined metric")
		t.FailNow()
	}
	logger.Trace("CI [lemon-cli] output [%s]", output)
}

// TestLemonFunctionality : fundamental functionality test for the [lemon-cli], output value must not be negative
func TestLemonFunctionality(t *testing.T) {
	logger.SetLevel(logger.ERROR)
	cfg := mapping.NewConfiguration("../test/lbclient_lemon_check_single.conf", "myTest")

	err := lbconfig.Evaluate(cfg, defaultTimeout)
	if err != nil {
		logger.Error("Detected an error when attempting to evaluate the alias [%s], Error [%s]", cfg.ConfigFilePath, err.Error())
		t.Fail()
	}
	if cfg.MetricValue < 0 {
		logger.Error("The metric output value returned negative [%d]. Failing the test...", cfg.MetricValue)
		t.Fail()
	}
}

// TestLemonConfigurationFile : integration test for all the functionality supplied by the lemon-cli
func TestLemonConfigurationFile(t *testing.T) {
	logger.SetLevel(logger.ERROR)

	cfg := mapping.NewConfiguration("../test/lbclient_lemon_check.conf", "lemonTest")
	err := lbconfig.Evaluate(cfg, defaultTimeout)
	if err != nil {
		logger.Error("Failed to run the client for the given configuration file [%s]. Error [%s]", cfg.ConfigFilePath, err.Error())
		t.Fail()
	}
	if cfg.MetricValue < 0 {
		logger.Error("The metric output value returned negative [%d]. Failing the test...", cfg.MetricValue)
		t.Fail()
	}
}

// TestLemonFailedConfigurationFile : integration test for all the functionality supplied by the lemon-cli, fail test
func TestLemonFailedConfigurationFile(t *testing.T) {
	logger.SetLevel(logger.ERROR)

	cfg := mapping.NewConfiguration("../test/lbclient_lemon_check_fail.conf", "lemonFailTest")
	err := lbconfig.Evaluate(cfg, defaultTimeout)
	if err == nil {
		logger.Error("Expecting an error for the given configuration file [%s]. Failing test...", cfg.ConfigFilePath)
		t.Fail()
	}
	if cfg.MetricValue >= 0 {
		logger.Error("The metric output value returned positive [%d] when expecting a negative output. Failing the test...", cfg.MetricValue)
		t.Fail()
	}
}
