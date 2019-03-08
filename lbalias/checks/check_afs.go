package checks

import (
	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
	"io/ioutil"
)

const afsDir = "/afs/cern.ch/user/"

// AFS : struct that represent the AFS : CLI implementation
type AFS struct {
}

// Run : Runs the AFS : CLI implementation function
func (afs AFS) Run(args ...interface{}) interface{} {
	_, err := ioutil.ReadDir(afsDir)
	if err != nil {
		logger.Info("The AFS directory is not accessible. Error [%s]", err.Error())
		return false
	}
	return true
}
