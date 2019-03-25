package ci

import (
	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/mapping"
	"testing"
)

// TestCollectdFunctionality : fundamental functionality test for the [collectd], output value must be = 9
func TestCollectdLoadFunctionality(t *testing.T) {
	logger.SetLevel(logger.ERROR)
	cfg := mapping.NewConfiguration("../test/lbclient_collectd_load_single.conf", "collectd_load_functionality_test")
	err := lbconfig.Evaluate(cfg, defaultTimeout)
	if err != nil {
		logger.Error("Detected an error when attempting to evaluate the alias [%s], Error [%s]", cfg.ConfigFilePath, err.Error())
		t.Fail()
	}
	if cfg.MetricValue != 98 {
		logger.Error("The expected metric value was [98] but got [%d] instead. Failing the test...", cfg.MetricValue)
		t.Fail()
	}
}

// TestCollectdLoadConfigurationFile : integration test for all the functionality supplied by the collectdctl, resulting metric must be = 7
func TestCollectdLoadConfigurationFile(t *testing.T) {
	logger.SetLevel(logger.ERROR)

	cfg := mapping.NewConfiguration("../test/lbclient_collectd_load.conf", "collectd_load_comprehensive_test")
	err := lbconfig.Evaluate(cfg, defaultTimeout)
	if err != nil {
		logger.Error("Failed to run the client for the given configuration file [%s]. Error [%s]", cfg.ConfigFilePath, err.Error())
		t.Fail()
	}
	if cfg.MetricValue != 72 {
		logger.Error("The expected metric value was [72] but got [%d] instead. Failing the test...", cfg.MetricValue)
		t.Fail()
	}
}

// TestCollectdLoadFailedConfigurationFile : integration test for all the functionality supplied by the lemon-cli, fail test
func TestCollectdLoadFailedConfigurationFile(t *testing.T) {
	logger.SetLevel(logger.FATAL)

	cfg := mapping.NewConfiguration("../test/lbclient_collectd_load_fail.conf", "collectd_load_intended_fail_test")
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
