package utils

import (
	"fmt"
	"bufio"
	"gitlab.cern.ch/lb-experts/golbclient/lbalias"
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
	"os"
	"regexp"
	"strings"
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

	if len(aliasNames) == 0 {
		return nil, fmt.Errorf("no alias definition was found inside the [lbaliases] file")
	}
	
	for i := range aliasNames {
		//Check if the file exist
		configFile := ""
		aliasName := aliasNames[i]

		configFileName := strings.Split(options.LbMetricFile, ".conf")[0]
		configFile = fmt.Sprintf("%s.%s.conf", configFileName, aliasName)

		if _, err := os.Stat(configFile); !os.IsNotExist(err) {
			logger.Debug("The specific configuration file exists for [%s]", aliasName)
		} else {
			// Log
			logger.Debug("The config file does not exist for [%s]. Defaulting to [%s]", aliasName,
				options.LbMetricFile)
			configFile = options.LbMetricFile
		}
		lbAliases = append(lbAliases, lbalias.LBalias{Name: aliasName,
		    ChecksDone:     make(map[string]bool),
			Syslog:         options.Syslog,
			ConfigFile:     configFile})

		logger.Debug("File [%s]", configFile)
	}

	return lbAliases, nil

}
