package checks

import (
	"bufio"
	"fmt"
	"lbalias/utils/logger"
	"os"
	"regexp"
)

const ROGER_CURRENT_FILE = "/etc/roger/current.yaml"

type RogerState struct {
	code int
}

func (rogerState RogerState) Code() int {
	return rogerState.code
}

func (rogerState RogerState) Run(a ...interface{}) interface{} {
	lbalias := a[1].(*LBalias)
	if lbalias.RogerState == "" {
		lbalias.RogerState = rogerState.get_roger_fact(lbalias)
	}

	if lbalias.RogerState == "production" {
		logger.Debug("Roger appstate [%s]", lbalias.RogerState)
		return true
	}
	if lbalias.RogerState == "ignore_roger" {
		return true
	}

	logger.Debug("The Node will be decommissioned from the LB alias since the roger appstate [%s] with not [production]", lbalias.RogerState)
	return false

}

func (rogerState RogerState) get_roger_fact(lbalias *LBalias) string {

	f, err := os.Open(ROGER_CURRENT_FILE)
	if err != nil {
		fmt.Println("Can't read file "+ROGER_CURRENT_FILE, err)
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	logger.Debug("Checking the roger facts")
	state, _ := regexp.Compile("^appstate: *([^ \t\n]+)")
	for scanner.Scan() {
		line := scanner.Text()
		match := state.FindStringSubmatch(line)
		if len(match) > 0 {
			logger.Debug("\tCached appstate [%s]", match[1])
			return match[1]
		}
	}
	logger.Debug("Cached appstate is [none]. Ignoring roger appstate")
	return "ignore_roger"
}
