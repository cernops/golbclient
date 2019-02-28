package ci

import (
	"testing"

	"gitlab.cern.ch/lb-experts/golbclient/lbalias"
	"gitlab.cern.ch/lb-experts/golbclient/utils"
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
)

// TestLemonLoadFunctionality : fundamental functionality test for the [lemon-cli], output value must be = 1
func TestLemonLoadFunctionality(t *testing.T) {
	logger.SetLevel(logger.ERROR)
	cfg := utils.NewConfiguration("../test/lbclient_lemon_load_single.conf", "myTest")
	err := lbalias.Evaluate(cfg)
	if err != nil {
		logger.Error("Detected an error when attempting to evaluate the alias [%s], Error [%s]", cfg.ConfigFilePath, err.Error())
		t.Fail()
	}
	if cfg.MetricValue != 1 {
		logger.Error("The expected metric value was [1] but got [%d] instead. Failing the test...", cfg.MetricValue)
		t.Fail()
	}
}

// TestLemonLoadConfigurationFile : integration test for all the functionality supplied by the lemon-cli, output value must be = 35
func TestLemonLoadConfigurationFile(t *testing.T) {
	logger.SetLevel(logger.ERROR)

	cfg := utils.NewConfiguration("../test/lbclient_lemon_load.conf", "lemonTest")
	err := lbalias.Evaluate(cfg)
	if err != nil {
		logger.Error("Failed to run the client for the given configuration file [%s]. Error [%s]", cfg.ConfigFilePath, err.Error())
		t.Fail()
	}
	if cfg.MetricValue != 27 {
		logger.Error("The expected metric value was [27] but got [%d] instead. Failing the test...", cfg.MetricValue)
		t.Fail()
	}
}

// TestLemonFailedConfigurationFile : integration test for all the functionality supplied by the lemon-cli, fail test
func TestLemonLoadFailedConfigurationFile(t *testing.T) {
	logger.SetLevel(logger.FATAL)

	cfg := utils.NewConfiguration("../test/lbclient_lemon_load_fail.conf", "lemonFailTest")
	err := lbalias.Evaluate(cfg)
	if err == nil {
		logger.Error("Expected an error for the given configuration file [%s]. Failing test...", cfg.ConfigFilePath)
		t.Fail()
	}
	if cfg.MetricValue >= 0 {
		logger.Error("The metric output value returned positive [%d] when expecting a negative output. Failing the test...", cfg.MetricValue)
		t.Fail()
	}
}
