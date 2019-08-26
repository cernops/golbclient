package param

import (
	"encoding/json"
	"fmt"
	logger "github.com/sirupsen/logrus"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/utils/runner"
	"strings"
	"sync"
)

type CollectdAlarmImpl struct {
	CommandPath string
	cache *alarmMetricCache
}

// Needed for eventual concurrent map access
var alarmsMutex = &sync.RWMutex{}

type alarmMetricCache struct {
        // This is a map from the state to the metrics in that state
	alarms map[string][]string
}

func (userAlarm alarmState) equivalentAlarmState(cacheState string) bool {
	if len(userAlarm.State) == 0 && cacheState == "UNKNOWN" || cacheState == "OKAY" {
		logger.Trace("Found matching metric state [UNKNOWN|OKAY]. Returning [true]...")
		return true
	}

	if userAlarm.State != cacheState {
		logger.Tracef("Expected the alarm state to be [%s] but found [%s]. Returning [false]...",
			userAlarm.State, cacheState)
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

	// Only run collectdctl for the user defined states
	userRequiredAlarms := make(map[string]interface{})
	for _, userAlarm := range parsingContainer.Alarm {
		userRequiredAlarms = append(userRequiredAlarms, userAlarm.State)
	}

	// Run the CLI for each metric found
	if ci.cache == nil {
		// Initialize the map
		ci.cache = &alarmMetricCache{alarms:make(map[string][]string)}
		resultsCh := make(chan error)

		logger.Tracef("No cache found for previous [collectd] alarm cli [%s]. Running the [collectd] cli...",
			ci.CommandPath)

		logger.Tracef("Expecting [%d] metrics", len(userRequiredAlarms))

		// Run the CLI for all the wanted states
		for _, alarmState := range userRequiredAlarms {
			go func(state string) {
				logger.Debugf("Running the [collectd] alarm cli [%s] for the state [%s]...", ci.CommandPath, state)
				rawOutput, err := runner.Run(
					ci.CommandPath,
					true,
					0,
					"listval",
					fmt.Sprintf("state=%s", state),
					`| egrep -o "/.*" | cut -c 2- | sort | uniq`)

				if err != nil {
					resultsCh <-fmt.Errorf("failed to run the [collectd] cli with the error [%s]", err.Error())
					return
				}

				logger.Tracef("Raw output from [collectdctl] [%v]", rawOutput)

				// @TODO find way to abort faster (i.e. avoid n) ...?
				cacheAllTheOutput := strings.Split(rawOutput, "\n")
				for _, line := range cacheAllTheOutput {
					if len(strings.TrimSpace(line)) == 0 {
						logger.Debugf("No metrics found for the state [%s]...", state)
						continue
					}

					logger.Trace("Attempting to cache metric...")
					alarmsMutex.Lock()
					if slice, exists := ci.cache.alarms[state]; exists {
						slice = append(slice, line)
						ci.cache.alarms[state] = slice
					} else {
						ci.cache.alarms[state] = []string{line}
					}
					alarmsMutex.Unlock()
					logger.Tracef("Cached value for metric [%s] with state [%s]...", line, state)
				}

				logger.Tracef("Cached all the metrics for state [%s]...", state)
				resultsCh <-nil
			}(alarmState)
		}

		// Wait for all the metrics to be cached
		for i := 0; i < len(userRequiredAlarms); i++ {
			logger.Tracef("Waiting for alarm lookup...")
			r := <-resultsCh
			if result, ok := r.(error); ok {
				return result
			}
		}
	}

	// Check that all the desired metrics have been found and have the desired state
	for _, al := range parsingContainer.Alarm {
		logger.Tracef("Checking that the desired metric [%+v] exists in the cached output [%+v]", al, ci.cache.alarms)

		alarmsMutex.RLock()
		cachedAlarmNames, stateFound := ci.cache.alarms[al.State]
		alarmsMutex.RUnlock()
		if !stateFound {
			return fmt.Errorf("the metric names [%v] was nout found with the desired state [%s]",
				al.Name, al.State)
		}

		found := false
		for _, name := range cachedAlarmNames {
			if name == al.Name {
				logger.Tracef("Found desired metric [%v]...", al)
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("failed to find a matching alarm state [%s] for the metrics [%v]",
				al.State, al.Name)
		}
	}

	logger.Debugf("Metric [%s] requirements successfully validated...", metrics[0])
	return nil
}
