package ci

import (
	"testing"
)

func TestCommand(t *testing.T) {
	myTests := []lbTest{

		lbTest{title: "Command", configuration: "../test/lbclient_command.conf", expectedMetricValue: 53},
		lbTest{title: "CommandDoesNotExist", configuration: "../test/lbclient_failed_command.conf", expectedMetricValue: -14},
		lbTest{title: "CommandFail", configuration: "../test/lbclient_command_failed.conf", expectedMetricValue: -14},
	}

	runMultipleTests(t, myTests)
}
