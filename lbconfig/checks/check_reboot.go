package checks

import (
	"os"

	logger "github.com/sirupsen/logrus"
)

type Reboot struct {
}

func (rb Reboot) Run(contextLogger *logger.Entry, args ...interface{}) (int, error) {

	file := "/run/systemd/shutdown/scheduled"
	_, err := os.Stat(file)

	if err == nil {
		contextLogger.Errorf("The machine is scheduled for reboot (see %s)", file)
		return -1, nil
	}
	contextLogger.Debug("The machine is not scheduled for reboot")
	return 1, nil
}
