package ci

import (
	"testing"
)


func TestDaemonCheck(t *testing.T) {
	myTests := []lbTest{
		{title: "DaemonLegacyTest",
			configuration: "../test/daemon/lbclient_daemon_legacy_check.conf", 								expectedMetricValue: 5},
		{title: "DaemonNewSyntaxTest",
			configuration: "../test/daemon/lbclient_daemon_check.conf", 									expectedMetricValue: 5},
		{title: "DaemonFailPart1",
			configurationContent: "check daemon", 															expectedMetricValue: -17, shouldFail: true},
		{title: "DaemonFailPart2",
			configurationContent: "check daemon {}", 														expectedMetricValue: -17, shouldFail: true},
		{title: "DaemonFailPart3",
			configurationContent: `check daemon {"protocol":"tcp"}`, 										expectedMetricValue: -17, shouldFail: true},
		{title: "DaemonFailPart4",
			configurationContent: `check daemon {"part": 24, "protocol":"tcp"}`, 							expectedMetricValue: -17, shouldFail: true},
		{title: "DaemonFailPart5",
			configurationContent: `check daemon {"port":22, {}}`, 											expectedMetricValue: -17, shouldFail: true},
		{title: "DaemonFailPart6",
			configurationContent: `check daemon {"port":-1}`, 												expectedMetricValue: -17, shouldFail: true},
		{title: "DaemonFailPart7",
			configurationContent: `check daemon {"port":22, "ip":[-1,"a"]}`, 								expectedMetricValue: -17, shouldFail: true},
		{title: "DaemonFailPart8",
			configurationContent: `check daemon {"port":22, "protocol":0}`, 								expectedMetricValue: -17, shouldFail: true},
		{title: "DaemonFailPart9",
			configurationContent: `check daemon {"port":22, "protocol":{"key":"value"}}`, 					expectedMetricValue: -17, shouldFail: true},
		{title: "DaemonWarningPart1",
			configurationContent: "check daemon {\"port\":22, \"another_argument\":22}\nload constant 44", 	expectedMetricValue: 44},
		{title: "DaemonWarningPart2",
			configurationContent: "check daemon {\"port\":22, \"port\":22}\nload constant 45", 				expectedMetricValue: 45},
	}

	runMultipleTests(t, myTests)
}
