//+build linux darwin

package ci

import (
	"github.com/stretchr/testify/assert"
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
	shouldFail           bool
	expectedMetricValue  int
	validateConfig       bool
	timeout              time.Duration
	setup                func(*testing.T)
	cleanup              func(*testing.T)
}

func (l *lbTest) setupConfigurationFile(t *testing.T) {
	if l.configuration != "" {
		return
	} else {
		file, err := ioutil.TempFile("/tmp", "lbclient_test")
		assert.Nil(t, err, "An unexpected error occurred when attempting to create the temporary test file. " +
				"Error [%v]", err)

		_, err = file.WriteString(l.configurationContent)
		assert.Nil(t, err, "An unexpected error occurred when attempting to write the temporary test file [%s]." +
				" Error [%v]", file.Name(), err)

		defer func() {
			if err := os.Remove(file.Name()); err != nil {
				logger.Warn("An unexpected error occurred when attempting to close the file [%s]. Error [%s]",
					file.Name(), err.Error())
			}
		}()
		l.configuration = file.Name()
	}
}

func runEvaluate(t *testing.T, test lbTest) bool {
	assert.NotNil(t, test)
	logger.Debug("Attempting to run the test [%v]...", test.title)
	if test.setup != nil {
		logger.Trace("Executing the [setUp] function [%v]...", test.setup)
		test.setup(t)
	}
	if test.cleanup != nil {
		logger.Trace("Executing the [tearDown] function [%v]...", test.cleanup)
		defer test.cleanup(t)
	}

	test.setupConfigurationFile(t)
	cfg := mapping.NewConfiguration(test.configuration)

	if test.timeout == 0 {	test.timeout = defaultTimeout	}
	err := lbconfig.Evaluate(cfg, test.timeout, test.validateConfig)

	if test.shouldFail {
		if err == nil {
			t.Fatalf("Received a null error when expecting the test to fail. Failing test...")
			return false
		}
	} else {
		if err != nil {
			t.Fatalf("Unexpected error received. Error [%s] ", err.Error())
			return false
		}
	}
	if cfg.MetricValue != test.expectedMetricValue {
		t.Fatalf("Received the metric value [%v] but was expecting [%v]. Failing the test...",
			cfg.MetricValue, test.expectedMetricValue)
		return false
	}
	return true
}

func runMultipleTests(t *testing.T, myTests []lbTest) {
	logger.SetLevel(logger.FATAL)
	for _, myTest := range myTests {
		if t.Run(myTest.title, func(t *testing.T) {
			runEvaluate(t, myTest)
		}) != true {
			logger.Error("The command [%v] failed. Repeating with [TRACE] verbosity level...", myTest.title)
			logger.SetLevel(logger.TRACE)
			runEvaluate(t, myTest)
			logger.SetLevel(logger.ERROR)
		}
	}
}
