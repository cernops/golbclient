package ci

import (
	"io/ioutil"
	"os"
	"testing"

	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
)

func createNoLogin(t *testing.T) {
	path := "/etc/nologin"
	err := ioutil.WriteFile(path, []byte("Hello"), 0755)
	logger.Info("Creating the nologin file")
	if err != nil {
		t.Errorf("Unable to write file: %v", err)
		t.FailNow()
	}
}
func removeNoLogin(t *testing.T) {
	path := "/etc/nologin"
	err := os.Remove(path)
	logger.Info("Removing the file\n")
	if err != nil {
		t.Errorf("Failed to remove the file %v", err)
		t.FailNow()
	}
}

func TestNologin(t *testing.T) {
	myTests := []lbTest{
		lbTest{title: "noLoginWorks", configuration: "../test/lbclient_nologin.conf", expectedMetricValue: 5},
		lbTest{title: "noLoginFails", configuration: "../test/lbclient_nologin.conf", expectedMetricValue: -1, setup: createNoLogin, cleanup: removeNoLogin},
	}

	runMultipleTests(t, myTests)
}
