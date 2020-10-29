package checks

import (
	"os"

	logger "github.com/sirupsen/logrus"
)

const afsDir = "/afs/cern.ch/user/"

// AFS : struct that represent the AFS : CLI implementation
type AFS struct{}

// Run : Runs the AFS : CLI implementation function
func (afs AFS) Run(contextLogger *logger.Entry, args ...interface{}) (int, error) {
	contextLogger.Debug("Checking the that AFS directory is accessible...")
	if _, err := os.Stat(afsDir); os.IsNotExist(err) {
		contextLogger.Error("AFS directory does not exist")
		return -1, nil
	}

	contextLogger.Trace("Successfully accessed the AFS directory")
	return 1, nil
}
