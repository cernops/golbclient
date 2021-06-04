package ci

import (
	"testing"

	"gitlab.cern.ch/lb-experts/golbclient/lbconfig"
)


func TestPostToErmis(t *testing.T) {
	type test struct {
		caseID           int
		Status           string
		LBPostFile       string
		ExpectedHTTPCode int
	}
	var l lbconfig.AppLauncher

	testCases := []test{
		{caseID: 1,
			Status:           "test1.cern.ch=4",
			LBPostFile:       "../test/post_conf/lbpost-single.yaml",
			ExpectedHTTPCode: 200,
		},
		{caseID: 2,
			Status:           "test1.cern.ch=-8,test2.cern.ch=9",
			LBPostFile:       "../test/post_conf/lbpost-multi.yaml",
			ExpectedHTTPCode: 200,
		},
	}

	for _, tc := range testCases {
		l.PostErmis = tc.Status
		output, err := l.PostToErmis(tc.LBPostFile)
		if output != tc.ExpectedHTTPCode {
			t.Errorf("failed in TestPostToErmis for caseID %v\n, error:%v", tc.caseID, err)
		}
	}

}
