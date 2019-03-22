package checks

import (
	"syscall"

	"gitlab.cern.ch/lb-experts/golbclient/helpers/logger"
)

const (
	acceptableBlockRate = 0.90
	acceptableINodeRate = 0.95
)

type TmpFull struct {
}

func (tmpFull TmpFull) Run(...interface{}) interface{} {
	var stat syscall.Statfs_t
	err := syscall.Statfs("/tmp", &stat)
	if err != nil {
		logger.Error("The [/tmp] directory is not accessible. Error [%s]", err.Error())
		return false
	}
	blockLevel := 1 - (float64(stat.Bavail) / float64(stat.Blocks))
	iNodeLevel := 1 - (float64(stat.Ffree) / float64(stat.Files))

	logger.Debug("Blocks occupancy [%.2f%%], inodes occupancy [%.2f%%]", blockLevel*100, iNodeLevel*100)
	return (blockLevel < acceptableBlockRate) && (iNodeLevel < acceptableINodeRate)
}
