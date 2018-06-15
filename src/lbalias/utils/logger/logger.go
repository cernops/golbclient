package logger

import (
	"fmt"
	"runtime"

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

// RFC3339
const log_time_format = "2006-01-02 15:04:05.999999Z07:00"

// Static logger for the application
var filelogger, err = lumber.NewFileLogger(logfile, log_file_level, logfile_rotate, logfile_rotate_line_limit, logfile_rotate_backup_limit, logfile_rotate_message_buffer)
var consolelogger = lumber.NewConsoleLogger(log_console_level)

// SetLevel : sets the logging level for the application
func SetLevel(level Level) {
	if consolelogger != nil {
		consolelogger.Level(int(level))
	}
	if filelogger != nil {
		filelogger.Level(int(level))
	}
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

func checkLogFileAccess(logToFile *bool) {
	if *logToFile {
		if err != nil {
			consolelogger.Error("The specified log file [%s] is not accessible\n\terror [%s]", logfile, err)
		}
		// Skip if error is detected
		*logToFile = (err == nil)
	}
}

func setCallerInformation() {
	funcDetails := "n/a"
	// Get the caller of LOG
	pc, _, ln, ok := runtime.Caller(2)
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

func setCustomLoggingConf() {
	if !log_conf_set {
		if consolelogger != nil {
			consolelogger.Level(lumber.TRACE)
			consolelogger.TimeFormat(log_time_format)
		}
		if filelogger != nil {
			filelogger.Level(lumber.TRACE)
			filelogger.TimeFormat(log_time_format)
		}

		log_conf_set = true
	}
}
