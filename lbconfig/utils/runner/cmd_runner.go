package runner

import (
	"bytes"
	"context"
	logger "github.com/sirupsen/logrus"
	"os/exec"
	"strings"
	"time"
)

// Run : runs a command with the given arguments if this is available. Returns a tuple of he output of the command in the desired format and an error
func Run(pathToCommand string, printRuntime bool, timeout time.Duration, v ...string) (output string, err error) {
	var now int64
	if printRuntime {
		now = time.Now().UnixNano() / int64(time.Millisecond)
		defer func() {
			newNow := time.Now().UnixNano() / int64(time.Millisecond)
			logger.Debugf("Runtime: %dms", newNow-now)
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
		return outBuff.String(), err
	}

	result := strings.TrimRight(outBuff.String(), "\r\n")
	return result, err
}

// RunCommand : runs a command with pipes. Note that all the flags should be directly given to the commands.
func RunCommand(pippedCommand string, printRuntime bool, timeout time.Duration) (string, error) {
	return Run("bash", printRuntime, timeout,"-c", pippedCommand)
}
