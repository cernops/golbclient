package checks

import (
	"os"

	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
)

const afsDir = "/afs/cern.ch/user/"

// AFS : struct that represent the AFS : CLI implementation
type AFS struct{}

// Run : Runs the AFS : CLI implementation function
func (afs AFS) Run(args ...interface{}) (interface{}, error) {
	logger.Debug("Checking the that AFS directory is accessible...")
	if _, err := os.Stat(afsDir); os.IsNotExist(err) {
		logger.Error("AFS directory does not exist")
		return false, err

	}

	logger.Trace("Successfully accessed the AFS directory")

	return true, nil
}
