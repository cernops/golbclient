package ci

import (
	"fmt"
	"testing"
	"time"

	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/mapping"
)

// defaultTimeout : timeout value to be used when executing the tests
const defaultTimeout = time.Second * 30

type lbTest struct {
	title         string
	configuration string
	shouldWork    bool
	metricValue   int
	setup         func(*testing.T)
	cleanup       func(*testing.T)
}

func runEvaluate(t *testing.T, configFile string, shouldWork bool, metricValue int, setup func(*testing.T), cleanup func(*testing.T)) bool {
	if setup != nil {
		fmt.Printf("Calling the setup function")
		setup(t)
	}
	cfg := mapping.NewConfiguration(configFile)
	err := lbconfig.Evaluate(cfg, defaultTimeout)
	if cleanup != nil {
		fmt.Printf("Setting the call for the cleanup")
		defer cleanup(t)
	}
	if shouldWork == true {
		if err != nil {
			logger.Error("Got back an error, and we  were not expecting that. Error [%s] ", err.Error())
			t.FailNow()
			return false
		}
	} else {
		if err == nil {
			logger.Error("The evaluation of the alias was supposed to fail")
			t.FailNow()
			return false
		}
	}
	if cfg.MetricValue != metricValue {
		logger.Error("We were expecting the value %v, and got %v", metricValue, cfg.MetricValue)
		t.FailNow()
		return false
	}
	return true
}

func runMultipleTests(t *testing.T, myTests []lbTest) {
	for _, myTest := range myTests {
		logger.Info("Running th test %v\n", myTest.title)
		if t.Run(myTest.title, func(t *testing.T) {
			runEvaluate(t, myTest.configuration, myTest.shouldWork, myTest.metricValue, myTest.setup, myTest.cleanup)
		}) != true {
			logger.Error("The command failed. Let's repeat it a bit more verbose")
			logger.SetLevel(logger.TRACE)
			runEvaluate(t, myTest.configuration, myTest.shouldWork, myTest.metricValue, myTest.setup, myTest.cleanup)
			logger.SetLevel(logger.ERROR)
		}
	}
}
