package ci

import (
	"strings"
	"testing"

	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/utils/runner"
)

// TestCICollectdCLI : checks if the alternative [collectd] used in the CI pipeline is OK
func TestCICollectdCLI(t *testing.T) {
	logger.SetLevel(logger.ERROR)
	output, err := runner.Run("/usr/bin/collectdctl",
		true, defaultTimeout, "getval", "test")
	if err != nil {
		logger.Error("An error was detected when running the CI [collectdctl]")
		t.FailNow()
	} else if len(strings.TrimSpace(output)) == 0 {
		logger.Error("The CI [collectdctl] failed to return a row value for a pre-defined metric")
		t.FailNow()
	}
	logger.Trace("CI [collectdctl] output [%s]", output)
}

func TestCollectd(t *testing.T) {
	var myTests [6]lbTest
	myTests[0] = lbTest{"CollectdFunctionality", "../test/lbclient_collectd_check_single.conf", true, 5, nil, nil}
	myTests[1] = lbTest{"ConfigurationFile", "../test/lbclient_collectd_check.conf", true, 3, nil, nil}
	myTests[2] = lbTest{"ConfigurationFileWithKeys", "../test/lbclient_collectd_check_with_keys.conf", true, 7, nil, nil}
	myTests[3] = lbTest{"FailedConfigurationFile", "../test/lbclient_collectd_check_fail.conf", false, -15, nil, nil}
	myTests[4] = lbTest{"FailedConfigurationFileWithWrongKey", "../test/lbclient_collectd_check_fail_with_wrong_key.conf", false, -15, nil, nil}
	myTests[5] = lbTest{"FailedConfigurationFileWithEmptyKey", "../test/lbclient_collectd_check_fail_with_empty_key.conf", false, -15, nil, nil}

	runMultipleTests(t, myTests[:])
}
