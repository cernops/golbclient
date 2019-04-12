package ci

import (
	"io/ioutil"
	"os"
	"testing"

	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
)

func createNoLogin(t *testing.T) {
	path := "/etc/nologin"
	logger.Debug("Creating the [nologin] file [%s]", path)

	if err := ioutil.WriteFile(path, []byte("Hello"), 0755); err != nil {
		t.Fatalf("Unable to write file: %v", err)
	}
}
func removeNoLogin(t *testing.T) {
	path := "/etc/nologin"
	logger.Debug("Removing the [nologin] file [%s]", path)

	if err := os.Remove(path); err != nil {
		t.Fatalf("Failed to remove the file %v", err)
	}
}

func TestNologin(t *testing.T) {
	myTests := []lbTest{
		{title: "noLoginWorks",
			configuration: "../test/lbclient_nologin.conf", expectedMetricValue: 5},
		{title: "noLoginFails",
			configuration: "../test/lbclient_nologin.conf", expectedMetricValue: -1,
			setup: createNoLogin, cleanup: removeNoLogin},
	}

	runMultipleTests(t, myTests)
}
