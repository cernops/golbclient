package ci

import (
	"gitlab.cern.ch/lb-experts/golbclient/lbalias"
	"gitlab.cern.ch/lb-experts/golbclient/utils"
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
	"testing"
)

var options utils.Options

// TestMissingLBClientFile : test if an error is given when a non-existent configuration file is given
func TestMissingLBClientFile(t *testing.T) {
	logger.SetLevel(logger.ERROR)
	options.LbAliasFile = "..//path/to/nonexistent/file"
	_, err := utils.ReadLBAliases(options)

	if err == nil {
		logger.Error("No error was detected when attempting to access a non-existent configuration file. Failing test...")
		t.Fail()
	}
}

// TestSingleLBAliasFile : attempts to read a single lbalias definitions from a given configuration file
func TestSingleLBAliasFile(t *testing.T) {
	logger.SetLevel(logger.ERROR)
	options.LbAliasFile = "../test/lbalias_single.conf"
	lbAliases, err := utils.ReadLBAliases(options)
	if err != nil {
		logger.Error("Failed to access the configuration file [%s]", options.LbAliasFile)
		t.Fail()
	}
	logger.Debug("Read the aliases [%v] from the configuration file [%s]", lbAliases, options.LbAliasFile)
	if len(lbAliases) != 1 {
		logger.Error("Found [%d] instead of [1] lbalias entry definitions in the configuration.")
		t.Fail()
	}
}

// TestDoubleLBAliasFile : attempts to read two lbalias entry definitions from a given configuration file
func TestDoubleLBAliasFile(t *testing.T) {
	logger.SetLevel(logger.ERROR)
	options.LbAliasFile = "../test/lbalias_double.conf"
	lbAliases, err := utils.ReadLBAliases(options)
	if err != nil {
		logger.Error("Failed to access the configuration file [%s]", options.LbAliasFile)
		t.Fail()
	}
	logger.Debug("Read the aliases [%v] from the configuration file [%s]", lbAliases, options.LbAliasFile)
	if len(lbAliases) != 2 {
		logger.Error("Found [%d] instead of [2] lbalias entry definitions in the configuration.")
		t.Fail()
	}
}

// TestMissingConfigurationFile : attempts to run the application with a non-existent configuration file
func TestMissingConfigurationFile(t *testing.T) {
	logger.SetLevel(logger.FATAL)
	lba := lbalias.LBalias{Name: "myTest",
		Syslog:     true,
		ConfigFile: "../test/lbtest.conf_does_not_exist"}

	err := lba.Evaluate()
	if err == nil {
		logger.Error("Expected an error when attempting to read the non-existent file [%s]", lba.ConfigFile)
		t.Fail()
	}
}

/*
//Let's check that, given a constant load, we get that number
func ExampleConstantLoad() {
	lbalias := checks.LBalias{Name: "myTest",
		ConfigFile: "test/lbclient_constant.conf"}

	lbalias.Evaluate()
	// Output: constant
	// [add_constant] value= 249
}
*/
