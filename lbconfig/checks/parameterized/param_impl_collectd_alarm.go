package param

import (
	"encoding/json"
	"fmt"
	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/utils/runner"
	"strings"
)

type CollectdAlarmImpl struct {
	CommandPath string
	cache *alarmMetricCache
}


type alarmMetricCache struct {
	alarms map[string]alarm
}

var supportedAlarmStates = []string{
	"UNKNOWN", "ERROR", "WARNING", "OKAY",
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
	Name 	string
	State 	*string `json:",omitempty"`
}

func (a alarm) compareAlarmState(a2 alarm) int {
	if a.State == nil {
		logger.Trace("No state was found to be defined by the user...")
		return 1
	}

	if a.State != a2.State {
		logger.Trace("Expected the metric [%s] alarm state to be [%s] but found [%s]. Returning [-1]...",
			a.Name, a.State, a2.State)
		return -1
	}
	logger.Trace("Found the metric [%s] with the correct desired state [%s]...", a.Name, a.State)
	return 1
}


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
			cachedAlarm, err := ci.cache.getAlarm(a.Name)
			if err != nil {
				return err
			}

			(*valueList)[a.Name] = a.compareAlarmState(cachedAlarm)
		}
	} else {
		logger.Trace("No cache found for previous [collectd] alarm cli [%state]. Running the [collectd] cli...",
			ci.CommandPath)

		// Run the CLI for all the wanted states
		for _, state := range supportedAlarmStates {
			logger.Debug("Running the [collectd] alarm cli [%state] for the state [%state]...", ci.CommandPath, state)
			rawOutput, err := runner.Run(
				ci.CommandPath,
				true,
				0,
				"listval",
				fmt.Sprintf("state=%state", state),
				`| egrep -o "/.*" | cut -c 2- | sort | uniq`)

			logger.Trace("Raw output from [collectdctl] [%v]", rawOutput)

			// Find all the alarms with a Regex
			// @TODO need to find a way to avoid n^2

			if err != nil {
				return fmt.Errorf("failed to run the [collectd] cli with the error [%state]", err.Error())
			}
		}
	}

	(*valueList)[metricName] = fetchedValue
	// Log
	logger.Trace("Result of the collectd command: [%v]", (*valueList)[metricName])
	return nil
}
