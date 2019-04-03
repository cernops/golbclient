package checks

import (
	"fmt"
	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
	"gitlab.cern.ch/lb-experts/golbclient/lbconfig/utils/runner"
)

const afsDir = "/afs/cern.ch/user/"

// AFS : struct that represent the AFS : CLI implementation
type AFS struct {}

// Run : Runs the AFS : CLI implementation function
func (afs AFS) Run(args ...interface{}) (interface{}, error) {
	logger.Debug("Checking the that AFS directory is accessible...")
	output, err := runner.RunCommand(fmt.Sprintf("ls -al %v", afsDir), true, 0)
	if err != nil  {
		return -1, err
	}
	logger.Trace("Successfully accessed the AFS directory with output [%v]", output)
	return 1, nil
}