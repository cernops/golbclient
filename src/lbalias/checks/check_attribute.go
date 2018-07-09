package checks

import (
	"lbalias/utils/logger"
)

type CheckAttribute struct {
	code int
	name string
}

func (checkAttribute CheckAttribute) Code() int {
	return checkAttribute.code
}

func (checkAttribute CheckAttribute) Run(args ...interface{}) interface{} {
	lbalias := args[1].(*LBalias)

	if lbalias.CheckAttributes == nil {
		lbalias.CheckAttributes = map[string]bool{}
	}
	// Log
	logger.Debug("Checking the attribute [%s]", checkAttribute.name)
	lbalias.CheckAttributes[checkAttribute.name] = true

	return true
}
