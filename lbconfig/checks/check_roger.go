package checks

import (
	"bufio"
	"os"
	"regexp"

	logger "github.com/sirupsen/logrus"
)

const rogerCurrentFile = "/etc/roger/current.yaml"

type RogerState struct {
}

func (rogerState RogerState) Run(contextLogger *logger.Entry, args ...interface{}) (int, error) {

	f, err := os.Open(rogerCurrentFile)
	if err != nil {
		return -1, err
	}
	defer func() { err = f.Close() }()

	scanner := bufio.NewScanner(f)
	myState := ""
	contextLogger.Trace("Checking the roger facts...")
	state, _ := regexp.Compile("^appstate: *([^ \t\n]+)")
	for scanner.Scan() {
		line := scanner.Text()
		match := state.FindStringSubmatch(line)
		if len(match) > 0 {
			myState = match[1]
		}
	}

	contextLogger.Tracef("Roger appstate [%s]", myState)

	if myState == "production" || myState == "ignore_roger" {
		return 1, nil
	}

	contextLogger.Errorf("The node will not be included in the LB alias since the roger appstate is [%s] instead of [production]", myState)
	return -1, nil

}
