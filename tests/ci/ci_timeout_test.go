package ci

import (
	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/mapping"
	"testing"
	"time"
)

// TestTimeoutNotKillExecution : Verifies that since the command execution takes less time than the timeout value of
// [9 seconds] the evaluation of the metric will be successful
func TestTimeoutNotKillExecution(t *testing.T) {
	logger.SetLevel(logger.ERROR)
	cfg := mapping.NewConfiguration("../test/lbclient_timeout_6s.conf", "myTest")

	err := lbconfig.Evaluate(cfg, time.Second*9)
	if err != nil {
		logger.Error("Received an error when expecting a successful run. Error [%s]. Failing the test...",
			err.Error())
		t.Fail()
	}
	if cfg.MetricValue < 0 {
		logger.Error("The metric output value returned negative [%d] when expecting a positive number." +
			" Failing the test...", cfg.MetricValue)
		t.Fail()
	}
}

// TestTimeoutKillExecution : Verifies that the timeout functionality aborts the execution of the evaluation if the
// timeout value of [3 seconds] is reached
func TestTimeoutKillExecution(t *testing.T) {
	logger.SetLevel(logger.FATAL)
	cfg := mapping.NewConfiguration("../test/lbclient_timeout_6s.conf", "myTest")

	err := lbconfig.Evaluate(cfg, time.Second*3)
	if err == nil {
		logger.Error("Expected an error to be returned due to the evaluation timeout of [3 seconds] against " +
			"the sleep command of [6 seconds]. Failing the test...")
		t.Fail()
	}
	if cfg.MetricValue > 0 {
		logger.Error("The metric output value returned positive [%d] when expecting a negative number." +
			" Failing the test...", cfg.MetricValue)
		t.Fail()
	}
}