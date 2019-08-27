package ci

import (
	"gitlab.cern.ch/lb-experts/golbclient/tests"
	"io/ioutil"
	"os"
	"testing"

	logger "github.com/sirupsen/logrus"
)

var noLoginTests []tests.LbTest

func init() {
	noLoginTests = []tests.LbTest{
		{Title: "noLoginWorks",
			Configuration: "../test/lbclient_nologin.conf", ExpectedMetricValue: 5},
		{Title: "noLoginFails",
			Configuration: "../test/lbclient_nologin.conf", ExpectedMetricValue: -1,
			Setup: createNoLogin, Cleanup: removeNoLogin},
	}
}

func createNoLogin(t testing.TB) {
	path := "/etc/nologin"
	logger.Debugf("Creating the [nologin] file [%s]", path)

	if err := ioutil.WriteFile(path, []byte("Hello"), 0755); err != nil {
		t.Logf("Unable to write file: %v", err)
		t.Fail()
	}
}

func removeNoLogin(t testing.TB) {
	path := "/etc/nologin"
	logger.Debugf("Removing the [nologin] file [%s]", path)

	if err := os.Remove(path); err != nil {
		t.Logf("Failed to remove the file %v", err)
		t.Fail()
	}
}

func TestNoLogin(t *testing.T) {
	tests.RunMultipleTests(t, noLoginTests)
}

func BenchmarkNoLogin(b *testing.B) {
	tests.RunMultipleTests(b, noLoginTests)
}
