package ci

import (
	"gitlab.cern.ch/lb-experts/golbclient/tests"
	"testing"
)

var collectdAlarmTests []tests.LbTest

func init() {
	collectdAlarmTests = []tests.LbTest{
		{Title: "CollectdAlarmFunctionality",
			Configuration: "../test/lbclient_collectd_alarm_check_single.conf", ExpectedMetricValue: 2},
		{Title: "CollectdAlarmFunctionalityMultipleAndInverted",
			Configuration: "../test/lbclient_collectd_alarm_check_multiple.conf", ExpectedMetricValue: 3},
		{Title: "CollectdAlarmFunctionalityMultipleAllStates",
			Configuration: "../test/lbclient_collectd_alarm_check_all_states.conf", ExpectedMetricValue: 4},
		{Title: "CollectdAlarmFunctionalityFailSingle",
			Configuration: "../test/lbclient_collectd_alarm_check_fail_single.conf", ExpectedMetricValue: -15, ShouldFail: true},
		{Title: "CollectdAlarmFunctionalityFailMultiple",
			Configuration: "../test/lbclient_collectd_alarm_check_fail_multiple.conf", ExpectedMetricValue: -15, ShouldFail: true},
	}
}

func TestCollectdAlarms(t *testing.T) {
	tests.RunMultipleTests(t, collectdAlarmTests)
}

func BenchmarkCollectdAlarms(b *testing.B) {
	tests.RunMultipleTests(b, collectdAlarmTests)
}
