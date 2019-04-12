package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig"
)

const (
	// OID : SNMP identifier
	OID = ".1.3.6.1.4.1.96.255.1"
	// Version number
	Version = "2.0.1"
	// Release number
	Release = "1"
)

func main() {
	// Create a new instance of the launcher that will be responsible for the runtime
	launcher := lbconfig.NewAppLauncher()

	// Parse the application arguments
	err := launcher.ParseApplicationArguments(os.Args[1:])
	if err != nil {
		// Check if the help flag [--help || -h] was detected
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			logger.Fatal("A fatal error occurred when attempting to parse the application arguments. Error [%s]",
				err.Error())
			os.Exit(1)
		}
	}

	// Check if the version flag was detected
	if launcher.AppOptions.Version {
		fmt.Printf("lbclient version %s.%s\n", Version, Release)
		os.Exit(0)
	}

	// Apply the logger settings
	err = launcher.ApplyLoggerSettings()
	if err != nil {
		logger.Fatal("A fatal error occurred when attempting to apply the logger settings. Error [%s]",
			err.Error())
		os.Exit(1)
	}

	// Run the launcher
	err = launcher.Run()
	if err != nil {
		logger.Fatal("A fatal error occurred when attempting to run the application. Error [%s]", err.Error())
		os.Exit(1)
	}
	if len(launcher.AppOptions.ExecutionConfiguration.CheckConfigFilePath) != 0  {
		logger.Info("The configuration file is correct")
	} else {
		// Print the output
		launcher.PrintOutput(OID)
	}
}
