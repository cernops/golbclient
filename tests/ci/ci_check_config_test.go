package ci

import (
	"testing"
)

//TestCheckConfig Verify that the option to check the configuration file works
func TestCheckConfig(t *testing.T) {

	myTests := []lbTest{
		lbTest{title: "Valid syntax", configurationContent: "dasdas #With comments\n #More comments\n\ncheck roger\ncheck nologin\ncheck afs\nload constant 5", expectedMetricValue: 5, validateConfig: true, setup: createRogerFileDraining},
		lbTest{title: "Roger exists on valid syntax", configurationContent: "#With comments\n #More comments\n\ncheck roger\ncheck nologin\ncheck afs\nload constant 5", expectedMetricValue: -13, setup: createRogerFileDraining},
		lbTest{title: "Wrong check", configurationContent: "check blabla", shouldFail: true, expectedMetricValue: -1, validateConfig: true},
		lbTest{title: "Wrong check constant", configurationContent: "check constant 4", shouldFail: true, expectedMetricValue: -1, validateConfig: true},
		lbTest{title: "Wrong lemon expression", configurationContent: "check lemon 5 + ", shouldFail: true, expectedMetricValue: -12, validateConfig: true},
		lbTest{title: "Wrong expression", configurationContent: "check roger\ncheck nologin\nload constant 5\nload constant 4\ncheck lemon 5 + ", shouldFail: true, expectedMetricValue: -12, validateConfig: true},
		lbTest{title: "Wrong collectd expression", configurationContent: "load collectd 4 + [dasdas ", shouldFail: true, expectedMetricValue: -15, validateConfig: true},
	}
	runMultipleTests(t, myTests)

}
