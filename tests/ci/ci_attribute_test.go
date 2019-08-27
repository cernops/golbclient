package ci

import (
	"gitlab.cern.ch/lb-experts/golbclient/tests"
	"testing"
)

var attributeTests []tests.LbTest

func init() {
	attributeTests = []tests.LbTest{
		{Title: "TestXsessions", ConfigurationContent: "check xsessions\nload constant 34", ExpectedMetricValue: 34},
		{Title: "TestSwapping", ConfigurationContent: "check swapping\nload constant 35", ExpectedMetricValue: 35},
		{Title: "TestSwapping_backwards_compatible", ConfigurationContent: "check swaping\nload constant 36", ExpectedMetricValue: 36},
		{Title: "TestAttributeBroken", ConfigurationContent: "check bbswapping\nload constant 35", ExpectedMetricValue: -1, ShouldFail: true},
	}
}

func TestAttribute(t *testing.T) {
	tests.RunMultipleTests(t, attributeTests)
}

func BenchmarkAttribute(b *testing.B) {
	tests.RunMultipleTests(b, attributeTests)
}