package runner

import (
	"bytes"
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
	"os/exec"
	"strings"
	"time"
)

// RunCommand : runs a command with the given arguments if this is available. Returns a tuple of he output of the command in the desired format and an error
func RunCommand(pathToCommand string, printRuntime bool, v ...string) (string, error) {
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
		return outBuff.String(), err
	}

	result := strings.TrimRight(outBuff.String(), "\r\n")
	return result, err
}

// RunDirectCommand : runs a command expecting that all the arguments are supplied in the first function parameter
func RunDirectCommand(commandAndArguments string, printRuntime bool) (string, error) {
	logger.Trace("Attempting to run direct command [%s]...", commandAndArguments)
	raw := strings.SplitN(commandAndArguments, " ", 2)
	if len(raw) < 2 {
		logger.Debug("No arguments were passed to the [RunDirectCommand]. In this case, please consider using [RunCommand] instead.")
		return RunCommand(raw[0], printRuntime)
	}
	return RunCommand(raw[0], printRuntime, raw[1])
}

// RunPippedCommand : runs a command with pipes. Note that all the flags should be directly given to the commands.
func RunPippedCommand(pippedCommand string, printRuntime bool) (string, error) {
	return RunCommand("bash", printRuntime, "-c", pippedCommand)
}
