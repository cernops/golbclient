package ci

import (
	"io/ioutil"
	"os"
	"testing"
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
func createRogerFileProduction(t *testing.T) {
	createRogerFile(t, "production")
}
func createRogerFileDraining(t *testing.T) {
	createRogerFile(t, "draining")
}

func TestRoger(t *testing.T) {
	myTests := []lbTest{
		lbTest{title: "LemonLoadSingle", configuration: "../test/lbclient_roger.conf", expectedMetricValue: 42, setup: createRogerFileProduction},
		lbTest{title: "LemonLoadSingle", configuration: "../test/lbclient_roger.conf", expectedMetricValue: -13, setup: createRogerFileDraining},
	}
	runMultipleTests(t, myTests)
}
