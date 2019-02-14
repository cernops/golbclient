package main

import (
	"fmt"
	"gitlab.cern.ch/lb-experts/golbclient/utils"
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
	"gitlab.cern.ch/lb-experts/golbclient/utils/metrics"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/jessevdk/go-flags"
)

// OID : SNMP identifier
const OID = ".1.3.6.1.4.1.96.255.1"

// Arguments
var options utils.Options

// Flags API
var parser = flags.NewParser(&options, flags.Default)

func main() {
	_, err := parser.Parse()
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	// Set the application logger level
	logger.SetLevelByString(options.DebugLevel)

	//Arguments parsed. Let's open the configuration file
	lbAliases, err := utils.ReadLBAliases(options)
	if err != nil {
		logger.Fatal("Error reading the configuration file. Error [%s]", err.Error())
		os.Exit(1)
	}

	logger.Debug("The aliases from the configuration file are [%v]", lbAliases)

	/* Caching the metric value */
	metricsCache := metrics.NewMetricsCache()

	// Concurrent lbAliases access
	var waitGroup sync.WaitGroup
	waitGroup.Add(len(lbAliases))

	for i := range lbAliases {
		go func(i int) {
			defer waitGroup.Done()
			aliasName := lbAliases[i].Name
			aliasCfg := lbAliases[i].ConfigFile
			logger.Debug("Evaluating the alias [%s]", aliasName)
			if !metricsCache.Contains(aliasCfg) {
				err = lbAliases[i].Evaluate()
				if err != nil {
					logger.Fatal("The evaluation of the alias [%s] failed!", aliasName)
					os.Exit(1)
				}
				metricsCache.Put(aliasCfg, lbAliases[i].Metric)
			} else {
				prevMetric, err := metricsCache.Get(aliasCfg)
				if err != nil {
					logger.Fatal("A fatal error occurred when attempting to fetch a value from the metrics cache" +
						". Error [%s]", err.Error())
				}
				logger.Debug("The alias [%s] configuration file [%s] has already been processed. " +
					"Reusing the metric value [%d]", aliasName, aliasCfg, prevMetric)
				lbAliases[i].Metric = prevMetric
			}
		}(i)
	}

	// Wait for concurrent loop to finish before proceeding
	waitGroup.Wait()
	var keyvaluelist []string
	for _, lbalias := range lbAliases {
		keyvaluelist = append(keyvaluelist, lbalias.Name+"="+strconv.Itoa(lbalias.Metric))
		// Log
		logger.Trace("Metric list: [%v]", keyvaluelist)
	}
	metricValue := strings.Join(keyvaluelist, ",")
	metricType := "string"

	logger.Debug("metric = [%s]", metricValue)
	// SNMP critical output
	fmt.Printf("%s\n%s\n%s\n", OID, metricType, metricValue)
}


