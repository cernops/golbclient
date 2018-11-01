package ci

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gitlab.cern.ch/lb-experts/golbclient/lbalias"
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
)

// Parent directory for all daemon tests (fail & success)
var daemonTestsDir string

// Get environment variable
var loggerLevel logger.Level

func init() {
	daemonTestsDir = "../test/daemon"
	loggerLevel = logger.GetLevelByString(os.Getenv("TESTS_LOGGING_LEVEL"))
}

// TestDaemonFunctionality : fundamental functionality test the daemon checks
func TestDaemonFunctionality(t *testing.T) {
	logger.SetLevel(logger.TRACE)

	lba := lbalias.NewLbAlias(
		"daemon_functionality_test",
		true,
		fmt.Sprintf("%s/lbclient_daemon_check.conf", daemonTestsDir))
	err := lba.Evaluate()
	if err != nil {
		logger.Error("Detected an error when attempting to evaluate the alias [%s], Error [%s]", lba.Name, err.Error())
		t.Fail()
	}
	if lba.Metric < 0 {
		logger.Error("The metric output value returned negative [%d]. Failing the test...", lba.Metric)
		t.Fail()
	}
}

// TestDaemonFailedConfigurationFile : integration test for all the functionality supplied by the lemon-cli, fail tests
func TestDaemonFailedConfigurationFile(t *testing.T) {
	logger.SetLevel(logger.FATAL)

	// Read all fail tests
	failTestsFileNamePattern := "fail_part"
	var failTestFiles []string
	err := filepath.Walk(daemonTestsDir, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, failTestsFileNamePattern) {
			failTestFiles = append(failTestFiles, path)
		}
		return nil
	})
	if err != nil {
		logger.Fatal("Failed to read the test directory [%s]", daemonTestsDir)
	}

	// Run the tests on all files found
	for _, file := range failTestFiles {
		lba := lbalias.NewLbAlias(file, false, file)
		lba.Evaluate()

		if lba.Metric > 0 {
			logger.Fatal("The metric output value returned positive [%d] when expecting a negative output. Failing the test for [%s]...", lba.Metric, file)
			t.FailNow()
			break
		}
	}
}

// TestDaemonWarningConfigurationFile : integration test for all the functionality supplied by the lemon-cli, warning tests
func TestDaemonWarningConfigurationFile(t *testing.T) {
	logger.SetLevel(loggerLevel)

	// Read all fail tests
	failTestsFileNamePattern := "warning_part"
	var failTestFiles []string
	err := filepath.Walk(daemonTestsDir, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, failTestsFileNamePattern) {
			failTestFiles = append(failTestFiles, path)
		}
		return nil
	})
	if err != nil {
		logger.Fatal("Failed to read the test directory [%s]", daemonTestsDir)
	}

	// Run the tests on all files found
	for _, file := range failTestFiles {
		lba := lbalias.NewLbAlias(file, false, file)
		lba.Evaluate()

		if lba.Metric < 0 {
			logger.Error("The metric output value returned negative [%d] when expecting a positive output. Failing the test for [%s]...", lba.Metric, file)
			t.FailNow()
			break
		}
	}
}
