package checks

import (
	"fmt"
	"lbalias/utils/logger"
	"os"
)

type NoLogin struct {
	code int
}

func (nologin NoLogin) Code() int {
	return nologin.code
}

func (nologin NoLogin) SetCode(ncode int) {
	nologin.code = ncode
}

func (nl NoLogin) Run(a ...interface{}) interface{} {
	lbalias := a[1].(*LBalias)

	if lbalias.NoLogin {
		return true
	}
	nologin := [2]string{"/etc/nologin", "/etc/iss.nologin"}

	if lbalias.Name != "" {
		nologin[1] += fmt.Sprintf(".%s", lbalias.Name)
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
