package checks

import (
	"bufio"
	"os"
	"regexp"

	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
)

const rogerCurrentFile = "/etc/roger/current.yaml"

type RogerState struct {
}

func (rogerState RogerState) Run(a ...interface{}) interface{} {

	f, err := os.Open(rogerCurrentFile)
	if err != nil {
		logger.Error("An error occurred when attempting to read the file [%s]. Error [%s]", rogerCurrentFile, err)
		return false
	}
	defer func() { err = f.Close() }()

	scanner := bufio.NewScanner(f)
	myState := ""
	logger.Trace("Checking the roger facts...")
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

	logger.Info("The node will not be included in the LB alias since the roger appstate is [%s] instead of [production]", myState)

	return false

}
