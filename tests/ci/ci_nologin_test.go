package ci

import (
	"io/ioutil"
	"os"
	"testing"

	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/mapping"
)

func RunEvaluate(t *testing.T, configFile string, shouldWork bool, metricValue int) {
	cfg := mapping.NewConfiguration(configFile)
	err := lbconfig.Evaluate(cfg, defaultTimeout)
	if shouldWork == true {
		if err != nil {
			logger.Error("Got back an error, and we  were not expecting that. Error [%s] ", err.Error())
			t.FailNow()
		}
	} else {
		if err == nil {
			logger.Error("The evaluation of the alias was supposed to fail")
			t.FailNow()
		}
	}
	if cfg.MetricValue != metricValue {
		logger.Error("We were expecting the value %i, and got %i", cfg.MetricValue, metricValue)
		t.FailNow()
	}

}

func TestNologin(t *testing.T) {
	logger.SetLevel(logger.ERROR)
	t.Run("nologinWorks", nologinWorks)
	t.Run("nologinFails", nologinFails)
}

func nologinWorks(t *testing.T) {
	RunEvaluate(t, "../test/lbclient_nologin.conf", true, 5)
}

func nologinFails(t *testing.T) {
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

	RunEvaluate(t, "../test/lbclient_nologin.conf", true, -1)
}
