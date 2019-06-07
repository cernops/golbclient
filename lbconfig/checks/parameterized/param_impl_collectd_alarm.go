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
	Name 	string  `json:"name"`
	State 	string `json:"state,omitempty"`
}

func (a alarm) equivalentAlarmState(a2 alarm) bool {
	a2.State = strings.ToUpper(a2.State)

	if len(a.State) == 0 ||
		((a.State == "UNKNOWN" || a.State == "OKAY") && (a2.State == "UNKNOWN" || a2.State == "OKAY")) {
		logger.Trace("Found matching metric state [UNKNOWN|OKAY]. Returning [true]...")
		return true
	}

	if a.State != a2.State {
		logger.Trace("Expected the metric [%s] alarm state to be [%s] but found [%s]. Returning [false]...",
			a.Name, a.State, a2.State)
		return false
	}

	return true
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
		return fmt.Errorf("could not parse [%s]. Error [%v]", []byte(metrics[0]), err)
	}

	// Run the CLI for each metric found
	if ci.cache != nil {
		logger.Trace("Found cached alarm states from previous [collectd] cli run. " +
			"Skipping the cli execution...")

		for _, a := range parsingContainer.Alarm {
			cachedAlarm, err := ci.cache.getAlarm(a.Name)
			if err != nil {
				return err
			}

			(*valueList)[a.Name] = a.equivalentAlarmState(cachedAlarm)
		}
	} else {
		// Initialize the map
		ci.cache = &alarmMetricCache{alarms:make(map[string]alarm)}
		resultsCh := make(chan interface{}, len(supportedAlarmStates))

		logger.Trace("No cache found for previous [collectd] alarm cli [%s]. Running the [collectd] cli...",
			ci.CommandPath)

		// Run the CLI for all the wanted states
		for _, alarmState := range supportedAlarmStates {
			go func(state string) {
				logger.Debug("Running the [collectd] alarm cli [%s] for the state [%s]...", ci.CommandPath, state)
				rawOutput, err := runner.Run(
					ci.CommandPath,
					true,
					0,
					"listval",
					fmt.Sprintf("state=%s", state),
					`| egrep -o "/.*" | cut -c 2- | sort | uniq`)

				if err != nil {
					resultsCh <-fmt.Errorf("failed to run the [collectd] cli with the error [%s]", err.Error())
					close(resultsCh)
				}

				logger.Trace("Raw output from [collectdctl] [%v]", rawOutput)

				// @TODO find way to abort faster (i.e. avoid n) ...?
				cacheAllTheOutput := strings.Split(rawOutput, "\n")
				for _, line := range cacheAllTheOutput {
					logger.Trace("Attempting to cache metric...")
					ci.cache.alarms[line] = alarm{line, state}
					logger.Trace("Cached value for metric [%s] with state [%s]...", line, state)
				}

				logger.Trace("Cached all the metrics for state [%s]...", state)
				resultsCh <-true
			}(alarmState)
		}

		// Wait for all the metrics to be cached
		it := 0
		for result := range resultsCh {
			if result, ok := result.(error); ok {
				return result
			}
			if _, ok := result.(bool); ok {
				it++
				if it >= len(supportedAlarmStates) {
					// Log
					logger.Debug("Finished caching all the metrics for the supported states...")
					close(resultsCh)
				}
			} else {
				return fmt.Errorf("an unexpected result was received from the caching routines [%+v]", result)
			}
		}
	}

	// Check that all the desired metrics have been found and have the desired state
	for _, al := range parsingContainer.Alarm {
		a, err := ci.cache.getAlarm(al.Name)
		if err != nil {
			return fmt.Errorf("unable to find the metric [%s]", al.Name)
		}

		if !a.equivalentAlarmState(al) {
			return fmt.Errorf("failed to find a matching alarm state between [%v] and [%v]", a, al)
		}
	}

	logger.Debug("Metric [%s] requirements successfully validated...", metrics[0])
	return nil
}
