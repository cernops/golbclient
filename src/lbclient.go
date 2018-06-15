package main

import (
	"bufio"
	"fmt"
	"lbalias"
	"lbalias/utils/logger"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/jessevdk/go-flags"
)

//Files for configuration
const LBALIASES_FILE = "/usr/local/etc/lbaliases"
const CONFIG_FILE = "/usr/local/etc/lbclient.conf"

type Options struct {
	// Example of verbosity with level
	CheckOnly bool `short:"c" long:"checkonly" description:"Return code shows if lbclient.conf is correct"`
	Debug     bool `short:"d" long:"debug" description:"Debug output"`
	Syslog    bool `short:"s" long:"logsyslog" description:"Log to syslog rather than stdout"`
	NoLogin   bool `short:"f" logn:"ignorenologin" description:"Ignore nologin files"`

	// Example of optional value
	GData []string `short:"g" long:"gdata" description:"Data for OID (required for snmp interface)"`
	NData []string `short:"n" long:"ndata" description:"Data for OID (required for snmp interface)"`
}

const OID = ".1.3.6.1.4.1.96.255.1"

var options Options

var parser = flags.NewParser(&options, flags.Default)

func readLBAliases(filename string) []lbalias.LBalias {

	aliasNames := []string{}
	lbAliases := []lbalias.LBalias{}
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	r, _ := regexp.Compile("^lbalias=([^\t\n\f\r ]+)$")

	for scanner.Scan() {
		line := scanner.Text()
		alias := r.FindStringSubmatch(line)

		if len(alias) > 0 {
			aliasNames = append(aliasNames, alias[1])
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	for i := 0; i < len(aliasNames); i++ {
		//Check if the file exist
		configFile := ""
		aliasName := ""
		if _, err := os.Stat(CONFIG_FILE + "." + aliasNames[i]); !os.IsNotExist(err) {
			if options.Debug {
				fmt.Println("[readLBAliases]  The specific configuration file exists for", aliasNames[i])
			}
			aliasName = aliasNames[i]
			configFile = CONFIG_FILE + "." + aliasNames[i]
		} else {
			if options.Debug {
				fmt.Println("[readLBAliases]  The config file does not exist for", aliasNames[i])
			}
			configFile = CONFIG_FILE
		}
		lbAliases = append(lbAliases, lbalias.LBalias{Name: aliasName,
			Debug:          options.Debug,
			NoLogin:        options.NoLogin,
			Syslog:         options.Syslog,
			ConfigFile:     configFile,
			CheckXsessions: 0})
	}

	return lbAliases

}

func main() {
	_, err := parser.Parse()
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	//Arguments parsed. Let's open the configuration file

	lbAliases := readLBAliases(LBALIASES_FILE)

	if options.Debug {
		logger.LOG(logger.DEBUG, false, "The aliases from the configuration file are [%s]", lbAliases)
	}

	for i := range lbAliases {
		lbAliases[i].Evaluate()
	}
	metricType := "integer"
	metricValue := ""
	if len(lbAliases) == 1 && lbAliases[0].Name == "" {
		metricValue = strconv.Itoa(lbAliases[0].Metric)
	} else {
		keyvaluelist := []string{}
		for _, lbalias := range lbAliases {
			keyvaluelist = append(keyvaluelist, lbalias.Name+"="+strconv.Itoa(lbalias.Metric))
			// Log
			logger.LOG(logger.TRACE, false, "Metric list: [%v]", keyvaluelist)
		}
		metricValue = strings.Join(keyvaluelist, ",")
		metricType = "string"
	}
	if options.Debug {
		fmt.Printf("[main] metric = %s\n", metricValue)
	}
	fmt.Printf("%s\n%s\n%s\n", OID, metricType, metricValue)

}
