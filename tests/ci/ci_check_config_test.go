package ci

import (
	"gitlab.cern.ch/lb-experts/golbclient/tests"
	"testing"
)

var configTests []tests.LbTest

func init() {
	configTests = []tests.LbTest{
		{Title: "Valid syntax",
			ConfigurationContent: "#With comments\n #More comments\n\ncheck roger\ncheck nologin\ncheck afs\nload constant 5",
			ExpectedMetricValue: 5, ValidateConfig: true, Setup: createRogerFileDraining},
		{Title: "Roger exists on valid syntax",
			ConfigurationContent: "#With comments\n #More comments\n\ncheck roger\ncheck nologin\ncheck afs\nload constant 5",
			ExpectedMetricValue: -13, Setup: createRogerFileDraining},
		{Title: "Wrong check",
			ConfigurationContent: "check blabla", ShouldFail: true, ExpectedMetricValue: -1, ValidateConfig: true},
		{Title: "Wrong check constant",
			ConfigurationContent: "check constant 4", ShouldFail: true, ExpectedMetricValue: -1, ValidateConfig: true},
		{Title: "Wrong lemon expression",
			ConfigurationContent: "check lemon 5 + ", ShouldFail: true, ExpectedMetricValue: -12, ValidateConfig: true},
		{Title: "Wrong expression",
			ConfigurationContent: "check roger\ncheck nologin\nload constant 5\nload constant 4\ncheck lemon 5 + ",
			ShouldFail: true, ExpectedMetricValue: -12, ValidateConfig: true},
		{Title: "Wrong collectd expression",
			ConfigurationContent: "load collectd 4 + [dasdas ", ShouldFail: true, ExpectedMetricValue: -15, ValidateConfig: true},
	}
}

//TestCheckConfig Verify that the option to check the configuration file works
func TestCheckConfig(t *testing.T) {
	tests.RunMultipleTests(t, configTests)
}

func BenchmarkCheckConfig(b *testing.B) {
	tests.RunMultipleTests(b, configTests)
}
