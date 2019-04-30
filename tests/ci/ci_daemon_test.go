package ci

import (
	"testing"
)


func TestDaemonCheck(t *testing.T) {
	myTests := []lbTest{
		{title: "DaemonLegacyTest",
			configuration: "../test/daemon/lbclient_daemon_legacy_check.conf", 			expectedMetricValue: 5},
		{title: "DaemonNewSyntaxTest",
			configuration: "../test/daemon/lbclient_daemon_check.conf", 				expectedMetricValue: 5},
		{title: "DaemonFailPart1",
			configuration: "../test/daemon/lbclient_daemon_check_fail_part1.conf", 		expectedMetricValue: -17, shouldFail: true},
		{title: "DaemonFailPart2",
			configuration: "../test/daemon/lbclient_daemon_check_fail_part2.conf", 		expectedMetricValue: -17, shouldFail: true},
		{title: "DaemonFailPart3",
			configuration: "../test/daemon/lbclient_daemon_check_fail_part3.conf", 		expectedMetricValue: -17, shouldFail: true},
		{title: "DaemonFailPart4",
			configuration: "../test/daemon/lbclient_daemon_check_fail_part4.conf", 		expectedMetricValue: -17, shouldFail: true},
		{title: "DaemonFailPart5",
			configuration: "../test/daemon/lbclient_daemon_check_fail_part5.conf", 		expectedMetricValue: -17, shouldFail: true},
		{title: "DaemonFailPart6",
			configuration: "../test/daemon/lbclient_daemon_check_fail_part6.conf", 		expectedMetricValue: -17, shouldFail: true},
		{title: "DaemonFailPart7",
			configuration: "../test/daemon/lbclient_daemon_check_fail_part7.conf", 		expectedMetricValue: -17, shouldFail: true},
		{title: "DaemonFailPart8",
			configuration: "../test/daemon/lbclient_daemon_check_fail_part8.conf", 		expectedMetricValue: -17, shouldFail: true},
		{title: "DaemonFailPart9",
			configuration: "../test/daemon/lbclient_daemon_check_fail_part9.conf", 		expectedMetricValue: -17, shouldFail: true},
		{title: "DaemonWarningPart1",
			configuration: "../test/daemon/lbclient_daemon_check_warning_part1.conf", 	expectedMetricValue: 5},
		{title: "DaemonWarningPart2",
			configuration: "../test/daemon/lbclient_daemon_check_warning_part2.conf", 	expectedMetricValue: 5},
	}

	runMultipleTests(t, myTests)
}
