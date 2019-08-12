package mapping

import (
	"bytes"
	"fmt"
	"gitlab.cern.ch/lb-experts/golbclient/helpers/appSettings"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/utils/filehandler"

	logger "github.com/sirupsen/logrus"
)

// ConfigurationMapping : object with the config
type ConfigurationMapping struct {
	ConfigFilePath 	string
	AliasNames     	[]string
	MetricValue    	int
	//ChecksDone     map[string]bool
	Default			bool
}

// NewConfiguration : Creates a Configuration object
func NewConfiguration(path string, aliasName ...string) *ConfigurationMapping {
	var cm ConfigurationMapping
	// cm.ChecksDone = make(map[string]bool)
	cm.ConfigFilePath = path
	if aliasName != nil && len(aliasName) > 0 {
		cm.AliasNames = aliasName
	} else {
		cm.Default = true
	}
	cm.MetricValue = 0
	return &cm
}



func (cm ConfigurationMapping) String() string {
	out := bytes.Buffer{}
	for i := 0; i < len(cm.AliasNames); i++ {
		out.WriteString(fmt.Sprintf("%s=%d", cm.AliasNames[i], cm.MetricValue))
		if i < len(cm.AliasNames)-1 {
			out.WriteString(",")
		}
	}
	return out.String()
}

func (cm *ConfigurationMapping) addAlias(alias string) {
	cm.AliasNames = append(cm.AliasNames, alias)
}

// ReadLBConfigFiles : Returns all the configuration files to be evaluated
func ReadLBConfigFiles(options appSettings.Options) (confFiles []*ConfigurationMapping, err error) {
	if len(options.ExecutionConfiguration.CheckConfigFilePath) != 0 {
		confFiles = append(confFiles, NewConfiguration(options.ExecutionConfiguration.CheckConfigFilePath))
		return
	}

	tmpConfMap := make(map[string]bool)
	var defaultMapping *ConfigurationMapping

	/* Read the configuration files */
	err = filepath.Walk(options.LbMetricConfDir,
		func(path string, info os.FileInfo, err error) error {
			if info == nil || info.IsDir() || err != nil {
				return nil
			}
			logger.Debugf("Checking the file [%v]", path)
			if info.Name() ==  options.LbMetricDefaultFileName {
				defaultMapping = NewConfiguration(path)
				logger.Trace("Added the default")
			} else if strings.HasSuffix(info.Name(), ".cern.ch") && strings.HasPrefix(info.Name(), "lbclient.conf") {
				aliasName := strings.TrimSpace(strings.Split(path, "lbclient.conf.")[1])
				logger.Tracef("Added config for %v", aliasName)
				confFiles = append(confFiles, NewConfiguration(path, aliasName))
				tmpConfMap[aliasName] = true
			}
			return nil
		})
	/* Abort if no configuration files were found */
	if len(confFiles) == 0 && defaultMapping == nil {
		return nil, fmt.Errorf("no configuration files found in the supplied directory [%s]",
			options.LbMetricConfDir)
	}

	/* Read the aliases */
	lbAliasesFileContent, err := filehandler.ReadAllLinesFromFile(options.LbAliasFile)
	if err != nil {
		logger.Debugf("There is no lbalias configuration file [%v]", options.LbAliasFile)
		return nil, err
	}

	formatLbLine := regexp.MustCompile(`^\s*lbalias\s*=\s*(\S+)`)

	for _, alias := range lbAliasesFileContent {
		formatLbLine.Match([]byte(alias))
		if !formatLbLine.Match([]byte(alias)) {
			logger.Tracef("Ignoring the line [%v]", alias)
			continue
		}

		/* Abort if a malformed alias is found */
		aliasName := formatLbLine.FindStringSubmatch(alias)[1]
		logger.Tracef("Looking for alias [%s]...", aliasName)

		if _, found := tmpConfMap[aliasName]; found {
			logger.Trace("Found configuration file... Skipping...")
			continue
		}
		logger.Tracef("Failed to find a configuration file... Adding alias to the generic metric...")
		/* Add the stranded alias to the default configuration file */
		if defaultMapping == nil {
			return nil, fmt.Errorf(" [%v/%v] file  not found, and the alias [%v]"+
				" does not have a specific configuration",
				options.LbMetricConfDir, options.LbMetricDefaultFileName, aliasName)
		}
		defaultMapping.addAlias(aliasName)
	}
	/* */
	if defaultMapping != nil && len(defaultMapping.AliasNames) > 0 {
		confFiles = append(confFiles, defaultMapping)
	}

	return confFiles, err
}

// GetReturnCode : checks if the return code should be a string or an integer
func GetReturnCode(appOutput bytes.Buffer, lbConfMappings []*ConfigurationMapping) (metricType, metricValue string) {
	if len(lbConfMappings) == 1 {
		metricType = "integer"
		metricValue = fmt.Sprintf("%v", lbConfMappings[0].MetricValue)
	} else {
		metricType = "string"
		metricValue = strings.TrimRight(appOutput.String(), ",")
	}
	return
}
