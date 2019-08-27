package ci

import (
	"gitlab.cern.ch/lb-experts/golbclient/tests"
	"testing"
	"time"
)

var timeoutTests []tests.LbTest

func init() {
	timeoutTests = []tests.LbTest{
		{Title: "TimeoutNotKillExecution", Configuration: "../test/lbclient_timeout_6s.conf", ExpectedMetricValue: 15},
		{Title: "TimeoutKillExecution", Configuration: "../test/lbclient_timeout_6s.conf", ExpectedMetricValue: -14, ShouldFail: true, Timeout: time.Second * 3},
	}
}

func TestTimeOut(t *testing.T) {
	tests.RunMultipleTests(t, timeoutTests)
}

func BenchmarkTimeOut(b *testing.B) {
	tests.RunMultipleTests(b, timeoutTests)
}
