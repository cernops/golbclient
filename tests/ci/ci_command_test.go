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
		logger.Error("Detected an error when attempting to evaluate the alias [%s], Error [%s]", lba.Name, err.Error())
		t.Fail()
	}
}
