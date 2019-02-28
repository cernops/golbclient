package checks

import (
	"fmt"
	"os"

	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
)

type NoLogin struct {
}

func (nl NoLogin) Run(a ...interface{}) interface{} {
	lbaliasNames, ok := a[1].([]string)
	logger.Trace("The alias name is %v and default %v", lbaliasNames, a[2])
	if !ok {
		logger.Error("Wrong type given as the alias name")
		return false
	}

	nologin := [2]string{"/etc/nologin", "/etc/iss.nologin"}

	if a[2].(bool) != true {
		nologin[1] += fmt.Sprintf(".%s", lbaliasNames[0])
	}
	for _, file := range nologin {
		_, err := os.Stat(file)

		if err == nil {
			logger.Debug("File [%s] is present", file)
			return false
		}
	}

	logger.Debug("Users are allowed to log in")
	return true
}
