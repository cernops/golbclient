package checks

import (
	"fmt"
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
	"os"
)

type NoLogin struct {
	
}
func (nl NoLogin) Run(a ...interface{}) interface{} {
	lbaliasName := a[1]

	nologin := [2]string{"/etc/nologin", "/etc/iss.nologin"}

	if lbaliasName != "" {
		nologin[1] += fmt.Sprintf(".%s", lbaliasName)
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
