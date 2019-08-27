package ci

import (
	"gitlab.cern.ch/lb-experts/golbclient/tests"
	"testing"
)

var daemonTests []tests.LbTest

func init() {
	daemonTests = []tests.LbTest{
		{Title: "DaemonLegacyTest",
			Configuration: "../test/daemon/lbclient_daemon_legacy_check.conf",
			ExpectedMetricValue: 5},
		{Title: "DaemonNewSyntaxTest",
			Configuration: "../test/daemon/lbclient_daemon_check.conf",
			ExpectedMetricValue: 5},
		{Title: "DaemonFailPart1",
			ConfigurationContent: "check daemon",
			ShouldFail: true,
			ExpectedMetricValue: -17},
		{Title: "DaemonFailPart2",
			ConfigurationContent: "check daemon {}",
			ShouldFail: true,
			ExpectedMetricValue: -17},
		{Title: "DaemonFailPart3",
			ConfigurationContent: `check daemon {"protocol":"tcp"}`,
			ShouldFail: true,
			ExpectedMetricValue: -17},
		{Title: "DaemonFailPart4",
			ConfigurationContent: `check daemon {"part": 24, "protocol":"tcp"}`,
			ShouldFail: true,
			ExpectedMetricValue: -17},
		{Title: "DaemonFailPart5",
			ConfigurationContent: `check daemon {"port":22, {}}`,
			ShouldFail: true,
			ExpectedMetricValue: -17},
		{Title: "DaemonFailPart6",
			ConfigurationContent: `check daemon {"port":-1}`,
			ShouldFail: true,
			ExpectedMetricValue: -17},
		{Title: "DaemonFailPart7",
			ConfigurationContent: `check daemon {"port":22, "ip":[-1,"a"]}`,
			ShouldFail: true,
			ExpectedMetricValue: -17},
		{Title: "DaemonFailPart8",
			ConfigurationContent: `check daemon {"port":22, "protocol":0}`,
			ShouldFail: true,
			ExpectedMetricValue: -17},
		{Title: "DaemonFailPart9",
			ConfigurationContent: `check daemon {"port":22, "protocol":{"key":"value"}}`,
			ShouldFail: true,
			ExpectedMetricValue: -17},
		{Title: "DaemonWarningPart1",
			ConfigurationContent: "check daemon {\"port\":22, \"another_argument\":22}\nload constant 44",
			ExpectedMetricValue: 44},
		{Title: "DaemonWarningPart2",
			ConfigurationContent: "check daemon {\"port\":22, \"port\":22}\nload constant 45",
			ExpectedMetricValue: 45},

		// More tests
		{Title: "UDPTest",
			ConfigurationContent: "check daemon {\"port\":922, \"protocol\":\"udp\"}\nload constant 32",
			ExpectedMetricValue: 32},
		{Title: "UDPTestOnTCP",
			ConfigurationContent: "check daemon {\"port\":22, \"protocol\":\"udp\"}\nload constant 32",
			ExpectedMetricValue: -17},
		{Title: "UDPTestOnTCPorUDP",
			ConfigurationContent: "check daemon {\"port\":202, \"protocol\":[\"udp\",\"tcp\"]}\nload constant 33",
			ExpectedMetricValue: 33},
		{Title: "UDPTestOnTCPorUDP",
			ConfigurationContent: "check daemon {\"port\":[22,922], \"protocol\":\"udp\"}\nload constant 33",
			ExpectedMetricValue: 33},
		{Title: "UDPTestIpv4",
			ConfigurationContent: "check daemon {\"port\":922, \"protocol\":\"udp\", \"ip\":\"ipv4\"}\nload constant 37",
			ExpectedMetricValue: 37},

		// Checking localhost
		{Title: "Localhost_202_IPv4",
			ConfigurationContent: "check daemon {\"port\":202, \"host\":\"127.0.0.1\", \"ip\":\"ipv4\"}\nload constant 37", ExpectedMetricValue: 37},
		{Title: "Localhost_202_IPv6_fail",

			ConfigurationContent: "check daemon {\"port\":202, \"host\":\"127.0.0.1\", \"ip\":\"ipv6\"}\nload constant 37",
			ExpectedMetricValue: -17},
		{Title: "Localhost_202_IPv4_fail",
			ConfigurationContent: "check daemon {\"port\":202, \"host\":\"::1\", \"ip\":\"ipv4\"}\nload constant 37",
			ExpectedMetricValue: -17},
		/*
				INC2031907

			{Title: "Localhost_202_IPv6",
				ConfigurationContent: "check daemon {\"port\":202, \"host\":\"::1\", \"ip\":\"ipv6\"}\nload constant 37", ExpectedMetricValue: 37},
			{Title: "Localhost_202_Ipv4or6",
				ConfigurationContent: "check daemon {\"port\":202, \"host\":\"::1\", \"ip\":[\"4\", \"ipv6\"]}\nload constant 37", ExpectedMetricValue: 37},
		*/
		{Title: "Localhost_202_Host_fail",
			ConfigurationContent: "check daemon {\"port\":202, \"host\":\"127.0.0\"}\nload constant 37",
			ShouldFail: true,
			ExpectedMetricValue: -17},
		{Title: "Localhost_202_Host_not_found",
			ConfigurationContent: "check daemon {\"port\":202, \"host\":\"129.0.0.1\"}\nload constant 37",
			ExpectedMetricValue: -17},
	}
}

func TestDaemonCheck(t *testing.T) {
	tests.RunMultipleTests(t, daemonTests)
}

func BenchmarkDaemonCheck(b *testing.B) {
	tests.RunMultipleTests(b, daemonTests)
}