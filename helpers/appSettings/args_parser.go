package appSettings

import (
	"time"

	"github.com/jessevdk/go-flags"
)

// ExecutionConf options for the execution
type ExecutionConf struct {
	MetricTimeout       time.Duration `hidden:"true" long:"timeout" default:"30s" description:"The timeout value used when executing a metric line"`
	CheckConfigFilePath string        `short:"t" long:"checkconfig" description:"Checks that the supplied configuration file is correct. Returns 0 if it is valid"  `
}

// Options : Supported application flags
type Options struct {
	/* Logging */
	LoggerMode string `short:"m" long:"logMode" default:"nested" description:"Logger mode [fluentd, fluentd_pretty, nested]"`
	DebugLevel string `short:"d" long:"loglevel" default:"FATAL" description:"Logger level [TRACE, DEBUG, INFO, WARN, ERROR, FATAL, CRITICAL]"`
	/* Configuration files */
	LbMetricConfDir         string `long:"cm" default:"/usr/local/etc/" description:"Set the directory where the client should fetch the configuration files from"`
	LbAliasFile             string `long:"ca" default:"/usr/local/etc/lbaliases" description:"Set an alternative path for the lbaliases configuration file"`
	LbMetricDefaultFileName string `short:"c" long:"conf-name" default:"lbclient.conf" description:"Set the default name to be used to lookup for the generic configuration file"`
	LbPostFile              string `short:"p" long:"post" description:"Set the default file for the configuration of the ermis communication"`
	/* Execution specific */
	ExecutionConfiguration ExecutionConf `group:"exec" namespace:"exec" env-namespace:"exec" description:"Execution specific instructions"`
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
