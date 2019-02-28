package ci

import (
	"testing"

	"gitlab.cern.ch/lb-experts/golbclient/lbalias"
	"gitlab.cern.ch/lb-experts/golbclient/utils"
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
)

// TestCommandFunctionality : fundamental functionality test for the [command]
func TestCommandFunctionality(t *testing.T) {
	logger.SetLevel(logger.ERROR)
	cfg := utils.NewConfiguration("../test/lbclient_command.conf", "command_load_functionality_test")
	err := lbalias.Evaluate(cfg)
	if err != nil {
		logger.Error("Detected an error when attempting to evaluate the configuration file [%s], Error [%s]", cfg.ConfigFilePath,
			err.Error())
		t.Fail()
	} else if cfg.MetricValue < 0 {
		logger.Error("Received a negative metric value [%d] when a positive number was expected. Failing test...",
			cfg.MetricValue)
		t.Fail()
	}
}

// TestCommandFailFunctionality : fundamental functionality test for the [command]
func TestCommandFailFunctionality(t *testing.T) {
	logger.SetLevel(logger.FATAL)
	cfg := utils.NewConfiguration("../test/lbclient_failed_command.conf", "command_load_functionality_test")
	err := lbalias.Evaluate(cfg)
	if err == nil {
		logger.Error("An error was expected when attempting to evaluate the configuration file [%s]. Failing the test...", cfg.ConfigFilePath)
		t.Fail()
	} else if cfg.MetricValue > 0 {
		logger.Error("Received a positive metric value [%d] when a negative number was expected. Failing test...",
			cfg.MetricValue)
		t.Fail()
	}
}
