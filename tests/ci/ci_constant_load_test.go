package ci

import (
	"testing"

	"gitlab.cern.ch/lb-experts/golbclient/lbalias"
	"gitlab.cern.ch/lb-experts/golbclient/utils"
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
)

// TestConstantLoadFunctionality : fundamental functionality test for the [load constant], output value must be = 250
func TestConstantLoadFunctionality(t *testing.T) {
	logger.SetLevel(logger.ERROR)
	cfg := utils.NewConfiguration("../test/lbclient_constant.conf", "constant_load_functionality_test")
	err := lbalias.Evaluate(cfg)
	if err != nil {
		logger.Error("Detected an error when attempting to evaluate the alias [%s], Error [%s]", cfg.ConfigFilePath, err.Error())
		t.Fail()
	}
	if cfg.MetricValue != 250 {
		logger.Error("The expected metric value was [250] but got [%d] instead. Failing the test...", cfg.MetricValue)
		t.Fail()
	}
}