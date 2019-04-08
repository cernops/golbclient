package ci

import (
	"testing"

	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/mapping"

	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig"
)

func TestLemonLoad(t *testing.T) {
	var myTests [3]lbTest
	myTests[0] = lbTest{"LemonLoadSingle", "../test/lbclient_lemon_load_single.conf", true, 1, nil, nil}
	myTests[1] = lbTest{"LemonLoad", "../test/lbclient_lemon_load.conf", true, 27, nil, nil}
	myTests[2] = lbTest{"LemonFailed", "../test/lbclient_lemon_check_fail.conf", false, -11, nil, nil}

	runMultipleTests(t, myTests[:])
}

// TestLemonFailedConfigurationFile : integration test for all the functionality supplied by the lemon-cli, fail test
func TestLemonLoadFailedConfigurationFile(t *testing.T) {
	logger.SetLevel(logger.FATAL)

	cfg := mapping.NewConfiguration("../test/lbclient_lemon_load_fail.conf", "lemonFailTest")
	err := lbconfig.Evaluate(cfg, defaultTimeout)
	if err == nil {
		logger.Error("Expected an error for the given configuration file [%s]. Failing test...", cfg.ConfigFilePath)
		t.Fail()
	}
	if cfg.MetricValue >= 0 {
		logger.Error("The metric output value returned positive [%d] when expecting a negative output. Failing the test...", cfg.MetricValue)
		t.Fail()
	}
}
