package utils

// Options : Supported application flags
type Options struct {
	// Example of verbosity with level
	CheckOnly  bool   `short:"c" long:"checkonly" description:"Return code shows if lbclient.conf is correct"`
	DebugLevel string `short:"d" long:"debuglevel" default:"ERROR" description:"Debugger output level [TRACE, WARN, DEBUG, INFO, ERROR, FATAL]"`
	Syslog     bool   `short:"s" long:"logsyslog" description:"Log to syslog rather than stdout"`
	NoLogin    bool   `short:"f" long:"ignorenologin" description:"Ignore nologin files"`
	// Configuration files
	LbAliasFile  string `long:"ca" default:"/usr/local/etc/lbaliases" description:"Set an alternative path for the lbaliases configuration file"`
	LbMetricFile string `long:"cm" default:"/usr/local/etc/lbclient.conf" description:"Set an alternative path for the lbclient.conf configuration file"`
	// Example of optional value
	GData []string `short:"g" long:"gdata" description:"Data for OID (required for snmp interface)"`
	NData []string `short:"n" long:"ndata" description:"Data for OID (required for snmp interface)"`
}
