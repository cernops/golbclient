package tests

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

// DefaultTimeout : Timeout value to be used when executing the tests
const DefaultTimeout = time.Second * 30

type LbTest struct {
	Title                string
	Configuration        string
	ConfigurationContent string
	ShouldFail           bool
	ExpectedMetricValue  int
	ValidateConfig       bool
	Timeout              time.Duration
	Setup                func(testing.TB)
	Cleanup              func(testing.TB)
	ContextLogger        *logger.Entry
}

func RunEvaluate(t testing.TB, test LbTest) bool {
	if test.ContextLogger == nil {
		test.ContextLogger = logger.WithFields(logger.Fields{
			"TEST":			test.Title,
			"SHOULD_FAIL":	strconv.FormatBool(test.ShouldFail),
			"SETUP":		test.Setup,
			"CLEANUP":		test.Cleanup,
		})
	}

	if test.Setup != nil {
		test.Setup(t)
	}
	var configFile string
	if test.Configuration != "" {
		configFile = test.Configuration
	} else {
		file, err := ioutil.TempFile("/tmp", "lbclient_test")
		if err != nil {
			t.Fatal(err)
		}

		if _, err = file.WriteString(test.ConfigurationContent); err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := os.Remove(file.Name()); err != nil {
				test.ContextLogger.Warnf("An error occurred when attempting to remove the temporary test file [%s]. " +
					"Error [%s]", file.Name(), err.Error())
			}
		}()
		configFile = file.Name()
	}
	cfg := mapping.NewConfiguration(configFile)

	if test.Timeout == 0 {
		test.Timeout = DefaultTimeout
	}
	err := lbconfig.Evaluate(cfg, test.Timeout, test.ValidateConfig)
	if test.Cleanup != nil {
		defer test.Cleanup(t)
	}
	if test.ShouldFail {
		if err == nil {
			test.ContextLogger.Fatal("A null error was received when an evaluation error was expected. Failing the test...")
			t.Fail()
			return false
		}
	} else {
		if err != nil {
			test.ContextLogger.WithError(err).Fatal("An unexpected error was received during the evaluation.")
			t.Fail()
			return false
		}
	}
	if cfg.MetricValue != test.ExpectedMetricValue {
		test.ContextLogger.WithFields(logger.Fields{
			"RECEIVED": cfg.MetricValue,
			"EXPECTED": test.ExpectedMetricValue,
		}).Error("Failed to receive the expected metric value. Failing the test...")
		t.Fail()
		return false
	}
	return true
}

func RunMultipleTests(t testing.TB, myTests []LbTest) {
	logger.SetLevel(logger.FatalLevel)

	for _, myTest := range myTests {
		logger.Infof("Running the test [%v]", myTest.Title)
		if !runTest(t, myTest) {
			logger.Errorf("The command [%v] failed. Repeating with [TRACE] verbose level...", myTest.Title)
			logger.SetLevel(logger.TraceLevel)
			RunEvaluate(t, myTest)
			logger.SetLevel(logger.ErrorLevel)
		}
	}
}

func runTest(testB testing.TB, myTest LbTest) bool {
	if test, ok := testB.(*testing.T); ok {
		return test.Run(myTest.Title, func(t *testing.T) {
			if !RunEvaluate(testB, myTest) {
				t.FailNow()
			}
		})
	} else if bench, ok := testB.(*testing.B); ok {
		return bench.Run(myTest.Title, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if !RunEvaluate(testB, myTest) {
					b.FailNow()
				}
			}
		})
	}

	return false
}
