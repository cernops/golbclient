package checks

import (
	"io/ioutil"
	"lbalias/utils/logger"
)

const afsDir = "/afs/cern.ch/user/"

// AFS : struct that represent the AFS : CLI implementation
type AFS struct {
	code int
}

// Code : Returns the code provided during initialization
func (afs AFS) Code() int {
	return afs.code
}

// SafeRun : Runs the default Run function but an error is returned. The error will be 'nil' if no error ocurred.
func (afs AFS) SafeRun(args ...interface{}) (interface{}, error) {
	var err error
	res := afs.Run(args)
	defer func() {
		if r, isErr := recover().(error); r != nil && isErr {
			err = r
		}
	}()
	return res, err
}

// Run : Runs the AFS : CLI implementation function
func (afs AFS) Run(args ...interface{}) interface{} {
	_, err := ioutil.ReadDir(afsDir)
	if err != nil {
		logger.Error("The following error was detected when checking AFS [%s]", err.Error())
		return true
	}
	return false
}
