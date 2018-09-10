package ci

import (
	"gitlab.cern.ch/lb-experts/golbclient/lbalias"
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
	"testing"
)


// TestDaemonFunctionality : fundamental functionality test the daemon checks
func TestDaemonFunctionality(t *testing.T) {
	logger.SetLevel(logger.TRACE)
	lba := lbalias.LBalias{Name: "daemon_functionality_test",
		Syslog:     true,
		ChecksDone: make(map[string]bool),
		ConfigFile: "../test/lbclient_daemon_check.conf"}
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

/*
// TestLemonFailedConfigurationFile : integration test for all the functionality supplied by the lemon-cli, fail test
func TestDaemonFailedConfigurationFile(t *testing.T) {
	logger.SetLevel(logger.FATAL)

	lba := lbalias.LBalias{Name: "daemon_intented_fail_test",
		ChecksDone: make(map[string]bool),
		ConfigFile: "../test/lbclient_lemon_deamon_fail.conf"}
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
*/