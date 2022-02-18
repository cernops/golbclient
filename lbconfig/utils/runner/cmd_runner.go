package runner

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	logger "github.com/sirupsen/logrus"
)

// Run : runs a command with the given arguments if this is available. Returns a tuple of he output of the command in the desired format and an error
func Run(pathToCommand string, printRuntime bool, timeout time.Duration, v ...string) (output string, err error, stderr string) {
	var now int64
	if printRuntime {
		now = time.Now().UnixNano() / int64(time.Millisecond)
		defer func() {
			newNow := time.Now().UnixNano() / int64(time.Millisecond)
			logger.WithFields(logger.Fields{
				"COMMAND": pathToCommand,
				"RUNTIME": fmt.Sprintf("Runtime: %dms", newNow-now),
			}).Debug("")
		}()
	}

	// Timeout implementation
	cmdContext := context.Background()
	if timeout > 0 {
		var cancel context.CancelFunc
		cmdContext, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(cmdContext, pathToCommand, v...)
	var errBuff bytes.Buffer
	var outBuff bytes.Buffer
	cmd.Stderr = &errBuff
	cmd.Stdout = &outBuff

	err = cmd.Run()
	if err != nil {
		stdout := strings.TrimRight(errBuff.String(), "\r\n")
		return outBuff.String(), err, stdout
	}

	result := strings.TrimRight(outBuff.String(), "\r\n")
	return result, err, ""
}

// RunCommand : runs a command with pipes. Note that all the flags should be directly given to the commands.
func RunCommand(pippedCommand string, printRuntime bool, timeout time.Duration) (string, error, string) {
	return Run("bash", printRuntime, timeout, "-c", pippedCommand)
}
