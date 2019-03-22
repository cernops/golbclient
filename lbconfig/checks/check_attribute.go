package checks

import (
	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
)

type CheckAttribute struct {
}

func (checkAttribute CheckAttribute) Run(args ...interface{}) interface{} {
	// Log
	logger.Debug("This will be used later on for the default load")

	return true
}
