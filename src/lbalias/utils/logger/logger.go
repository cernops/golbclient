package logger

/*
	@TODO : implement direct logger level functions
*/

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/jcelliott/lumber"
)

// ApplicationName : name of the application (possible preffix usage)
const ApplicationName = "lbclient"

// Level : lumber logging levels wrapper
type Level int

// Supported levels of logging
const (
	TRACE Level = iota
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
)

// Logging levels
const log_console_level = lumber.INFO
const log_file_level = lumber.TRACE

// Log file configuration
const logfile = "/usr/local/etc/lbclient.log"
const logfile_rotate = lumber.ROTATE
const logfile_rotate_line_limit = 5000
const logfile_rotate_backup_limit = 9
const logfile_rotate_message_buffer = 100

// Log format configuration
var log_conf_set = false

// Time-format using RFC3339 standard
const log_time_format = "2006-01-02 15:04:05.999999Z07:00"

// Static logger for the application
var filelogger, err = lumber.NewFileLogger(logfile, log_file_level, logfile_rotate, logfile_rotate_line_limit, logfile_rotate_backup_limit, logfile_rotate_message_buffer)
var consolelogger = lumber.NewConsoleLogger(log_console_level)
var consoleLoggerLevel, fileLoggerLevel Level

// SetLevel : sets the logging level for the application
func SetLevel(level Level) {
	if consolelogger != nil {
		consoleLoggerLevel = level
		consolelogger.Level(int(level))
	}
	if filelogger != nil {
		fileLoggerLevel = level
		filelogger.Level(int(level))
	}
}

// getCodeByLevel : internal helper method that returns the logger level based on a string
func getCodeByLevel(level string) Level {
	lwlvl := strings.ToUpper(level)
	switch lwlvl {
	case "TRACE":
		return TRACE
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	default:
		return INFO
	}
}

// SetLevelByString : sets the logger level based on a string [TRACE, WARN, DEBUG, INFO, ERROR, FATAL]. Will default to [INFO] if the given parameter is not valid or an error occurs.
func SetLevelByString(lvlstr string) {
	level := getCodeByLevel(lvlstr)
	if consolelogger != nil {
		consoleLoggerLevel = level
		consolelogger.Level(int(level))
	}
	if filelogger != nil {
		fileLoggerLevel = level
		filelogger.Level(int(level))
	}
}

// GetLevel : gets the logging level for the application
func GetLevel() Level {
	return consoleLoggerLevel | fileLoggerLevel
}

// Variable used for setting the correct caller when using LOGC
var subCaller = false

// LOGC : logging function (logs only to the console)
func LOGC(level Level, format string, v ...interface{}) {
	subCaller = true
	defer func() { subCaller = false }()
	LOG(level, false, format, v...)
}

// LOG : logging function
func LOG(level Level, logToFile bool, format string, v ...interface{}) {
	// Set custom logging once
	setCustomLoggingConf()

	// Get information from the caller function
	setCallerInformation()

	// Warn the user that the log file is not accessible
	checkLogFileAccess(&logToFile)

	// Handle the logging levels
	switch level {
	case TRACE:
		consolelogger.Trace(format, v...)
		if logToFile {
			filelogger.Trace(format, v...)
		}
	case DEBUG:
		consolelogger.Debug(format, v...)
		if logToFile {
			filelogger.Debug(format, v...)
		}
	case INFO:
		consolelogger.Info(format, v...)
		if logToFile {
			filelogger.Info(format, v...)
		}
	case WARN:
		consolelogger.Warn(format, v...)
		if logToFile {
			filelogger.Warn(format, v...)
		}
	case ERROR:
		consolelogger.Error(format, v...)
		if logToFile {
			filelogger.Error(format, v...)
		}
	case FATAL:
		consolelogger.Fatal(format, v...)
		if logToFile {
			filelogger.Fatal(format, v...)
		}
	}
}

// checkLogFileAccess : checks if the logger can access the specified log file path (@TODO test)
func checkLogFileAccess(logToFile *bool) {
	if *logToFile {
		if err != nil {
			consolelogger.Error("The specified log file [%s] is not accessible\n\terror [%s]", logfile, err)
		}
		// Skip if error is detected
		*logToFile = (err == nil)
	}
}

// setCallerInformation : internal helper method that uses the call stack to retrieve the name of the method that calls this function
func setCallerInformation() {
	funcDetails := "n/a"
	// Get the caller of LOG
	depth := 2
	if subCaller {
		depth = 3
	}
	pc, _, ln, ok := runtime.Caller(depth)
	if ok {
		fun := runtime.FuncForPC(pc)
		if ok && fun != nil {
			funcDetails = fmt.Sprintf("[%s:%d]", fun.Name(), ln)
		}
	}

	// Set the prefix for both loggers
	if consolelogger != nil {
		consolelogger.Prefix(funcDetails)
	}
	if filelogger != nil {
		filelogger.Prefix(funcDetails)
	}

	// Set prefixes
	if consolelogger != nil {
		consolelogger.Prefix(funcDetails)
	}
	if filelogger != nil {
		filelogger.Prefix(funcDetails)
	}
}

// setCustomLoggingConf : internal method used to set the default configuration for the logger (e.g. timeformat)
func setCustomLoggingConf() {
	if !log_conf_set {
		if consolelogger != nil {
			consolelogger.TimeFormat(log_time_format)
		}
		if filelogger != nil {
			filelogger.TimeFormat(log_time_format)
		}
		log_conf_set = true
	}
}

/*************************** EXTENSIONS ***************************/

// Trace : log with the [TRACE] logging level
func Trace(format string, v ...interface{}) {
	LOGC(TRACE, format, v...)
}

// Warn : log with the [WARN] logging level
func Warn(format string, v ...interface{}) {
	LOGC(WARN, format, v...)
}

// Debug : log with the [DEBUG] logging level
func Debug(format string, v ...interface{}) {
	LOGC(DEBUG, format, v...)
}

// Info : log with the [INFO] logging level
func Info(format string, v ...interface{}) {
	LOGC(INFO, format, v...)
}

// Error : log with the [ERROR] logging level
func Error(format string, v ...interface{}) {
	LOGC(ERROR, format, v...)
}

// Fatal : log with the [FATAL] logging level
func Fatal(format string, v ...interface{}) {
	LOGC(FATAL, format, v...)
}
