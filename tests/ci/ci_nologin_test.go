package ci

import (
	"io/ioutil"
	"os"
	"testing"

	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/mapping"
)

func TestNologinFunctionality(t *testing.T) {
	logger.SetLevel(logger.ERROR)
	cfg := mapping.NewConfiguration("../test/lbclient_nologin.conf", "Check that the nologin works")
	err := lbconfig.Evaluate(cfg, defaultTimeout)
	if err != nil {
		logger.Error("Detected an error when attempting to evaluate the alias [%s], Error [%s]", cfg.ConfigFilePath, err.Error())
		t.Fail()
	}
	if cfg.MetricValue != 5 {
		logger.Error("The expected metric value was [5] but got [%d] instead. Failing the test...", cfg.MetricValue)
		t.Fail()
	}
}

func TestNologinFailedFunctionality(t *testing.T) {
	logger.SetLevel(logger.ERROR)

	path := "/etc/nologin"
	err := ioutil.WriteFile(path, []byte("Hello"), 0755)
	if err != nil {
		t.Errorf("Unable to write file: %v", err)
		t.FailNow()
	}

	defer func() {
		err := os.Remove(path)
		if err != nil {
			t.Errorf("Failed to remove the file %v", err)
			t.FailNow()
		}
	}()
	cfg := mapping.NewConfiguration("../test/lbclient_nologin.conf", "blabla") //Check that nologin fails if the file exists")
	err = lbconfig.Evaluate(cfg, defaultTimeout)
	if err == nil {
		logger.Error("There was no error with this configuration. We were expecting a 'nologing error'")
		t.Fail()
	}
	if cfg.MetricValue != -1 {
		logger.Error("The expected metric value was [-1] but got [%d] instead. Failing the test...", cfg.MetricValue)
		t.Fail()
	}
}
