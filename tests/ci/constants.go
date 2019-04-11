package ci

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/mapping"
)

// defaultTimeout : timeout value to be used when executing the tests
const defaultTimeout = time.Second * 30

type lbTest struct {
	title                string
	configuration        string
	configurationContent string
	shouldWork           bool
	expectedMetricValue  int
	validateConfig       bool
	setup                func(*testing.T)
	cleanup              func(*testing.T)
}

func runEvaluate(t *testing.T, test lbTest) bool {
	if myTest.setup != nil {
		myTest.setup(t)
	}
	var configFile string
	if myTest.configuration != "" {
		configFile = myTest.configuration
	} else {
		file, err := ioutil.TempFile("/tmp", "lbclient_test")
		if err != nil {
			t.FailNow()
		}
		_, err = file.WriteString(myTest.configurationContent)
		if err != nil {
			t.FailNow()
		}
		defer os.Remove(file.Name())
		configFile = file.Name()
	}
	cfg := mapping.NewConfiguration(configFile)
	err := lbconfig.Evaluate(cfg, defaultTimeout, myTest.validateConfig)
	if myTest.cleanup != nil {
		defer myTest.cleanup(t)
	}
	if myTest.shouldWork == true {
		if err != nil {
			logger.Error("Got back an error, and we were not expecting that. Error [%s] ", err.Error())
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
	if cfg.MetricValue != myTest.metricValue {
		logger.Error("We were expecting the value [%v], and got [%v]", myTest.metricValue, cfg.MetricValue)
		t.FailNow()
		return false
	}
	return true
}

func runMultipleTests(t *testing.T, myTests []lbTest) {
	logger.SetLevel(logger.ERROR)
	for _, myTest := range myTests {
		logger.Info("Running the test [%v]", myTest.title)
		if t.Run(myTest.title, func(t *testing.T) {
			runEvaluate(t, myTest)
		}) != true {
			logger.Error("The command [%v] failed. Repeating with [TRACE] verbose level...", myTest.title)
			logger.SetLevel(logger.TRACE)
			runEvaluate(t, myTest)
			logger.SetLevel(logger.ERROR)
		}
	}
}
