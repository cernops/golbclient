package logger

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"gitlab.cern.ch/lb-experts/golbclient/helpers/appSettings"
	"gitlab.cern.ch/lb-experts/golbclient/lbalias/utils/filehandler"

	"github.com/jcelliott/lumber"
)

// Level : lumber logging levels wrapper
type Level = int

// Supported levels of logging
const (
	TRACE Level = iota
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
)

// Log format configuration
var logConfSet = false

// Time-format using RFC3339 standard
const logTimeFormat = "2006-01-02 15:04:05.999999Z07:00"

// Loggers to be used by the application
var fileLogger lumber.Logger

// Default logger for tests
var consoleLogger = lumber.NewConsoleLogger(FATAL)

// GetLevelByString : returns the logger level representation by string [TRACE, WARN, DEBUG, INFO, ERROR, FATAL].
// Will default to [INFO] if the given parameter is not valid or an error occurs
func GetLevelByString(lvlstr string) (Level, error) {
	return getCodeByLevel(lvlstr)
}

// SetLevel : sets the logging level (console) for the application
func SetLevel(level Level) {
	if consoleLogger != nil {
		consoleLogger.Level(level)
	}
}

// GetLevel : returns the current logging level (console) being used by the application
func GetLevel() Level {
	return consoleLogger.GetLevel()
}

// SetupConsoleLogger : initializes the logger @see lumber.ConsoleLogger instance pointer
func SetupConsoleLogger(level string) error {
	if parsedLevel, err := GetLevelByString(level); err != nil {
		return err
	} else {
		consoleLogger = lumber.NewConsoleLogger(parsedLevel)
		return nil
	}
}

// SetupFileLogger : sets the location that will be used to write the application log file & initializes the
// logger @see lumber.FileLogger instance pointer
func SetupFileLogger(logFileLocation string, l string, cfg appSettings.LogRotateCfg) error {
	if _, err := filehandler.CreateFileInDir(logFileLocation, os.ModePerm); err != nil {
		return err
	}

	// Abort if the level could not be parsed
	level, err := GetLevelByString(l)
	if err != nil {
		return err
	}

	// Is auto rotation enabled?
	logfileRotate := lumber.ROTATE
	if !cfg.Enabled {
		logfileRotate = lumber.APPEND
	}

	// Attempt to create the logger instance
	if fLogger, err := lumber.NewFileLogger(logFileLocation, level, logfileRotate, cfg.LineLimit, cfg.BackupLimit,
		cfg.MsgBuffer); err != nil {
		return err
	} else {
		fileLogger = fLogger
	}
	return nil
}

// LOG : logging function
func log(printFunc func(format string, v ...interface{}), format string, v ...interface{}) {
	// Set custom logging once
	setCustomLoggingConf()

	// Get information from the caller function
	setCallerInformation()

	// Call the printer function of the correlated logger
	printFunc(format, v...)
}

// setCallerInformation : internal helper method that uses the call stack to retrieve the name of the method that
// calls this function
func setCallerInformation() {
	funcDetails := "n/a"
	// Get the caller of LOG
	depth := 3
	pc, _, ln, ok := runtime.Caller(depth)
	if ok {
		fun := runtime.FuncForPC(pc)
		if ok && fun != nil {
			funcDetails = fmt.Sprintf("[%s:%d]", fun.Name(), ln)
		}
	}

	// Set the prefix for both loggers
	if consoleLogger != nil {
		consoleLogger.Prefix(funcDetails)
	}
	if fileLogger != nil {
		fileLogger.Prefix(funcDetails)
	}
}

// setCustomLoggingConf : internal method used to set the default configuration for the logger (e.g. timeformat)
func setCustomLoggingConf() {
	if !logConfSet {
		setCallerInformation()
		if consoleLogger != nil {
			consoleLogger.TimeFormat(logTimeFormat)
		}
		if fileLogger != nil {
			fileLogger.TimeFormat(logTimeFormat)
		}
		logConfSet = true
	}
}

// getCodeByLevel : internal helper method that returns the logger level based on a string
func getCodeByLevel(level string) (l Level, err error) {
	switch strings.ToUpper(level) {
	case "TRACE":
		l = TRACE
	case "DEBUG":
		l = DEBUG
	case "INFO":
		l = INFO
	case "WARN":
		l = WARN
	case "ERROR":
		l = ERROR
	case "FATAL":
		l = FATAL
	default:
		err = fmt.Errorf("the specified level [%s] could not be parsed", level)
	}
	return
}

/*************************** EXTENSIONS ***************************/

// Trace : log with the [TRACE] logging level
func Trace(format string, v ...interface{}) {
	log(consoleLogger.Trace, format, v...)
	if fileLogger != nil {
		log(fileLogger.Trace, format, v...)
	}
}

// Warn : log with the [WARN] logging level
func Warn(format string, v ...interface{}) {
	log(consoleLogger.Warn, format, v...)
	if fileLogger != nil {
		log(fileLogger.Warn, format, v...)
	}
}

// Debug : log with the [DEBUG] logging level
func Debug(format string, v ...interface{}) {
	log(consoleLogger.Debug, format, v...)
	if fileLogger != nil {
		log(fileLogger.Debug, format, v...)
	}
}

// Info : log with the [INFO] logging level
func Info(format string, v ...interface{}) {
	log(consoleLogger.Info, format, v...)
	if fileLogger != nil {
		log(fileLogger.Info, format, v...)
	}
}

// Error : log with the [ERROR] logging level
func Error(format string, v ...interface{}) {
	log(consoleLogger.Error, format, v...)
	if fileLogger != nil {
		log(fileLogger.Error, format, v...)
	}
}

// Fatal : log with the [FATAL] logging level
func Fatal(format string, v ...interface{}) {
	log(consoleLogger.Fatal, format, v...)
	if fileLogger != nil {
		log(fileLogger.Fatal, format, v...)
	}
}
