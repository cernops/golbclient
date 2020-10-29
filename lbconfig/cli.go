//+build linux darwin

package lbconfig

import logger "github.com/sirupsen/logrus"

// CLI : generic interface for all the functions that run a CLI command
type CLI interface {
	Run(contextLogger *logger.Entry, args ...interface{}) (int, error)
}
