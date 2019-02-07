package checks

import (
	"io/ioutil"
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
)

const afsDir = "/afs/cern.ch/user/"

// AFS : struct that represent the AFS : CLI implementation
type AFS struct {
}


// Run : Runs the AFS : CLI implementation function
func (afs AFS) Run(args ...interface{}) interface{} {
	_, err := ioutil.ReadDir(afsDir)
	if err != nil {
		logger.Error("The following error was detected when checking AFS [%s]", err.Error())
		return false
	}
	return true
}
