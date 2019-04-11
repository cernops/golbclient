package ci

import (
	"testing"
)

func TestLemonLoad(t *testing.T) {

	myTests := []lbTest{
		lbTest{title: "LemonLoadSingle", configuration: "../test/lbclient_lemon_load_single.conf", shouldWork: true, metricValue: 1},
		lbTest{title: "LemonLoad", configuration: "../test/lbclient_lemon_load.conf", shouldWork: true, metricValue: 27},
		lbTest{title: "LemonFailed", configuration: "../test/lbclient_lemon_check_fail.conf", shouldWork: true, metricValue: -12}}

	runMultipleTests(t, false, myTests)
}
