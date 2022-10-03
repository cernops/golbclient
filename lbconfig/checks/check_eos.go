package checks

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"

	logger "github.com/sirupsen/logrus"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/utils/runner"
)

const mountsProcFile = "/proc/mounts"
const testCmdBase = "/usr/bin/eosxd get eos.mgmurl "

type EOS struct{}

func (eos EOS) Run(contextLogger *logger.Entry, args ...interface{}) (int, error) {

	f, err := os.Open(mountsProcFile)
	if err != nil {
		return -1, err
	}
	defer func() { err = f.Close() }()
	return DoEOSCheck(f, testCmdBase, contextLogger)
}

func DoEOSCheck(filehandle io.Reader, baseCmd string, contextLogger *logger.Entry) (int, error) {
	scanner := bufio.NewScanner(filehandle)
	contextLogger.Trace("Checking the mount entries...")
	eosEntry, _ := regexp.Compile("^[0-9A-Za-z_-]+ (/eos/[0-9A-Za-z_-]+|/eos/[0-9A-Za-z_-]+/[0-9A-Za-z_-]+) fuse .*")
	for scanner.Scan() {
		line := scanner.Text()
		match := eosEntry.FindStringSubmatch(line)
		if len(match) > 0 {
			usrCmd := baseCmd + match[1]
			contextLogger.Tracef("Attempting to run command [%s]", usrCmd)
			out, err, stderr := runner.RunCommand(usrCmd, true, 0)
			if err != nil {
				contextLogger.Errorf("The command [%s] failed. Error [%v] Stderr[%v]", usrCmd, err, stderr)
				if stderr != "" {
					if strings.Contains(stderr, "Transport endpoint is not connected") || strings.Contains(stderr, "Operation not supported") || strings.Contains(stderr, "Input/output error") {
						return -1, nil
					}
				}
				if exitErr, ok := err.(*exec.ExitError); ok {
					if exitErr.Exited() {
						// If command exited with status != 0
						if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
							if status.ExitStatus() == CommandNotFoundCode || status.ExitStatus() == CommandNotExecutable {
								// If command not found or not executable
								return -1, err
							}
						}
					} else { // Not Exited()
						return -1, err
					}
				} else { // Not exec.ExitError()
					return -1, err
				}
			}
			contextLogger.Debugf("Command output [%s]", out)
		}
	}
	return 1, nil
}
