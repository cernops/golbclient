package checks

import (
	"gitlab.cern.ch/lb-experts/golbclient/lbalias/utils/runner"
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
	"regexp"
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

	if len(found) > 0 {
		out, err := runner.RunDirectCommand(found[1], true, true)
		if err != nil {
			logger.Error("The following error was detected when running the [Command] CLI [%s]", err.Error())
			return false
		}
		logger.Debug("Output [%s]. Return code [0]", string(out))

		return true
	}
	return false
}
