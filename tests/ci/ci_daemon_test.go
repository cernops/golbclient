package ci

import (
	"testing"
)

func TestDaemonCheck(t *testing.T) {
	myTests := []lbTest{
		{title: "DaemonLegacyTest",
			configuration:       "../test/daemon/lbclient_daemon_legacy_check.conf",
			expectedMetricValue: 5},
		{title: "DaemonNewSyntaxTest",
			configuration:       "../test/daemon/lbclient_daemon_check.conf",
			expectedMetricValue: 5},
		{title: "DaemonFailPart1",
			configurationContent: "check daemon",
			shouldFail:           true,
			expectedMetricValue:  -17},
		{title: "DaemonFailPart2",
			configurationContent: "check daemon {}",
			shouldFail:           true,
			expectedMetricValue:  -17},
		{title: "DaemonFailPart3",
			configurationContent: `check daemon {"protocol":"tcp"}`,
			shouldFail:           true,
			expectedMetricValue:  -17},
		{title: "DaemonFailPart4",
			configurationContent: `check daemon {"part": 24, "protocol":"tcp"}`,
			shouldFail:           true,
			expectedMetricValue:  -17},
		{title: "DaemonFailPart5",
			configurationContent: `check daemon {"port":22, {}}`,
			shouldFail:           true,
			expectedMetricValue:  -17},
		{title: "DaemonFailPart6",
			configurationContent: `check daemon {"port":-1}`,
			shouldFail:           true,
			expectedMetricValue:  -17},
		{title: "DaemonFailPart7",
			configurationContent: `check daemon {"port":22, "ip":[-1,"a"]}`,
			shouldFail:           true,
			expectedMetricValue:  -17},
		{title: "DaemonFailPart8",
			configurationContent: `check daemon {"port":22, "protocol":0}`,
			shouldFail:           true,
			expectedMetricValue:  -17},
		{title: "DaemonFailPart9",
			configurationContent: `check daemon {"port":22, "protocol":{"key":"value"}}`,
			shouldFail:           true,
			expectedMetricValue:  -17},
		{title: "DaemonWarningPart1",
			configurationContent: "check daemon {\"port\":22, \"another_argument\":22}\nload constant 44",
			expectedMetricValue:  44},
		{title: "DaemonWarningPart2",
			configurationContent: "check daemon {\"port\":22, \"port\":22}\nload constant 45",
			expectedMetricValue:  45},

		// More tests
		{title: "UDPTest",
			configurationContent: "check daemon {\"port\":922, \"protocol\":\"udp\"}\nload constant 32",
			expectedMetricValue:  32},
		{title: "UDPTestOnTCP",
			configurationContent: "check daemon {\"port\":22, \"protocol\":\"udp\"}\nload constant 32",
			expectedMetricValue:  -17},
		{title: "UDPTestOnTCPorUDP",
			configurationContent: "check daemon {\"port\":202, \"protocol\":[\"udp\",\"tcp\"]}\nload constant 33",
			expectedMetricValue:  33},
		{title: "UDPTestOnTCPorUDP",
			configurationContent: "check daemon {\"port\":[22,922], \"protocol\":\"udp\"}\nload constant 33",
			expectedMetricValue:  33},
		{title: "UDPTestIpv4",
			configurationContent: "check daemon {\"port\":922, \"protocol\":\"udp\", \"ip\":\"ipv4\"}\nload constant 37",
			expectedMetricValue:  37},

		// Checking localhost
		{title: "Localhost_202_IPv4",
			configurationContent: "check daemon {\"port\":202, \"host\":\"127.0.0.1\", \"ip\":\"ipv4\"}\nload constant 37", expectedMetricValue: 37},
		{title: "Localhost_202_IPv6_fail",

			configurationContent: "check daemon {\"port\":202, \"host\":\"127.0.0.1\", \"ip\":\"ipv6\"}\nload constant 37",
			expectedMetricValue:  -17},
		{title: "Localhost_202_IPv4_fail",
			configurationContent: "check daemon {\"port\":202, \"host\":\"::1\", \"ip\":\"ipv4\"}\nload constant 37",
			expectedMetricValue:  -17},
		/*
				INC2031907

			{title: "Localhost_202_IPv6",
				configurationContent: "check daemon {\"port\":202, \"host\":\"::1\", \"ip\":\"ipv6\"}\nload constant 37", expectedMetricValue: 37},
			{title: "Localhost_202_Ipv4or6",
				configurationContent: "check daemon {\"port\":202, \"host\":\"::1\", \"ip\":[\"4\", \"ipv6\"]}\nload constant 37", expectedMetricValue: 37},
		*/
		{title: "Localhost_202_Host_fail",
			configurationContent: "check daemon {\"port\":202, \"host\":\"127.0.0\"}\nload constant 37",
			shouldFail:           true,
			expectedMetricValue:  -17},
		{title: "Localhost_202_Host_not_found",
			configurationContent: "check daemon {\"port\":202, \"host\":\"129.0.0.1\"}\nload constant 37",
			expectedMetricValue:  -17},
	}

	runMultipleTests(t, myTests)
}
