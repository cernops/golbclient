package ci

import (
	"io/ioutil"
	"os"
	"strconv"
	"testing"
	"time"

	logger "github.com/sirupsen/logrus"
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
	contextLogger        *logger.Entry
}

func runEvaluate(t *testing.T, test lbTest) bool {
	if test.contextLogger == nil {
		test.contextLogger = logger.WithFields(logger.Fields{
			"TEST":        test.title,
			"SHOULD_FAIL": strconv.FormatBool(test.shouldFail),
			"SETUP":       test.setup,
			"CLEANUP":     test.cleanup,
		})
	}

	if test.setup != nil {
		test.setup(t)
	}
	var configFile string
	if test.configuration != "" {
		configFile = test.configuration
	} else {
		file, err := ioutil.TempFile("/tmp", "lbclient_test")
		if err != nil {
			t.Fatal(err)
		}

		if _, err = file.WriteString(test.configurationContent); err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := os.Remove(file.Name()); err != nil {
				test.contextLogger.Warnf("An error occurred when attempting to remove the temporary test file [%s]. "+
					"Error [%s]", file.Name(), err.Error())
			}
		}()
		configFile = file.Name()
	}
	cfg := mapping.NewConfiguration(configFile)

	if test.timeout == 0 {
		test.timeout = defaultTimeout
	}
	err := lbconfig.Evaluate(cfg, test.timeout, test.validateConfig)
	if test.cleanup != nil {
		defer test.cleanup(t)
	}
	if test.shouldFail {
		if err == nil {
			test.contextLogger.Fatal("A null error was received when an evaluation error was expected. Failing the test...")
			t.FailNow()
			return false
		}
	} else {
		if err != nil {
			test.contextLogger.WithError(err).Fatal("An unexpected error was received during the evaluation.")
			t.FailNow()
			return false
		}
	}
	if cfg.MetricValue != test.expectedMetricValue {
		test.contextLogger.WithFields(logger.Fields{
			"RECEIVED": cfg.MetricValue,
			"EXPECTED": test.expectedMetricValue,
		}).Error("Failed to receive the expected metric value. Failing the test...")
		t.FailNow()
		return false
	}
	return true
}

func runMultipleTests(t *testing.T, myTests []lbTest) {
	logger.SetLevel(logger.FatalLevel)
	for _, myTest := range myTests {
		logger.Infof("Running the test [%v]", myTest.title)
		if !t.Run(myTest.title, func(t *testing.T) {
			runEvaluate(t, myTest)
		}){
			logger.Errorf("The command [%v] failed. Repeating with [TRACE] verbose level...", myTest.title)
			logger.SetLevel(logger.TraceLevel)
			runEvaluate(t, myTest)
			logger.SetLevel(logger.ErrorLevel)
		}
	}
}
