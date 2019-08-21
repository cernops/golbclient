package checks

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"syscall"

	logger "github.com/sirupsen/logrus"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/utils/runner"
)

type Command struct{}

/*
	@TODO use the runner API to enable pipped commands support
*/

const (
	CommandNotFoundCode 	= 127
	CommandNotExecutable 	= 126
)

func (command Command) Run(args ...interface{}) (int, error) {
	cmd, _ := regexp.Compile("(?i)(^check[ ]+command)")
	line := args[0].(string)
	found := cmd.Split(line, -1)

	if len(found) > 1 {
		usrCmd := strings.TrimSpace(found[1])
		logger.Tracef("Attempting to run command [%s]", usrCmd)
		out, err := runner.RunCommand(usrCmd, true, 0)
		if err != nil {
			logger.Errorf("The command [%s] failed. Error [%v]", usrCmd, err)
			if exitErr, ok := err.(*exec.ExitError); ok {
				if exitErr.Exited() {
					if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
						if status.ExitStatus() == CommandNotFoundCode || status.ExitStatus() == CommandNotExecutable {
							// If command not found or not executable
							return -1, err
						}
					}
					// If command exited with status != 0
					return -1, nil
				}
			}

			return -1, err
		}

		logger.Debugf("Command output [%s]", out)
		return 1, nil
	}

	return -1, fmt.Errorf("there was no command to execute in the line [%s]", line)
}

