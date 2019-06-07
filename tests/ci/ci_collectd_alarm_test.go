package ci

import (
	"testing"
)

func TestCollectdAlarms(t *testing.T) {
	myTests := []lbTest{
		{title: "CollectdAlarmFunctionality",
			configuration: "../test/lbclient_collectd_alarm_check_single.conf", expectedMetricValue: 2},
	}

	runMultipleTests(t, myTests)
}
