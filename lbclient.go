package main

import (
	"bytes"
	"fmt"
	"os"
	"sync"

	"github.com/jessevdk/go-flags"
	"gitlab.cern.ch/lb-experts/golbclient/lbalias"
	"gitlab.cern.ch/lb-experts/golbclient/utils"
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
)

const (
	// OID : SNMP identifier
	OID = ".1.3.6.1.4.1.96.255.1"
	// Version number
	Version = "2.0"
	// Release number
	Release = "7"
)

// Arguments
var options utils.Options

// Flags API
var parser = flags.NewParser(&options, flags.Default)

func main() {
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

	// Set the application logger level
	logger.SetLevelByString(options.DebugLevel)

	// Arguments parsed. Let's open the configuration file
	var lbConfMappings []*utils.ConfigurationMapping
	lbConfMappings, err = utils.ReadLBConfigFiles(options)
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

		go func(confMapping *utils.ConfigurationMapping) {
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

	metricType, metricValue := utils.GetReturnCode(appOutput, lbConfMappings)

	logger.Debug("metric = [%s]", metricValue)
	// SNMP critical output
	fmt.Printf("%s\n%s\n%s\n", OID, metricType, metricValue)
}
