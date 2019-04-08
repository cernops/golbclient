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
	logger.SetLevel(logger.TRACE)

	var myTests [2]lbTest
	myTests[0] = lbTest{"noLoginWorks", "../test/lbclient_nologin.conf", true, 5, nil, nil}
	myTests[1] = lbTest{"noLoginFails", "../test/lbclient_nologin.conf", false, -1, createNoLogin, removeNoLogin}
	for _, myTest := range myTests {
		t.Run(myTest.title, func(t *testing.T) {
			RunEvaluate(t, myTest.configuration, myTest.shouldWork, myTest.metricValue, myTest.setup, myTest.cleanup)
		})
	}
}
