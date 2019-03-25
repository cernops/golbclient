package ci

import (
	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/mapping"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/utils/runner"
	"strings"
	"testing"
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

// TestCollectdFunctionality : fundamental functionality test for the [collectd], output value must not be negative
func TestCollectdFunctionality(t *testing.T) {
	logger.SetLevel(logger.ERROR)
	cfg := mapping.NewConfiguration("../test/lbclient_collectd_check_single.conf", "my-test-alias.cern.ch")
	err := lbconfig.Evaluate(cfg, defaultTimeout)
	if err != nil {
		logger.Error("Detected an error when attempting to evaluate the configuration file [%s], Error [%s]", cfg.ConfigFilePath, err.Error())
		t.Fail()
	}
	if cfg.MetricValue < 0 {
		logger.Error("The metric output value returned negative [%d]. Failing the test...", cfg.MetricValue)
		t.Fail()
	}
}

// TestCollectdConfigurationFile : integration test for all the functionality supplied by the collectdctl
func TestCollectdConfigurationFile(t *testing.T) {
	logger.SetLevel(logger.ERROR)

	cfg := mapping.NewConfiguration("../test/lbclient_collectd_check.conf", "collectd_comprehensive_test")
	err := lbconfig.Evaluate(cfg, defaultTimeout)
	if err != nil {
		logger.Error("Failed to run the client for the given configuration file [%s]. Error [%s]", cfg.ConfigFilePath,
			err.Error())
		t.Fail()
	}
	if cfg.MetricValue < 0 {
		logger.Error("The metric output value returned negative [%d]. Failing the test...", cfg.MetricValue)
		t.Fail()
	}
}

// TestCollectdFailedConfigurationFile : integration test for all the functionality supplied by the collectdctl, fail test
func TestCollectdFailedConfigurationFile(t *testing.T) {
	logger.SetLevel(logger.FATAL)

	cfg := mapping.NewConfiguration("../test/lbclient_collectd_check_fail.conf", "collectd_intended_fail_test")
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

// TestCollectdConfigurationFileWithKeys : integration test for all the functionality supplied by the collectdctl
func TestCollectdConfigurationFileWithKeys(t *testing.T) {
	logger.SetLevel(logger.ERROR)

	cfg := mapping.NewConfiguration("../test/lbclient_collectd_check_with_keys.conf", "collectd_comprehensive_test_with_keys")
	err := lbconfig.Evaluate(cfg, defaultTimeout)
	if err != nil {
		logger.Error("Failed to run the client for the given configuration file [%s]. Error [%s]", cfg.ConfigFilePath,
			err.Error())
		t.Fail()
	}
	if cfg.MetricValue < 0 {
		logger.Error("The metric output value returned negative [%d]. Failing the test...", cfg.MetricValue)
		t.Fail()
	}
}

// TestCollectdFailedConfigurationFileWithWrongKey : integration test for all the functionality supplied by the collectdctl with wrong key, fail test
func TestCollectdFailedConfigurationFileWithWrongKey(t *testing.T) {
	logger.SetLevel(logger.FATAL)

	cfg := mapping.NewConfiguration("../test/lbclient_collectd_check_fail_with_wrong_key.conf", "collectd_intended_fail_test_with_wrong_key")
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

// TestCollectdFailedConfigurationFileWithEmptyKey : integration test for all the functionality supplied by the collectdctl with empty, fail test
func TestCollectdFailedConfigurationFileWithEmptyKey(t *testing.T) {
	logger.SetLevel(logger.FATAL)

	cfg := mapping.NewConfiguration("../test/lbclient_collectd_check_fail_with_empty_key.conf", "collectd_intended_fail_test_with_empty_key")
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
