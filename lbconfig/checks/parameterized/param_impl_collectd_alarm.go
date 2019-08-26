package param

import (
	"encoding/json"
	"fmt"
	logger "github.com/sirupsen/logrus"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/utils/runner"
	"reflect"
	"strings"
	"sync"
)

type CollectdAlarmImpl struct {
	CommandPath string
	cache alarmMetricCache
}

// Needed for eventual concurrent map access
var alarmsMutex = &sync.RWMutex{}
// This is a tuple of metric to its corresponding state
type alarmMetricCache map[string]string

func (ci CollectdAlarmImpl) Name() string {
	return "collectd_alarm"
}

// Run : Runs the [collectdctl] for the found metric's list and populates the expression [valueList] with the values fetched from the CLI.
func (ci CollectdAlarmImpl) Run(metrics []string, valueList *map[string]interface{}) error {
	if len(ci.CommandPath) == 0 {
		ci.CommandPath = "/usr/bin/collectdctl"
	}

	// Parse the state line into the schema struct
	parsingContainer := make([]map[string]interface{}, 0)
	err := json.Unmarshal([]byte(metrics[0]), &parsingContainer)
	if err != nil || len(parsingContainer) == 0 {
		return fmt.Errorf(`could not parse ['%s']. Error [%v]`, []byte(metrics[0]), err)
	}

	// Only run collectdctl for the user defined states
	userRequiredStates := make(map[string]interface{})
	for metric, rawStates := range parsingContainer[0] {
		logger.Tracef("Processing metric [%s] with desired states [%v]...", metric, rawStates)
		requiredStates, err := ci.parseStates(rawStates)
		if err != nil {
			return err
		}

		for _, state := range requiredStates {
			userRequiredStates[state] = nil
		}
	}

	// Run the CLI for each state found
	if ci.cache == nil {
		// Initialize the map
		ci.cache = make(map[string]string)
		resultsCh := make(chan error)

		logger.Tracef("No cache found for previous [collectd] alarm cli [%s]. Running the [collectd] cli...",
			ci.CommandPath)

		logger.Tracef("Expecting [%d] states", len(userRequiredStates))

		// Run the CLI for all the wanted states
		for alarmState := range userRequiredStates {
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

				logger.Tracef("Raw output from [collectdctl] [%v]",
					strings.ReplaceAll(rawOutput, "\n", " "))
				cacheAllTheOutput := strings.Split(rawOutput, "\n")
				for _, line := range cacheAllTheOutput {
					if len(strings.TrimSpace(line)) == 0 {
						logger.Debugf("No metrics found for the state [%s]...", state)
						continue
					}

					logger.Trace("Attempting to cache state...")
					alarmsMutex.Lock()
					ci.cache[line] = state
					alarmsMutex.Unlock()
					logger.Tracef("Cached value for state [%s] with state [%s]...", line, state)
				}

				logger.Tracef("Cached all the metrics for state [%s]...", state)
				resultsCh <-nil
			}(alarmState)
		}

		// Wait for all the metrics to be cached
		for i := 0; i < len(userRequiredStates); i++ {
			logger.Tracef("Waiting for alarm lookup...")
			r := <-resultsCh
			if result, ok := r.(error); ok {
				return result
			}
		}
	}

	// Check that all the desired metrics have been found and have the desired state
	for metric, desiredStates := range parsingContainer[0] {
		logger.Debugf("Checking that the desired [metric:states] pair [%v:%v] exists in the cached output [%+v]",
			metric, desiredStates, ci.cache)
		// Check that at least one state was found
		parsedStates, _ := ci.parseStates(desiredStates)
		if ci.alarmWasFoundInCache(metric, parsedStates) {
			continue
		}

		return fmt.Errorf("the desired [metric:states] pair [%v:%v] was not found", metric, desiredStates)
	}

	logger.Debugf("Metric [%s] requirements successfully validated...", metrics[0])
	return nil
}


func (ci CollectdAlarmImpl) parseStates(states interface{}) (out []string, err error) {
	switch v := states.(type) {
	case []interface{}:
		for _, state := range v {
			if stateValue, correctType := state.(string); correctType {
				out = append(out, stateValue)
			} else {
				return nil, fmt.Errorf("incorrect syntax for value [%v]. Expected [string] but got [%s]",
					state, reflect.ValueOf(state).Kind().String())
			}
		}
	case interface{}:
		if stateValue, correctType := v.(string); correctType {
			out = append(out, stateValue)
		} else {
			return nil, fmt.Errorf("incorrect syntax for value [%v]. Expected [string] but got [%s]",
				v, reflect.ValueOf(v).Kind().String())
		}
	}
	return
}

func (ci CollectdAlarmImpl) alarmWasFoundInCache(metric string, desiredStates []string) bool {
	if alarmState, alarmFound := ci.cache[metric]; alarmFound {
		for _, desiredState := range desiredStates {
			if alarmState == desiredState {
				logger.Tracef("Metric [%s] found with the state [%s]...", metric, desiredState)
				return true
			}
		}
	}
	return false
}