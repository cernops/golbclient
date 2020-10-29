package lbconfig

import (
	"bytes"
	"fmt"
	nested "github.com/antonfisher/nested-logrus-formatter"
	fluentd "github.com/joonix/log"
	logger "github.com/sirupsen/logrus"
	"gitlab.cern.ch/lb-experts/golbclient/helpers/appSettings"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/mapping"
	"os"
	"strings"
)

// AppLauncher : Helper struct that encapsulates the logic of running [lbclient]
type AppLauncher struct {
	AppOptions              appSettings.Options
	lbConfMappings          []*mapping.ConfigurationMapping
	MetricType, MetricValue string
}

func init() {
	logger.SetFormatter(&nested.Formatter{
		ShowFullLevel: 	true,
		FieldsOrder:	[]string{"EVALUATION", "CLI"},
	})
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
	level, err := logger.ParseLevel(l.AppOptions.DebugLevel)
	if err == nil {
		logger.SetLevel(level)
	} else {
		logger.
			WithError(err).
			WithField("logger_level", l.AppOptions.DebugLevel).Error("Unable to parse the desired logger level")
	}

	logger.SetReportCaller(true)
	logger.SetOutput(os.Stdout)

	switch strings.ToLower(l.AppOptions.LoggerMode) {
	case "fluentd":
		logger.SetFormatter(fluentd.NewFormatter())
	case "fluentd_pretty":
		logger.SetFormatter(fluentd.NewFormatter(fluentd.PrettyPrintFormat))
	case "nested":
		return nil
	default:
		logger.WithFields(logger.Fields{"LOG_MODE": l.AppOptions.LoggerMode}).
			Errorf("Unable to set logger format from the given value [%s]", l.AppOptions.LoggerMode)
		logger.Exit(1)
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
		logger.Tracef("Processing configuration file [%s] for aliases [%v]", confMapping.ConfigFilePath, confMapping.AliasNames)
		err := Evaluate(confMapping,
			l.AppOptions.ExecutionConfiguration.MetricTimeout, len(l.AppOptions.ExecutionConfiguration.CheckConfigFilePath) != 0)
		/* Abort if an error occurs */
		if err != nil {
			logger.Warnf("The evaluation of configuration file [%s] failed.", confMapping.ConfigFilePath)
			returnCode = err
		}
		appOutput.WriteString(confMapping.String() + ",")
	}

	l.MetricType, l.MetricValue = mapping.GetReturnCode(appOutput, lbConfMappings)
	logger.Debugf("metric = [%s]", l.MetricValue)

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
