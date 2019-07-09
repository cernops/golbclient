package ci

import (
	"testing"
)

func TestCollectdAlarms(t *testing.T) {
	myTests := []lbTest{
		{title: "CollectdAlarmFunctionality",
			configuration: "../test/lbclient_collectd_alarm_check_single.conf", expectedMetricValue: 2},
		{title: "CollectdAlarmFunctionalityMultipleAndInverted",
			configuration: "../test/lbclient_collectd_alarm_check_multiple.conf", expectedMetricValue: 3},
		{title: "CollectdAlarmFunctionalityMultipleAllStates",
			configuration: "../test/lbclient_collectd_alarm_check_all_states.conf", expectedMetricValue: 4},
		{title: "CollectdAlarmFunctionalityFailSingle",
			configuration: "../test/lbclient_collectd_alarm_check_fail_single.conf", expectedMetricValue: -15, shouldFail: true},
		{title: "CollectdAlarmFunctionalityFailMultiple",
			configuration: "../test/lbclient_collectd_alarm_check_fail_multiple.conf", expectedMetricValue: -15, shouldFail: true},
	}

	runMultipleTests(t, myTests)
}
