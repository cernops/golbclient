package param

import (
	"encoding/json"
	"fmt"
	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/utils/runner"
	"regexp"
	"strings"
)

type CollectdAlarmImpl struct {
	CommandPath string
	cache *alarmMetricCache
}


type alarmMetricCache struct {
	alarms map[string]alarm
}

func (as alarmMetricCache) getAlarm(key string) (alarm, error) {
	v, ok := as.alarms[key]
	if !ok {
		return alarm{}, fmt.Errorf("unexpected error - unable to find the metric with key [%s] in the " +
			"application cache", key)
	}
	return v, nil
}

type alarmsParsingSchema struct {
	Alarm []alarm `json:"alarms"`
}

type alarm struct {
	name, state string
}

func (a alarm) compareAlarmState(a2 alarm) int {
	if a.state != a2.state {
		logger.Trace("Expected the metric [%s] alarm state to be [%s] but found [%s]. Returning [-1]...",
			a.name, a.state, a2.state)
		return -1
	}
	logger.Trace("Found the metric [%s] with the correct desired state [%s]...", a.name, a.state)
	return 1
}

type alarmState = string
const (
	unknown 	= "UNKNOWN"
	okay		= "OKAY"
	critical	= "CRITICAL"
	warning		= "WARNING"
)

func (ci CollectdAlarmImpl) Name() string {
	return "collectd_alarm"
}

// Run : Runs the [collectdctl] for the found metric's list and populates the expression [valueList] with the values fetched from the CLI.
func (ci CollectdAlarmImpl) Run(metrics []string, valueList *map[string]interface{}) error {
	if len(ci.CommandPath) == 0 {
		ci.CommandPath = "/usr/bin/collectdctl"
	}

	// Parse the metric line into the schema struct
	var parsingContainer alarmsParsingSchema
	err := json.Unmarshal([]byte(metrics[0]), &parsingContainer)
	if err != nil {
		return err
	}

	// Run the CLI for each metric found
	var fetchedValue string
	if ci.cache != nil {
		logger.Trace("Found cached alarm states from previous [collectd] cli run. " +
			"Skipping the cli execution...")

		for _, a := range parsingContainer.Alarm {
			cachedAlarm, err := ci.cache.getAlarm(a.name)
			if err != nil {
				return err
			}

			(*valueList)[a.name] = a.compareAlarmState(cachedAlarm)
		}
	} else {
		logger.Trace("No cache found for previous [collectd] alarm cli [%s]. Running the [collectd] cli...")
		rawOutput, err := runner.Run(ci.CommandPath, true, 0, "listval", "state=")

		logger.Trace("Raw output from [collectdctl] [%v]", rawOutput)
		if err != nil {
			return fmt.Errorf("failed to run the [collectd] cli with the error [%s]", err.Error())
		}
	}

	(*valueList)[metricName] = fetchedValue
	// Log
	logger.Trace("Result of the collectd command: [%v]", (*valueList)[metricName])
	return nil
}
