package ci

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"

	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
)

func createNoLogin(t *testing.T) {
	path := "/etc/nologin"
	logger.Debug("Attempting to create the [nologin] file [%s]...", path)
	err := ioutil.WriteFile(path, []byte("Hello"), 0755)

	assert.Nil(t, err, "Unable to write file: %v", err)
}
func removeNoLogin(t *testing.T) {
	path := "/etc/nologin"
	logger.Debug("Attempting to remove the [nologin] file [%s]...", path)
	err := os.Remove(path)

	assert.Nil(t, err, "Failed to remove the file %v", err)
}

func TestNoLogin(t *testing.T) {
	myTests := []lbTest{
		{title: "noLoginWorks",
			configuration: "../test/lbclient_nologin.conf", expectedMetricValue: 5},
		{title: "noLoginFails",
			configuration: "../test/lbclient_nologin.conf", expectedMetricValue: -1,
			setup: createNoLogin, cleanup: removeNoLogin},
	}

	runMultipleTests(t, myTests)
}
