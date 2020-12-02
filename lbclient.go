package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	logger "github.com/sirupsen/logrus"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig"
)

const (
	// OID : SNMP identifier
	OID = ".1.3.6.1.4.1.96.255.1"
	// Version number
	Version = "2.2.0"
	// Release number
	Release = "4"
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
			logger.Fatalf("A fatal error occurred when attempting to parse the application arguments. Error [%s]",
				err.Error())
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
		logger.Fatalf("A fatal error occurred when attempting to apply the logger settings. Error [%s]",
			err.Error())
	}

	// Run the launcher
	err = launcher.Run()
	if err != nil {
		logger.Fatalf("A fatal error occurred when attempting to run the application. Error [%s]", err.Error())
	}
	if len(launcher.AppOptions.ExecutionConfiguration.CheckConfigFilePath) != 0 {
		logger.Info("The configuration file is correct")
	} else {
		// Print the output
		launcher.PrintOutput(OID)
	}
}
