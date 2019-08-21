package ci

import (
	"testing"
)

func TestCommand(t *testing.T) {
	myTests := []lbTest{
		{title: "Command",
			configuration: "../test/lbclient_command.conf",
			expectedMetricValue: 53},
		{title: "CommandDoesNotExist",
			configuration: "../test/lbclient_failed_command.conf",
			shouldFail: true,
			expectedMetricValue: -14},
		{title: "CommandFail",
			configuration: "../test/lbclient_command_failed.conf",
			shouldFail: true,
			expectedMetricValue: -14},
		{title: "NotZeroButNoError",
			configurationContent: "check command false",
			expectedMetricValue: -14},
		{title: "BinaryNotExecutable",
			configurationContent: "check command ../test/command/commandNotExecutable",
			expectedMetricValue: -14,
			shouldFail: true},
		{title: "BinaryExecutableButExitNotZero",
			configurationContent: "check command ../test/command/commandExecutableButExitNotZero",
			expectedMetricValue: -14},
		{title: "BinaryExecutableExitZero",
			configurationContent: "check command ../test/command/commandExecutableAndExitZero\nload constant 99",
			expectedMetricValue: 99},
	}

	runMultipleTests(t, myTests)
}
