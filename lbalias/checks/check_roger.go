package checks

import (
	"bufio"
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
	"os"
	"regexp"
)

const ROGER_CURRENT_FILE = "/etc/roger/current.yaml"

type RogerState struct {
}

func (rogerState RogerState) Run(a ...interface{}) interface{} {

	f, err := os.Open(ROGER_CURRENT_FILE)
	if err != nil {
		logger.Error("Can't read file "+ROGER_CURRENT_FILE, err)
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	myState := ""
	logger.Trace("Checking the roger facts")
	state, _ := regexp.Compile("^appstate: *([^ \t\n]+)")
	for scanner.Scan() {
		line := scanner.Text()
		match := state.FindStringSubmatch(line)
		if len(match) > 0 {
			myState = match[1]
		}
	}

	logger.Trace("Roger appstate [%s]", myState)

	if myState == "production" || myState == "ignore_roger" {
		return true
	}

	logger.Debug("The Node will be decommissioned from the LB alias since the roger appstate is [%s] instead of [production]", myState)

	return false

}
