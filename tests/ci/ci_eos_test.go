package ci

import (
	"testing"
)

func TestEOS(t *testing.T) {
	myTests := []eosTest{
		{title: "EosOk",
			procmounts:          "../test/procmounts_OK",
			expectedMetricValue: 1},
		{title: "EosErrorENOTCONN",
			procmounts:          "../test/procmounts_ENOTCONN",
			expectedMetricValue: -1},
		{title: "EosErrorEOPNOTSUPP",
			procmounts:          "../test/procmounts_EOPNOTSUPP",
			expectedMetricValue: -1},
		{title: "EosErrorEIOERR",
			procmounts:          "../test/procmounts_EIOERR",
			expectedMetricValue: -1},
		{title: "EosErrorOther",
			procmounts:          "../test/procmounts_other_error",
			expectedMetricValue: 1},
	}

	runMultipleEosTests(t, myTests)
}
