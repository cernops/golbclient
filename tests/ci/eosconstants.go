package ci

import (
	"os"
	"strconv"
	"testing"
	"time"

	logger "github.com/sirupsen/logrus"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/checks"
)

type eosTest struct {
	title               string
	procmounts          string
	eosxdcommand        string
	shouldFail          bool
	expectedMetricValue int
	timeout             time.Duration
	setup               func(*testing.T)
	cleanup             func(*testing.T)
	contextLogger       *logger.Entry
}

func runEOSCheck(t *testing.T, test eosTest) bool {
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

	procmounts := "/proc/mounts"
	if test.procmounts != "" {
		procmounts = test.procmounts
	}
	f, err := os.Open(procmounts)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { err = f.Close() }()

	testCmdBase := "../../scripts/eosxd get eos.mgmurl "
	if test.eosxdcommand != "" {
		testCmdBase = test.eosxdcommand + " get eos.mgmurl "
	}

	if test.timeout == 0 {
		test.timeout = defaultTimeout
	}
	metricValue, err := checks.DoEOSCheck(f, testCmdBase, test.contextLogger)
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
	if metricValue != test.expectedMetricValue {
		test.contextLogger.WithFields(logger.Fields{
			"RECEIVED": metricValue,
			"EXPECTED": test.expectedMetricValue,
		}).Error("Failed to receive the expected metric value. Failing the test...")
		t.FailNow()
		return false
	}
	return true
}

func runMultipleEosTests(t *testing.T, myTests []eosTest) {
	logger.SetLevel(logger.FatalLevel)
	for _, myTest := range myTests {
		logger.Infof("Running the test [%v]", myTest.title)
		if t.Run(myTest.title, func(t *testing.T) {
			runEOSCheck(t, myTest)
		}) != true {
			logger.Errorf("The command [%v] failed. Repeating with [TRACE] verbose level...", myTest.title)
			logger.SetLevel(logger.TraceLevel)
			runEOSCheck(t, myTest)
			logger.SetLevel(logger.ErrorLevel)
		}
	}
}
