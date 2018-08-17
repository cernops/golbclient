package utils

import (
	"bufio"
	"gitlab.cern.ch/lb-experts/golbclient/lbalias"
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
	"os"
	"regexp"
)

func ReadLBAliases(options Options) (lbAliases []lbalias.LBalias, err error) {

	filename := options.LbAliasFile
	aliasNames := []string{}
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
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
		logger.Fatal("[%v]", err)
		return nil, err
	}
	for i := range aliasNames {
		//Check if the file exist
		configFile := ""
		aliasName := aliasNames[i]
		if _, err := os.Stat(options.LbMetricFile + "." + aliasNames[i]); !os.IsNotExist(err) {
			// Log
			logger.Debug("The specific configuration file exists for [%s]", aliasNames[i])

			configFile = options.LbMetricFile + "." + aliasNames[i]
		} else {
			// Log
			logger.Debug("The config file does not exist for [%s]", aliasNames[i])
			configFile = options.LbMetricFile
		}
		lbAliases = append(lbAliases, lbalias.LBalias{Name: aliasName,
		    ChecksDone:     make(map[string]bool),
			Syslog:         options.Syslog,
			ConfigFile:     configFile})
	}

	return lbAliases, nil

}
