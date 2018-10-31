package ci

import (
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
	"gitlab.cern.ch/lb-experts/golbclient/lbalias"
	"gitlab.cern.ch/lb-experts/golbclient/lbalias/utils/runner"
	"testing"
	"strings"
)


// TestCILemonCLI : checks if the alternative [lemon-cli] used in the CI pipeline is OK
func TestCILemonCLI(t *testing.T) {
	logger.SetLevel(logger.ERROR)
	output, err := runner.RunCommand("/usr/sbin/lemon-cli",
		true, true, "--script", "-m", "13163")
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
	lba := lbalias.LBalias{Name: "myTest",
		Syslog:     true,
		ConfigFile: "../test/lbclient_lemon_check_single.conf"}
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

// TestLemonConfigurationFile : integration test for all the functionality supplied by the lemon-cli
func TestLemonConfigurationFile(t *testing.T) {
	logger.SetLevel(logger.ERROR)

	lba := lbalias.LBalias{Name: "lemonTest",
		ConfigFile: "../test/lbclient_lemon_check.conf"}
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

// TestLemonFailedConfigurationFile : integration test for all the functionality supplied by the lemon-cli, fail test
func TestLemonFailedConfigurationFile(t *testing.T) {
	logger.SetLevel(logger.FATAL)

	lba := lbalias.LBalias{Name: "lemonFailTest",
		ConfigFile: "../test/lbclient_lemon_check_fail.conf"}
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
