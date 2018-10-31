package ci

import (
	"gitlab.cern.ch/lb-experts/golbclient/lbalias"
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
	"testing"
)

// TestCollectdFunctionality : fundamental functionality test for the [collectd], output value must be = 9
func TestCollectdLoadFunctionality(t *testing.T) {
	logger.SetLevel(logger.ERROR)
	lba := lbalias.LBalias{Name: "collectd_load_functionality_test",
		Syslog:     true,
		ConfigFile: "../test/lbclient_collectd_load_single.conf"}
	err := lba.Evaluate()
	if err != nil {
		logger.Error("Detected an error when attempting to evaluate the alias [%s], Error [%s]", lba.Name, err.Error())
		t.Fail()
	}
	if lba.Metric != 98 {
		logger.Error("The expected metric value was [98] but got [%d] instead. Failing the test...", lba.Metric)
		t.Fail()
	}
}

// TestCollectdLoadConfigurationFile : integration test for all the functionality supplied by the collectdctl, resulting metric must be = 7
func TestCollectdLoadConfigurationFile(t *testing.T) {
	logger.SetLevel(logger.ERROR)

	lba := lbalias.LBalias{Name: "collectd_load_comprehensive_test",
		ConfigFile: "../test/lbclient_collectd_load.conf"}
	err := lba.Evaluate()
	if err != nil {
		logger.Error("Failed to run the client for the given configuration file [%s]. Error [%s]", lba.ConfigFile, err.Error())
		t.Fail()
	}
	if lba.Metric != 72 {
		logger.Error("The expected metric value was [72] but got [%d] instead. Failing the test...", lba.Metric)
		t.Fail()
	}
}

// TestCollectdLoadFailedConfigurationFile : integration test for all the functionality supplied by the lemon-cli, fail test
func TestCollectdLoadFailedConfigurationFile(t *testing.T) {
	logger.SetLevel(logger.FATAL)

	lba := lbalias.LBalias{Name: "collectd_load_intended_fail_test",
		ConfigFile: "../test/lbclient_collectd_load_fail.conf"}
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
