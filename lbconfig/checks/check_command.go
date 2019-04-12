package checks

import (
	"fmt"
	"regexp"
	"strings"

	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/utils/runner"
)

type Command struct{}

/*
	@TODO use the runner API to enable pipped commands support
*/

func (command Command) Run(args ...interface{}) (int, error) {
	cmd, _ := regexp.Compile("(?i)(^check[ ]+command)")
	line := args[0].(string)
	found := cmd.Split(line, -1)

	if len(found) > 1 {
		usrCmd := strings.TrimSpace(found[1])
		logger.Trace("Attempting to run command [%s]", usrCmd)
		out, err := runner.RunCommand(usrCmd, true, 0)
		if err != nil {
			logger.Error("The command [%s] failed", usrCmd)
			return -1, err
		}

		logger.Debug("Command output [%s]", out)
		return 1, nil
	}

	return -1, fmt.Errorf("there was no command to execute in the line [%s]", line)
}
