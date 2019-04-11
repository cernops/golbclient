package ci

import (
	"testing"
)

func TestCollectdLoad(t *testing.T) {
	myTests := []lbTest{
		lbTest{title: "CollectdLoad", configuration: "../test/lbclient_collectd_load_single.conf", expectedMetricValue: 98},
		lbTest{title: "LoadConfigurationFile", configuration: "../test/lbclient_collectd_load.conf", expectedMetricValue: 72},
		lbTest{title: "LoadFailedConfigurationFile", configuration: "../test/lbclient_collectd_load_fail.conf", shouldFail: true, expectedMetricValue: -15},
	}
	runMultipleTests(t, myTests)
}
