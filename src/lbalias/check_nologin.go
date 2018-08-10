package lbalias

import (
	"os"
)

func checkNoLogin(lbalias *LBalias, line string) bool {
	if lbalias.NoLogin {
		return false
	}
	nologin := [2]string{"/etc/nologin", "/etc/iss.nologin"}

	if lbalias.Name != "" {
		nologin[1] += "." + lbalias.Name
	}
	for _, file := range nologin {
		_, err := os.Stat(file)

		if err == nil {
			lbalias.DebugMessage("[check_nologin] ", file, " present")
			return true
		}
	}
	lbalias.DebugMessage("[check_nologin] users allowed to log in")
	return false

}
