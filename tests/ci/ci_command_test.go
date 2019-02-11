package ci

import (
	"gitlab.cern.ch/lb-experts/golbclient/lbalias"
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
	"testing"
)

// TestCommandFunctionality : fundamental functionality test for the [command]
func TestCommandFunctionality(t *testing.T) {
	logger.SetLevel(logger.ERROR)
	lba := lbalias.LBalias{Name: "command_load_functionality_test",
		Syslog:     true,
		ChecksDone: make(map[string]bool),
		ConfigFile: "../test/lbclient_command.conf"}
	err := lba.Evaluate()
	if err != nil {
		logger.Error("Detected an error when attempting to evaluate the alias [%s], Error [%s]", lba.Name,
			err.Error())
		t.Fail()
	} else if lba.Metric < 0 {
		logger.Error("Received a negative metric value [%d] when a positive number was expected. Failing test...",
			lba.Metric)
		t.Fail()
	}
}

// TestCommandFailFunctionality : fundamental functionality test for the [command]
func TestCommandFailFunctionality(t *testing.T) {
	logger.SetLevel(logger.FATAL)
	lba := lbalias.LBalias{Name: "command_load_functionality_test",
		Syslog:     true,
		ChecksDone: make(map[string]bool),
		ConfigFile: "../test/lbclient_failed_command.conf"}
	err := lba.Evaluate()
	if err == nil {
		logger.Error("An error was expected when attempting to evaluate the alias [%s]. Failing the test...", lba.Name)
		t.Fail()
	} else if lba.Metric > 0 {
		logger.Error("Received a positive metric value [%d] when a negative number was expected. Failing test...",
			lba.Metric)
		t.Fail()
	}
}
