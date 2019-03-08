package main

import (
	"bytes"
	"fmt"
	"os"
	"sync"

	"gitlab.cern.ch/lb-experts/golbclient/helpers/appSettings"
	"gitlab.cern.ch/lb-experts/golbclient/lbalias/mapping"

	"github.com/jessevdk/go-flags"
	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
	"gitlab.cern.ch/lb-experts/golbclient/lbalias"
)

const (
	// OID : SNMP identifier
	OID = ".1.3.6.1.4.1.96.255.1"
	// Version number
	Version = "2.0"
	// Release number
	Release = "8"
)

// Flags API
var options appSettings.Options
var parser = flags.NewParser(&options, flags.Default)

// init : function responsible for the parsing of the application arguments
func init() {
	_, err := parser.Parse()
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}
	if options.Version {
		fmt.Printf("lbclient version %s.%s\n", Version, Release)
		os.Exit(0)
	}

	// Logger settings
	if err := applyLoggerSettings(); err != nil {
		logger.Fatal("A fatal error occurred when attempting to apply the given settings to the logger. "+
			"Error [%s]", err.Error())
		os.Exit(1)
	}

}

func main() {
	// Arguments parsed. Let's open the configuration file
	var lbConfMappings []*mapping.ConfigurationMapping
	lbConfMappings, err := mapping.ReadLBConfigFiles(options)
	if err != nil {
		logger.Error("An error occurred when attempting to process the configuration & aliases. Error [%s]",
			err.Error())
	}

	// Application output
	var appOutput bytes.Buffer

	// Concurrent evaluation
	var waitGroup sync.WaitGroup
	waitGroup.Add(len(lbConfMappings))

	// Evaluate for each of the configuration files found
	for _, value := range lbConfMappings {

		go func(confMapping *mapping.ConfigurationMapping) {
			defer waitGroup.Done()
			logger.Trace("Processing configuration file [%s] for aliases [%v]", confMapping.ConfigFilePath, confMapping.AliasNames)

			err := lbalias.Evaluate(confMapping)

			/* Abort if an error occurs */
			if err != nil {
				logger.Info("The evaluation of configuration file [%s] failed.", confMapping.ConfigFilePath)
			}
			appOutput.WriteString(confMapping.String() + ",")
		}(value)
	}

	// Wait for concurrent loop to finish before proceeding
	waitGroup.Wait()

	metricType, metricValue := mapping.GetReturnCode(appOutput, lbConfMappings)

	logger.Debug("metric = [%s]", metricValue)
	// SNMP critical output
	fmt.Printf("%s\n%s\n%s\n", OID, metricType, metricValue)
}

// applyLoggerSettings : apply the parsed application settings to the logger instance
// 		If the console debug level cannot be parsed, the default value of INFO will be used instead
//		If the file debug level cannot be parsed, the default value of TRACE will be used instead
func applyLoggerSettings() error {
	if err := logger.SetupConsoleLogger(options.ConsoleDebugLevel); err != nil {
		logger.Error("An error occurred when attempting to parse the given console debug level - defaulting to"+
			" [FATAL]. Error [%s]", err.Error())
		_ = logger.SetupConsoleLogger("FATAL")
	}
	if err := logger.SetupFileLogger(options.LogFileLocation, options.FileDebugLevel,
		options.LogAutoFileRotation); err != nil {
		logger.Error("An error occurred when attempting to parse the given console debug level - defaulting to"+
			" [TRACE]. Error [%s]", err.Error())
		err = logger.SetupFileLogger(options.LogFileLocation, "TRACE", options.LogAutoFileRotation)
		// If with a valid logger level the FileLogger cannot be created, return the generated error
		if err != nil {
			return err
		}
	}
	return nil
}
