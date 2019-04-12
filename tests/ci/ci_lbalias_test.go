package ci

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"

	"gitlab.cern.ch/lb-experts/golbclient/helpers/appSettings"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/mapping"

	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig"
)

var options appSettings.Options

// TestMissingLBClientFile : test if an error is given when a non-existent configuration file is given
func TestMissingLBClientFile(t *testing.T) {
	logger.SetLevel(logger.ERROR)
	options.LbAliasFile = "..//path/to/nonexistent/file"
	_, err := mapping.ReadLBConfigFiles(options)

	assert.NotNil(t, err,
		"No error was detected when attempting to access a non-existent configuration file. Failing test...")
}

// TestSingleLBAliasFile : attempts to read a single lbalias definitions from a given configuration file
func TestSingleLbAliasConfigFile(t *testing.T) {
	logger.SetLevel(logger.ERROR)
	options.LbAliasFile = "../test/conf/lbaliases.single"
	options.LbMetricConfDir = "../test/conf/"
	options.LbMetricDefaultFileName = "lbclient.single.conf"
	lbAliasesMappings, err := mapping.ReadLBConfigFiles(options)

	assert.Nil(t, err, "Failed to access the configuration file [%s]", options.LbAliasFile)
	logger.Debug("Read the aliases [%v] from the configuration file [%s]", lbAliasesMappings, options.LbAliasFile)
	assert.Equal(t, 1, len(lbAliasesMappings),
		"Found [%d] instead of [1] lbalias mapping entry definitions in the configuration.", len(lbAliasesMappings))
}

// TestMissingConfigurationFile : attempts to run the application with a non-existent configuration file
func TestMissingConfigurationFile(t *testing.T) {
	logger.SetLevel(logger.FATAL)
	cfg := mapping.NewConfiguration("../test/lbtest.conf_does_not_exist", "myTest")
	err := lbconfig.Evaluate(cfg, defaultTimeout, false)

	assert.NotNil(t, err,
		"Expected an error when attempting to read the non-existent file [%s]", cfg.ConfigFilePath)
}

func TestOneConfigFileMultipleAliases(t *testing.T) {
	logger.SetLevel(logger.ERROR)
	options.LbAliasFile = "../test/conf_returnValue/lbaliases.single"
	options.LbMetricConfDir = "../test/conf_returnValue/"
	options.LbMetricDefaultFileName = "lbclient.single.conf"
	lbAliasesMappings, err := mapping.ReadLBConfigFiles(options)
	
	assert.Nil(t, err, "Failed to access the configuration file [%s]", options.LbAliasFile)
	logger.Debug("Read the aliases [%v] from the configuration file [%s]", lbAliasesMappings, options.LbAliasFile)
	assert.Equal(t, 1, len(lbAliasesMappings),
		"Found [%d] instead of [1] lbalias mapping entry definitions in the configuration.", len(lbAliasesMappings))

	var appOutput bytes.Buffer
	appOutput.WriteString(lbAliasesMappings[0].String() + ",")
	err = lbconfig.Evaluate(lbAliasesMappings[0], defaultTimeout, false)

	assert.Nil(t, err,"We got an error evaluating the alias [%v]", err)

	metricType, metricValue := mapping.GetReturnCode(appOutput, lbAliasesMappings)
	logger.Trace("The return code is [%v] [%v]", metricType, metricValue)

	assert.Equal(t, "integer", metricType,
		"We were expecting to have an integer, and we got [%v] with value [%v]", metricType, metricValue)
}

func TestOneConfigFileMultipleAliasesString(t *testing.T) {
	logger.SetLevel(logger.ERROR)
	options.LbAliasFile = "../test/conf_returnString/lbaliases.single"
	options.LbMetricConfDir = "../test/conf_returnString/"
	options.LbMetricDefaultFileName = "lbclient.conf"
	lbAliasesMappings, err := mapping.ReadLBConfigFiles(options)

	assert.Nil(t, err, "Failed to access the configuration file [%s]", options.LbAliasFile)
	logger.Debug("Read the aliases [%v] from the configuration file [%s]", lbAliasesMappings, options.LbAliasFile)
	assert.Equal(t, 2, len(lbAliasesMappings),
		"Found [%d] instead of [2] lbalias mapping entry definitions in the configuration.", len(lbAliasesMappings))

	var appOutput bytes.Buffer
	for _, value := range lbAliasesMappings {
		appOutput.WriteString(value.String() + ",")
		err = lbconfig.Evaluate(value, defaultTimeout, false)

		assert.Nil(t, err,"We got an error evaluating the alias [%v]", err)
	}
	metricType, metricValue := mapping.GetReturnCode(appOutput, lbAliasesMappings)
	logger.Info("The return code is [%v] [%v]", metricType, metricValue)

	assert.Equal(t, "string", metricType,
		"We were expecting to have a string, and we got [%v] with value [%v]", metricType, metricValue)
}
