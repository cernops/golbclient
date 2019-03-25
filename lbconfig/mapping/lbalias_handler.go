package mapping

import (
	"bytes"
	"fmt"
	"gitlab.cern.ch/lb-experts/golbclient/helpers/appSettings"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/utils/filehandler"

	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
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
func (cm *ConfigurationMapping) AddConstant(exp string) bool {
	logger.Debug("Adding Constant [%s]", exp)
	// @TODO: Replace with the parser.ParseInterfaceAsFloat (reflection?)
	f, err := strconv.ParseFloat(exp, 32)
	if err != nil {
		logger.Error("Error parsing the floating point number from the value [%s]", exp)
		return false
	}

	cm.MetricValue += int(f)
	return true
}

func (cm *ConfigurationMapping) addAlias(alias string) {
	cm.AliasNames = append(cm.AliasNames, alias)
}

// ReadLBConfigFiles : Returns all the configuration files to be evaluated
func ReadLBConfigFiles(options appSettings.Options) (confFiles []*ConfigurationMapping, err error) {

	tmpConfMap := make(map[string]bool)
	var defaultMapping *ConfigurationMapping

	/* Read the configuration files */
	err = filepath.Walk(options.LbMetricConfDir,
		func(path string, info os.FileInfo, err error) error {
			if info == nil || info.IsDir() || err != nil {
				return nil
			}
			logger.Debug("Checking the file [%v]", path)
			if strings.HasSuffix(path, options.LbMetricDefaultFileName) {
				defaultMapping = NewConfiguration(path)
				logger.Trace("Added the default")
			} else if strings.HasSuffix(path, ".cern.ch") {
				aliasName := strings.Split(path, "lbclient.conf.")[1]
				logger.Trace("Added config for %v", aliasName)
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
		logger.Debug("There is no lbalias configuration file [%v]", options.LbAliasFile)
		return
	}

	formatLbLine := regexp.MustCompile(`^\s*lbalias\s*=\s*(\S+)`)

	for _, alias := range lbAliasesFileContent {
		formatLbLine.Match([]byte(alias))
		if !formatLbLine.Match([]byte(alias)) {
			logger.Trace("Ignoring the line [%v]", alias)
			continue
		}

		/* Abort if a malformed alias is found */
		aliasName := formatLbLine.FindStringSubmatch(alias)[1]
		logger.Trace("Looking for alias [%s]...", aliasName)

		if _, found := tmpConfMap[aliasName]; found {
			logger.Trace("Found configuration file... Skipping...")
			continue
		}
		logger.Trace("Failed to find a configuration file... Adding alias to the generic metric...")
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
