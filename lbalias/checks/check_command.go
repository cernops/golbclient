package checks

import (
	"os/exec"
	"regexp"
	"strings"
	"gitlab.cern.ch/lb-experts/golbclient/lbalias/utils/runner"
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
)

type Command struct {

}

/*
	@TODO use the runner API to enable pipped commands support
 */

func (command Command) Run(a ...interface{}) interface{} {
	cmd, _ := regexp.Compile("(?i)(^CHECK[ ]+command)[ ]*([^ ]+)[ ]*(.*)")

	line := a[0].(string)
	found := cmd.FindStringSubmatch(line)

	if len(found) > 0 {
		args := []string{}
		if found[2] != "" {
			args = strings.Split(found[2], " ")
		}

		out, err := runner.RunCommand(found[1], true, true, args...)
		if err != nil {
			logger.Error("The following error was detected when running the [Command] CLI [%s]", err.Error())
			rc := err.(*exec.ExitError)
			logger.Error("Recovered panic: [%s] [%s]. Ignoring script return code [%s].", found[1], found[2], err.Error())
			logger.Debug("Return code [%s]", rc.Error()) // @TODO test
			return false

		}
		logger.Debug("Output [%s]. Return code [0]", string(out))

		return true
	}
	return false
}
