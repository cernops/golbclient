package ci

import (
	"testing"
	"time"
)

func TestTimeOut(t *testing.T) {
	myTests := []lbTest{
		{title: "TimeoutNotKillExecution", configuration: "../test/lbclient_timeout_6s.conf", expectedMetricValue: 15},
		{title: "TimeoutKillExecution", configuration: "../test/lbclient_timeout_6s.conf", expectedMetricValue: -14, shouldFail: true, timeout: time.Second * 3},
	}

	runMultipleTests(t, myTests)
}
