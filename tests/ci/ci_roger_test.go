package ci

import (
	"io/ioutil"
	"os"
	"testing"

	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/mapping"
)

func createRogerFile(state string) error {
	path := "/etc/roger/current.yaml"
	if _, err := os.Stat("/etc/roger"); os.IsNotExist(err) {
		err = os.Mkdir("/etc/roger", os.ModePerm)
		if err != nil {
			return err
		}
	}
	err := ioutil.WriteFile(path, []byte("---\nappstate: "+state+"\n"), 0755)
	if err != nil {
		return err
	}
	return nil
}
func TestRogerFunctionality(t *testing.T) {
	logger.SetLevel(logger.TRACE)

	err := createRogerFile("production")
	if err != nil {
		t.Errorf("Error creating the file %v", err)
		t.FailNow()
	}
	cfg := mapping.NewConfiguration("../test/lbclient_roger.conf")
	err = lbconfig.Evaluate(cfg, defaultTimeout)
	if err != nil {
		logger.Error("Detected an error when attempting to evaluate the alias [%s], Error [%s]", cfg.ConfigFilePath, err.Error())
		t.Fail()
	}
	if cfg.MetricValue != 42 {
		logger.Error("The expected metric value was [42] but got [%d] instead. Failing the test...", cfg.MetricValue)
		t.Fail()
	}
}

func TestRogerFailedFunctionality(t *testing.T) {
	logger.SetLevel(logger.TRACE)

	err := createRogerFile("draining")
	if err != nil {
		t.Errorf("Error creating the file %v", err)
		t.FailNow()
	}

	cfg := mapping.NewConfiguration("../test/lbclient_roger.conf")
	err = lbconfig.Evaluate(cfg, defaultTimeout)
	if err == nil {
		logger.Error("There was no error with this configuration. We were expecting a 'roger error'")
		t.Fail()
	}
	if cfg.MetricValue != -13 {
		logger.Error("The expected metric value was [-1] but got [%d] instead. Failing the test...", cfg.MetricValue)
		t.Fail()
	}
}
