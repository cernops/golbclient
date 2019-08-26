package checks

import logger "github.com/sirupsen/logrus"

type CheckAttribute struct {
}

func (checkAttribute CheckAttribute) Run(contextLogger *logger.Entry, args ...interface{}) (int, error) {
	// This will be used later on for the default load
	contextLogger.Trace("This check always returns true. It will be used to calculate the default load")
	return 1, nil
}
