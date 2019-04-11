package lbconfig

import (
	"bytes"
	"fmt"

	"gitlab.cern.ch/lb-experts/golbclient/helpers/appSettings"
	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/mapping"
)

// AppLauncher : Helper struct that encapsulates the logic of running [lbclient]
type AppLauncher struct {
	AppOptions              appSettings.Options
	lbConfMappings          []*mapping.ConfigurationMapping
	MetricType, MetricValue string
}

// NewAppLauncher : Factory-pattern function that creates and returns a new @see AppLauncher struct instance pointer
func NewAppLauncher() *AppLauncher {
	return &AppLauncher{}
}

// ParseApplicationArguments : Helper function that wraps the functionality responsible for the parsing of the application
// arguments. Receives the arguments in the format of a slice of strings
func (l *AppLauncher) ParseApplicationArguments(args []string) error {
	return appSettings.ParseApplicationSettings(&l.AppOptions, args)
}

// applyLoggerSettings : apply the parsed application settings to the logger instance
// 		If the console debug level cannot be parsed, the default value of INFO will be used instead
//		If the file debug level cannot be parsed, the default value of TRACE will be used instead
func (l *AppLauncher) ApplyLoggerSettings() error {
	if err := logger.SetupConsoleLogger(l.AppOptions.ConsoleDebugLevel); err != nil {
		logger.Error("An error occurred when attempting to parse the given console debug level - defaulting to"+
			" [FATAL]. Error [%s]", err.Error())
		_ = logger.SetupConsoleLogger("FATAL")
	}

	// The file logger is disabled by default
	if l.AppOptions.FileLoggingEnabled {
		if err := logger.SetupFileLogger(l.AppOptions.LogFileLocation, l.AppOptions.FileDebugLevel,
			l.AppOptions.LogAutoFileRotation); err != nil {
			logger.Error("An error occurred when attempting to parse the given console debug level - defaulting to"+
				" [TRACE]. Error [%s]", err.Error())
			err = logger.SetupFileLogger(l.AppOptions.LogFileLocation, "TRACE", l.AppOptions.LogAutoFileRotation)
			// If with a valid logger level the FileLogger cannot be created, return the generated error
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Run : This function encapsulates the following steps of the application runtime:
// 	1 - Reads the configuration files and creates the correlated @see mapping.ConfigurationMapping instances
// 	2 - Once this function exits, the correlated @see AppLauncher instance gets its MetricType and MetricValue fields
//      populated and ready to be used
func (l *AppLauncher) Run() error {
	lbConfMappings, err := mapping.ReadLBConfigFiles(l.AppOptions)
	if err != nil {
		return err
	}

	// Application output
	var appOutput bytes.Buffer
	var returnCode error
	// Evaluate for each of the configuration files found
	for _, confMapping := range lbConfMappings {
		logger.Trace("Processing configuration file [%s] for aliases [%v]", confMapping.ConfigFilePath, confMapping.AliasNames)
		err := Evaluate(confMapping, l.AppOptions.ExecutionConfiguration.MetricTimeout, l.AppOptions.ExecutionConfiguration.CheckConfig)
		/* Abort if an error occurs */
		if err != nil {
			logger.Warn("The evaluation of configuration file [%s] failed.", confMapping.ConfigFilePath)
			returnCode = err
		}
		appOutput.WriteString(confMapping.String() + ",")
	}

	l.MetricType, l.MetricValue = mapping.GetReturnCode(appOutput, lbConfMappings)
	logger.Debug("metric = [%s]", l.MetricValue)

	return returnCode
}

// Output : Returns the formatted output of the @see AppLauncher instance
func (l *AppLauncher) Output(oid string) string {
	return fmt.Sprintf("%s\n%s\n%s\n", oid, l.MetricType, l.MetricValue)
}

// PrintOutput : Prints the output of the application
func (l *AppLauncher) PrintOutput(oid string) {
	fmt.Printf(l.Output(oid))
}
