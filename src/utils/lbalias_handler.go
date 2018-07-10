package utils

import (
	"bufio"
	"lbalias/checks"
	"lbalias/utils/logger"
	"log"
	"os"
	"regexp"
)

func ReadLBAliases(options Options) []checks.LBalias {

	filename := options.LbAliasFile
	aliasNames := []string{}
	lbAliases := []checks.LBalias{}
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
		lbAliases = append(lbAliases, checks.LBalias{Name: aliasName,
			NoLogin:        options.NoLogin,
			Syslog:         options.Syslog,
			ConfigFile:     configFile,
			CheckXsessions: 0})
	}

	return lbAliases

}
