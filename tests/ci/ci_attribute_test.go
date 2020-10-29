package ci

import (
	"testing"
)

func TestAttributest(t *testing.T) {
	myTests := []lbTest{
		{title: "TestXsessions", configurationContent: "check xsessions\nload constant 34", expectedMetricValue: 34},
		{title: "TestSwapping", configurationContent: "check swapping\nload constant 35", expectedMetricValue: 35},
		{title: "TestSwapping_backwards_compatible", configurationContent: "check swaping\nload constant 36", expectedMetricValue: 36},
		{title: "TestAttributeBroken", configurationContent: "check bbswapping\nload constant 35", expectedMetricValue: -1, shouldFail: true},
	}

	runMultipleTests(t, myTests)
}
