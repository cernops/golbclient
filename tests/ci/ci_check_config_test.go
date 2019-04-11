package ci

import (
	"testing"
)

//TestCheckConfig Verify that the option to check the configuration file works
func TestCheckConfig(t *testing.T) {

	myTests := []lbTest{
		lbTest{title: "Valid syntax", configurationContent: "#With comments\n #More comments\n\ncheck roger\ncheck nologin\ncheck afs\nload constant 5", expectedMetricValue: 5},
		lbTest{title: "Wrong check", configurationContent: "check blabla", shouldFail: true, expectedMetricValue: -1},
		lbTest{title: "Wrong check constant", configurationContent: "check constant 4", shouldFail: true, expectedMetricValue: -1},
		lbTest{title: "Wrong lemon expression", configurationContent: "check lemon 5 + ", shouldFail: true, expectedMetricValue: -12},
		lbTest{title: "Wrong expression", configurationContent: "check roger\ncheck nologin\nload constant 5\nload constant 4\ncheck lemon 5 + ", shouldFail: false, expectedMetricValue: -12},
		lbTest{title: "Wrong collectd expression", configurationContent: "load collectd 4 + [dasdas ", shouldFail: true, expectedMetricValue: -15},
	}
	runMultipleTests(t, myTests)

}
