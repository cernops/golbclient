package runner

import (
	"bytes"
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
	"os/exec"
	"strings"
	"time"
)

// RunCommand : runs a command with the given arguments if this is available. Returns a tuple of he output of the command in the desired format and an error
func RunCommand(pathToCommand string, printErrors bool, printRuntime bool, v ...string) (string, error) {
	var now int64
	if printRuntime {
		now = time.Now().UnixNano() / int64(time.Millisecond)
		defer func() {
			newNow := time.Now().UnixNano() / int64(time.Millisecond)
			logger.Debug("\t Runtime: %dms", newNow-now)
		}()
	}

	cmd := exec.Command(pathToCommand, v...)
	var errBuff bytes.Buffer
	var outBuff bytes.Buffer
	cmd.Stderr = &errBuff
	cmd.Stdout = &outBuff
	err := cmd.Run()
	if err != nil {
		if printErrors {
			errString := strings.TrimRight(errBuff.String(), "\r\n")
			logger.Error("Execution failed with the following error [%s]", errString)
		}
		return outBuff.String(), err
	}

	result := strings.TrimRight(string(outBuff.String()), "\r\n")
	return result, err
}

// RunPippedCommand : runs a command with the given arguments if this is available. Returns a tuple of he output of the command in the desired format and an error
func RunPippedCommand(pippedCommand string, printErrors bool, printRuntime bool) (string, error) {
	return RunCommand("bash", printErrors, printRuntime, "-c", pippedCommand)
}