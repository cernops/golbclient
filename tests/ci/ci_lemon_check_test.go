package ci

import (
	"strings"
	"testing"

	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/utils/runner"
)

// TestCILemonCLI : checks if the alternative [lemon-cli] used in the CI pipeline is OK
func TestCILemonCLI(t *testing.T) {
	logger.SetLevel(logger.ERROR)
	output, err := runner.Run("/usr/sbin/lemon-cli",
		true, defaultTimeout, "--script", "-m", "13163")
	if err != nil {
		logger.Error("An error was detected when running the CI [lemon-cli]")
		t.FailNow()
	} else if len(strings.TrimSpace(output)) == 0 {
		logger.Error("The CI [lemon-cli] failed to return a row value for a pre-defined metric")
		t.FailNow()
	}
	logger.Trace("CI [lemon-cli] output [%s]", output)
}

func TestLemon(t *testing.T) {
	var myTests [3]lbTest
	myTests[0] = lbTest{"CollectdFunctionality", "../test/lbclient_lemon_check_single.conf", true, 12, nil, nil}
	myTests[1] = lbTest{"ConfigurationFile", "../test/lbclient_lemon_check.conf", true, 8, nil, nil}
	myTests[2] = lbTest{"LemonFailed", "../test/lbclient_lemon_check_fail.conf", false, -12, nil, nil}

	runMultipleTests(t, myTests[:])
}
