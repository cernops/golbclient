package checks

import (
	"fmt"
	"os"

	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
)

type NoLogin struct {
}

func (nl NoLogin) Run(a ...interface{}) interface{} {
	// Abort execution if the caller does not fulfill the contract
	if a == nil || len(a) < 2 {
		logger.Error("Wrong number or arguments supplied. Please supply an alias name [string] and " +
			"default value [boolean]")
		return false
	}

	lbaliasNames, ok := a[1].([]string)
	logger.Trace("Supplied alias [%v], default value [%v]", lbaliasNames, a[2])
	if !ok {
		logger.Error("Wrong type given as the alias name, please use the [string] type")
		return false
	}

	noLogin := [2]string{"/etc/noLogin", "/etc/iss.noLogin"}
	isDefault, ok := a[2].(bool)
	if !ok {
		logger.Error("Wrong type given as the default value, please use the [boolean] type")
		return false
	}

	if !isDefault {
		noLogin[1] += fmt.Sprintf(".%s", lbaliasNames[0])
	}

	for _, file := range noLogin {
		_, err := os.Stat(file)

		if err == nil {
			logger.Debug("File [%s] is present", file)
			return false
		}
	}

	logger.Debug("Users are allowed to log in")
	return true
}
