package ci

import (
	"gitlab.cern.ch/lb-experts/golbclient/lbalias"
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
	"os"
	"path/filepath"
	"strings"
	"testing"
)


// TestDaemonFunctionality : fundamental functionality test the daemon checks
func TestDaemonFunctionality(t *testing.T) {
	logger.SetLevel(logger.ERROR)
	lba := lbalias.LBalias{Name: "daemon_functionality_test",
		Syslog:     true,
		ChecksDone: make(map[string]bool),
		ConfigFile: "../test/daemon/lbclient_daemon_check.conf"}
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

// TestLemonFailedConfigurationFile : integration test for all the functionality supplied by the lemon-cli, fail test
func TestDaemonFailedConfigurationFile(t *testing.T) {
	logger.SetLevel(logger.FATAL)

	// Read all fail tests
	failTestDir := "../test/daemon/"
	failTestsFileNamePattern := "fail_part"
	var failTestFiles []string
	err := filepath.Walk(failTestDir, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, failTestsFileNamePattern) {
			failTestFiles = append(failTestFiles, path)
		}
		return nil
	})
	if  err != nil {
		logger.Fatal("Failed to read the test directory [%s]", failTestDir)
	}

	// Run the tests on all files found
	for _, file := range failTestFiles {
		lba := lbalias.LBalias{Name: file,
			ChecksDone: make(map[string]bool),
			ConfigFile: file}
		lba.Evaluate()

		if lba.Metric > 0 {
			logger.Error("The metric output value returned positive [%d] when expecting a negative output. Failing the test for [%s]...", lba.Metric, file)
			t.FailNow()
			break
		}
	}
}
