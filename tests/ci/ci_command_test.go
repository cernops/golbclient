package ci

import (
	"gitlab.cern.ch/lb-experts/golbclient/tests"
	"testing"
)

var commandTests []tests.LbTest

func init() {
	commandTests = []tests.LbTest{
		{Title: "Command",
			Configuration: "../test/lbclient_command.conf",
			ExpectedMetricValue: 53},
		{Title: "CommandDoesNotExist",
			Configuration: "../test/lbclient_failed_command.conf",
			ShouldFail: true,
			ExpectedMetricValue: -14},
		{Title: "CommandFail",
			Configuration: "../test/lbclient_command_failed.conf",
			ShouldFail: true,
			ExpectedMetricValue: -14},
		{Title: "NotZeroButNoError",
			ConfigurationContent: "check command false",
			ExpectedMetricValue: -14},
		{Title: "BinaryNotExecutable",
			ConfigurationContent: "check command ../test/command/commandNotExecutable",
			ExpectedMetricValue: -14,
			ShouldFail: true},
		{Title: "BinaryExecutableButExitNotZero",
			ConfigurationContent: "check command ../test/command/commandExecutableButExitNotZero",
			ExpectedMetricValue: -14},
		{Title: "BinaryExecutableExitZero",
			ConfigurationContent: "check command ../test/command/commandExecutableAndExitZero\nload constant 99",
			ExpectedMetricValue: 99},
	}
}

func TestCommand(t *testing.T) {
	tests.RunMultipleTests(t, commandTests)
}

func BenchmarkCommand(b *testing.B) {
	tests.RunMultipleTests(b, commandTests)
}
