package ci

import (
	"testing"
)

func TestCommand(t *testing.T) {
	myTests := []lbTest{
		{title: "Command",
			configuration: "../test/lbclient_command.conf", expectedMetricValue: 53},
		{title: "CommandDoesNotExist",
			configuration: "../test/lbclient_failed_command.conf", shouldFail: true, expectedMetricValue: -14},
		{title: "CommandFail",
			configuration: "../test/lbclient_command_failed.conf", shouldFail: true, expectedMetricValue: -14},
	}

	runMultipleTests(t, myTests)
}
