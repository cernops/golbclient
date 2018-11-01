package ci

import (
	"strings"
	"testing"

	"gitlab.cern.ch/lb-experts/golbclient/lbalias"
	"gitlab.cern.ch/lb-experts/golbclient/lbalias/utils/runner"
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
)

// TestCICollectdCLI : checks if the alternative [collectd] used in the CI pipeline is OK
func TestCICollectdCLI(t *testing.T) {
	logger.SetLevel(logger.ERROR)
	output, err := runner.RunCommand("/usr/bin/collectdctl",
		true, true, "getval", "test")
	if err != nil {
		logger.Error("An error was detected when running the CI [collectdctl]")
		t.FailNow()
	} else if len(strings.TrimSpace(output)) == 0 {
		logger.Error("The CI [collectdctl] failed to return a row value for a pre-defined metric")
		t.FailNow()
	}
	logger.Trace("CI [collectdctl] output [%s]", output)
}

// TestCollectdFunctionality : fundamental functionality test for the [collectd], output value must not be negative
func TestCollectdFunctionality(t *testing.T) {
	logger.SetLevel(logger.ERROR)
	lba := lbalias.NewLbAlias(
		"collectd_functionality_test",
		true,
		"../test/lbclient_collectd_check_single.conf")
	err := lba.Evaluate()
	if err != nil {
		logger.Error("Detected an error when attempting to evaluate the alias [%s], Error [%s]", lba.Name, err.Error())
		t.Fail()
	}
	if lba.Metric < 0 {
		logger.Error("The metric output value returned negative [%d]. Failing the test...", lba.Metric)
		t.Fail()
	}
}

// TestCollectdConfigurationFile : integration test for all the functionality supplied by the collectdctl
func TestCollectdConfigurationFile(t *testing.T) {
	logger.SetLevel(logger.ERROR)

	lba := lbalias.NewLbAlias(
		"collectd_comprehensive_test",
		false,
		"../test/lbclient_collectd_check.conf")
	err := lba.Evaluate()
	if err != nil {
		logger.Error("Failed to run the client for the given configuration file [%s]. Error [%s]", lba.ConfigFile, err.Error())
		t.Fail()
	}
	if lba.Metric < 0 {
		logger.Error("The metric output value returned negative [%d]. Failing the test...", lba.Metric)
		t.Fail()
	}
}

// TestCollectdFailedConfigurationFile : integration test for all the functionality supplied by the collectdctl, fail test
func TestCollectdFailedConfigurationFile(t *testing.T) {
	logger.SetLevel(logger.FATAL)

	lba := lbalias.NewLbAlias(
		"collectd_intended_fail_test",
		false,
		"../test/lbclient_collectd_check_fail.conf")
	err := lba.Evaluate()
	if err != nil {
		logger.Error("Failed to run the client for the given configuration file [%s]. Error [%s]", lba.ConfigFile, err.Error())
		t.Fail()
	}
	if lba.Metric >= 0 {
		logger.Error("The metric output value returned positive [%d] when expecting a negative output. Failing the test...", lba.Metric)
		t.Fail()
	}
}
