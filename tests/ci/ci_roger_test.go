package ci

import (
	"io/ioutil"
	"os"
	"testing"
)

func createRogerFile(t *testing.T, state string) {
	path := "/etc/roger/current.yaml"
	if _, err := os.Stat("/etc/roger"); os.IsNotExist(err) {
		if err = os.Mkdir("/etc/roger", os.ModePerm); err != nil {
			t.Fatal(err)
		}
	}

	if err := ioutil.WriteFile(path, []byte("---\nappstate: "+state+"\n"), 0755); err != nil {
		t.Fatal(err)
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
		{title: "LemonLoadSingle",
			configuration: "../test/lbclient_roger.conf", expectedMetricValue: 42,
			setup: createRogerFileProduction},
		{title: "LemonLoadSingle",
			configuration: "../test/lbclient_roger.conf", expectedMetricValue: -13,
			setup: createRogerFileDraining},
	}

	runMultipleTests(t, myTests)
}
