package checks

import "gitlab.cern.ch/lb-experts/golbclient/helpers/logger"

type CheckAttribute struct {
}

func (checkAttribute CheckAttribute) Run(...interface{}) (int, error) {
	// This will be used later on for the default load
	logger.Trace("This check always returns true. It will be used to calculate the default load")
	return 1, nil
}
