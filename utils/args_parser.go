package utils

// Options : Supported application flags
type Options struct {
	// Example of verbosity with level
	DebugLevel string `short:"d" long:"debuglevel" default:"FATAL" description:"Debugger output level [TRACE, WARN, DEBUG, INFO, ERROR, FATAL]"`

	NoLogin bool `short:"f" long:"ignorenologin" description:"Ignore nologin files"`
	// Configuration files
	LbMetricConfDir         string `long:"cm" default:"/usr/local/etc/" description:"Set the directory where the client should fetch the configuration files from"`
	LbAliasFile             string `long:"ca" default:"/usr/local/etc/lbaliases" description:"Set an alternative path for the lbaliases configuration file"`
	LbMetricDefaultFileName string `short:"c" long:"conf-name" default:"lbclient.conf" description:"Set the default name to be used to lookup for the generic configuration file"`
	// Example of optional value
	GData   []string `short:"g" long:"gdata" description:"Data for OID (required for snmp interface)"`
	NData   []string `short:"n" long:"ndata" description:"Data for OID (required for snmp interface)"`
	Version bool     `short:"v" long:"version" description:"Version of the file"`
}
