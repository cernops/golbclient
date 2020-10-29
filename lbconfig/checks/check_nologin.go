package checks

import (
	"fmt"
	"os"

	logger "github.com/sirupsen/logrus"
)

type NoLogin struct {
}

func (nl NoLogin) Run(contextLogger *logger.Entry, args ...interface{}) (int, error) {
	// Abort execution if the caller does not fulfill the contract
	if args == nil || len(args) < 2 {
		return -1, fmt.Errorf("wrong number or arguments supplied. Please supply an alias name [string] and " +
			"default value [boolean]")
	}

	lbaliasNames, ok := args[1].([]string)
	contextLogger.Tracef("Supplied alias [%v], default value [%v]", lbaliasNames, args[2])
	if !ok {
		return -1, fmt.Errorf("wrong type given as the alias name, please use the [string] type")
	}

	noLogin := [2]string{"/etc/nologin", "/etc/iss.nologin"}
	isDefault, ok := args[2].(bool)
	if !ok {
		return -1, fmt.Errorf("wrong type given as the default value, please use the [boolean] type")
	}

	if !isDefault {
		noLogin[1] += fmt.Sprintf(".%s", lbaliasNames[0])
	}

	for _, file := range noLogin {
		_, err := os.Stat(file)

		if err == nil {
			contextLogger.Errorf("File [%s] is present", file)
			return -1, nil
		}
	}

	contextLogger.Debug("Users are allowed to log in")
	return 1, nil
}
