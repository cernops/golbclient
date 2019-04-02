package appSettings

import (
	"time"

	"github.com/jessevdk/go-flags"
)

// LogRotateCfg : Encapsulated options to be used to configure the rotation of the log files
type LogRotateCfg struct {
	Enabled     bool `long:"enabled" description:"Enable the automatic rotation of the log files. (default: false)"`
	LineLimit   int  `long:"linelimit" default:"5000" description:"The maximum amount of lines per log file."`
	BackupLimit int  `long:"backuplimit" default:"9" description:"The maximum amount of backups."`
	MsgBuffer   int  `long:"msgBuffer" default:"100" description:"The message buffer capacity."`
}

type ExecutionConf struct {
	MetricTimeout time.Duration `long:"timeout" default:"30s" description:"The timeout value used when executing a metric line"`
}

// Options : Supported application flags
type Options struct {
	/* Logging */
	LogAutoFileRotation LogRotateCfg `group:"rotatecfg" namespace:"rotatecfg" env-namespace:"rotatecfg" description:"Location of the of log file"`
	LogFileLocation     string       `short:"l" long:"log" default:"log/app.log" description:"Location of the of log file"`
	ConsoleDebugLevel   string       `short:"d" long:"loglevel" default:"FATAL" description:"Console debug level [TRACE, DEBUG, INFO, WARN, ERROR, FATAL]"`
	FileDebugLevel      string       `long:"fdevel" default:"TRACE" description:"File debug level [TRACE, DEBUG, INFO, WARN, ERROR, FATAL]"`
	FileLoggingEnabled  bool         `long:"flog" description:"Activate the file logging service"`
	/* Configuration files */
	LbMetricConfDir         string `long:"cm" default:"/usr/local/etc/" description:"Set the directory where the client should fetch the configuration files from"`
	LbAliasFile             string `long:"ca" default:"/usr/local/etc/lbaliases" description:"Set an alternative path for the lbaliases configuration file"`
	LbMetricDefaultFileName string `short:"c" long:"conf-name" default:"lbclient.conf" description:"Set the default name to be used to lookup for the generic configuration file"`
	/* Execution specific */
	ExecutionConfiguration ExecutionConf `hidden:"true" group:"exec" namespace:"exec" env-namespace:"exec" description:"Execution specific instructions"`
	/* Misc */
	Version bool   `short:"v" long:"version" description:"Version of the file"`
	GData   string `short:"g" long:"gdata" description:"Option needed by the snmp calls"`
	NData   string `short:"n" long:"ndata" description:"Option needed by the snmp calls"`
}

// ParseApplicationSettings : Helper function to handle the parsing of the @see AppArgs schema against a given slice of
// arguments in slice format
func ParseApplicationSettings(args *Options, values []string) error {
	appSettingsParser := flags.NewParser(args, flags.Default)
	_, err := appSettingsParser.ParseArgs(values)
	return err
}
