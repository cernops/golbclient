package ci

import (
	"io/ioutil"
	"os"
	"testing"

	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
)

func createRogerFile(t *testing.T, state string) {
	path := "/etc/roger/current.yaml"
	if _, err := os.Stat("/etc/roger"); os.IsNotExist(err) {
		err = os.Mkdir("/etc/roger", os.ModePerm)
		if err != nil {
			t.FailNow()
		}
	}
	err := ioutil.WriteFile(path, []byte("---\nappstate: "+state+"\n"), 0755)
	if err != nil {
		t.FailNow()
	}
}

func TestRogerFunctionality(t *testing.T) {
	logger.SetLevel(logger.ERROR)

	createRogerFile(t, "production")
	RunEvaluate(t, "../test/lbclient_roger.conf", true, 42, nil, nil)
}

func TestRogerFailedFunctionality(t *testing.T) {
	logger.SetLevel(logger.FATAL)

	createRogerFile(t, "draining")
	RunEvaluate(t, "../test/lbclient_roger.conf", false, -13, nil, nil)
}
