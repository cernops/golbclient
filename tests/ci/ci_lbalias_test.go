package ci

import (
	"bytes"
	"testing"

	"gitlab.cern.ch/lb-experts/golbclient/lbalias"
	"gitlab.cern.ch/lb-experts/golbclient/utils"
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
)

var options utils.Options

// TestMissingLBClientFile : test if an error is given when a non-existent configuration file is given
func TestMissingLBClientFile(t *testing.T) {
	logger.SetLevel(logger.ERROR)
	options.LbAliasFile = "..//path/to/nonexistent/file"
	_, err := utils.ReadLBConfigFiles(options)

	if err == nil {
		logger.Error("No error was detected when attempting to access a non-existent configuration file. Failing test...")
		t.Fail()
	}
}

// TestSingleLBAliasFile : attempts to read a single lbalias definitions from a given configuration file
func TestSingleLbAliasConfigFile(t *testing.T) {
	logger.SetLevel(logger.ERROR)
	options.LbAliasFile = "../test/conf/lbaliases.single"
	options.LbMetricConfDir = "../test/conf/"
	options.LbMetricDefaultFileName = "lbclient.single.conf"
	lbAliasesMappings, err := utils.ReadLBConfigFiles(options)
	if err != nil {
		logger.Error("Failed to access the configuration file [%s]", options.LbAliasFile)
		t.Fail()
	}
	logger.Debug("Read the aliases [%v] from the configuration file [%s]", lbAliasesMappings, options.LbAliasFile)
	if len(lbAliasesMappings) != 1 {
		logger.Error("Found [%d] instead of [1] lbalias mapping entry definitions in the configuration.", len(lbAliasesMappings))
		t.Fail()
	}
}

// TestMissingConfigurationFile : attempts to run the application with a non-existent configuration file
func TestMissingConfigurationFile(t *testing.T) {
	logger.SetLevel(logger.FATAL)
	cfg := utils.NewConfiguration("../test/lbtest.conf_does_not_exist", "myTest")

	err := lbalias.Evaluate(cfg)
	if err == nil {
		logger.Error("Expected an error when attempting to read the non-existent file [%s]", cfg.ConfigFilePath)
		t.Fail()
	}
}

func TestOneConfigFileMultipleAliases(t *testing.T) {
	logger.SetLevel(logger.TRACE)
	options.LbAliasFile = "../test/conf_returnValue/lbaliases.single"
	options.LbMetricConfDir = "../test/conf_returnValue/"
	options.LbMetricDefaultFileName = "lbclient.single.conf"
	lbAliasesMappings, err := utils.ReadLBConfigFiles(options)
	if err != nil {
		logger.Error("Failed to access the configuration file [%s]", options.LbAliasFile)
		t.Fail()
	}

	logger.Debug("Read the aliases [%v] from the configuration file [%s]", lbAliasesMappings, options.LbAliasFile)
	if len(lbAliasesMappings) != 1 {
		logger.Error("Found [%d] instead of [1] lbalias mapping entry definitions in the configuration.", len(lbAliasesMappings))
		t.Fail()
	}
	var appOutput bytes.Buffer
	err = lbalias.Evaluate(lbAliasesMappings[0])
	appOutput.WriteString(lbAliasesMappings[0].String() + ",")
	if err != nil {
		logger.Error("We got an error evaluating the alias [%v]", err)
		t.Fail()
	}
	metricType, metricValue := utils.GetReturnCode(appOutput, lbAliasesMappings)
	logger.Info("The return code is [%v] [%v]", metricType, metricValue)
	if metricType != "integer" {
		logger.Error("We were expecting to have an integer, and we got [%v] with value [%v]", metricType, metricValue)
		t.Fail()
	}
}

func TestOneConfigFileMultipleAliasesString(t *testing.T) {
	logger.SetLevel(logger.ERROR)
	options.LbAliasFile = "../test/conf_returnString/lbaliases.single"
	options.LbMetricConfDir = "../test/conf_returnString/"
	options.LbMetricDefaultFileName = "lbclient.conf"
	lbAliasesMappings, err := utils.ReadLBConfigFiles(options)
	if err != nil {
		logger.Error("Failed to access the configuration file [%s]", options.LbAliasFile)
		t.Fail()
	}

	logger.Debug("Read the aliases [%v] from the configuration file [%s]", lbAliasesMappings, options.LbAliasFile)
	if len(lbAliasesMappings) != 2 {
		logger.Error("Found [%d] instead of [2] lbalias mapping entry definitions in the configuration.", len(lbAliasesMappings))
		t.Fail()
	}
	var appOutput bytes.Buffer
	for _, value := range lbAliasesMappings {
		err = lbalias.Evaluate(value)
		appOutput.WriteString(value.String() + ",")
		if err != nil {
			logger.Error("We got an error evaluating the alias [%v]", err)
			t.Fail()
		}
	}
	metricType, metricValue := utils.GetReturnCode(appOutput, lbAliasesMappings)
	logger.Info("The return code is [%v] [%v]", metricType, metricValue)

	if metricType != "string" {
		logger.Error("We were expecting to have a string, and we got [%v] with value [%v]", metricType, metricValue)
		t.Fail()
	}

}
