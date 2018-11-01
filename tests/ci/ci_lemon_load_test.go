package ci

import (
	"testing"

	"gitlab.cern.ch/lb-experts/golbclient/lbalias"
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
)

// TestLemonLoadFunctionality : fundamental functionality test for the [lemon-cli], output value must be = 1
func TestLemonLoadFunctionality(t *testing.T) {
	logger.SetLevel(logger.ERROR)
	lba := lbalias.NewLbAlias(
		"myTest",
		true,
		"../test/lbclient_lemon_load_single.conf")
	err := lba.Evaluate()
	if err != nil {
		logger.Error("Detected an error when attempting to evaluate the alias [%s], Error [%s]", lba.Name, err.Error())
		t.Fail()
	}
	if lba.Metric != 1 {
		logger.Error("The expected metric value was [1] but got [%d] instead. Failing the test...", lba.Metric)
		t.Fail()
	}
}

// TestLemonLoadConfigurationFile : integration test for all the functionality supplied by the lemon-cli, output value must be = 35
func TestLemonLoadConfigurationFile(t *testing.T) {
	logger.SetLevel(logger.ERROR)

	lba := lbalias.NewLbAlias(
		"lemonTest",
		false,
		"../test/lbclient_lemon_load.conf")
	err := lba.Evaluate()
	if err != nil {
		logger.Error("Failed to run the client for the given configuration file [%s]. Error [%s]", lba.ConfigFile, err.Error())
		t.Fail()
	}
	if lba.Metric != 27 {
		logger.Error("The expected metric value was [27] but got [%d] instead. Failing the test...", lba.Metric)
		t.Fail()
	}
}

// TestLemonFailedConfigurationFile : integration test for all the functionality supplied by the lemon-cli, fail test
func TestLemonLoadFailedConfigurationFile(t *testing.T) {
	logger.SetLevel(logger.FATAL)

	lba := lbalias.NewLbAlias(
		"lemonFailTest",
		false,
		"../test/lbclient_lemon_load_fail.conf")
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
