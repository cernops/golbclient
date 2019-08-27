package ci

import (
	"gitlab.cern.ch/lb-experts/golbclient/tests"
	"io/ioutil"
	"os"
	"testing"
)

var rogerTests []tests.LbTest

func init() {
	rogerTests = []tests.LbTest{
		{Title: "LemonLoadSingle",
			Configuration: "../test/lbclient_roger.conf", ExpectedMetricValue: 42,
			Setup: createRogerFileProduction},
		{Title: "LemonLoadSingle",
			Configuration: "../test/lbclient_roger.conf", ExpectedMetricValue: -13,
			Setup: createRogerFileDraining},
	}
}

func createRogerFile(t testing.TB, state string) {
	path := "/etc/roger/current.yaml"
	if _, err := os.Stat("/etc/roger"); os.IsNotExist(err) {
		if err = os.Mkdir("/etc/roger", os.ModePerm); err != nil {
			t.Log(err)
			t.Fail()
		}
	}

	if err := ioutil.WriteFile(path, []byte("---\nappstate: "+state+"\n"), 0755); err != nil {
		t.Log(err)
		t.Fail()
	}
}
func createRogerFileProduction(t testing.TB) {
	createRogerFile(t, "production")
}
func createRogerFileDraining(t testing.TB) {
	createRogerFile(t, "draining")
}

func TestRoger(t *testing.T) {
	tests.RunMultipleTests(t, rogerTests)
}

func BenchmarkRoger(b *testing.B) {
	tests.RunMultipleTests(b, rogerTests)
}
