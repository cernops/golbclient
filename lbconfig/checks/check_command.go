package checks

import (
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/utils/runner"
	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
	"regexp"
	"strings"
)

type Command struct {
}

/*
	@TODO use the runner API to enable pipped commands support
*/

func (command Command) Run(a ...interface{}) interface{} {
	cmd, _ := regexp.Compile("(?i)(^check[ ]+command)")

	line := a[0].(string)
	found := cmd.Split(line, -1)

	if len(found) > 1 {
		usrCmd := strings.TrimSpace(found[1])
		logger.Trace("Attempting to run command [%s]", usrCmd)
		out, err := runner.RunPippedCommand(usrCmd, true)
		if err != nil {
			logger.Error("The following error was detected when running the [Command] CLI [%v]", err)
			return false
		}

		logger.Debug("Command output [%s]", out)
		return true
	}
	return false
}
