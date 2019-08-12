package checks

import (
	"syscall"

	logger "github.com/sirupsen/logrus"
)

const (
	acceptableBlockRate = 0.90
	acceptableINodeRate = 0.95
)

type TmpFull struct {
}

func (tmpFull TmpFull) Run(...interface{}) (int, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs("/tmp", &stat)
	if err != nil {
		logger.Error("Error occured checking the /tmp directory")
		return -1, err
	}
	blockLevel := 1 - (float64(stat.Bavail) / float64(stat.Blocks))
	iNodeLevel := 1 - (float64(stat.Ffree) / float64(stat.Files))

	logger.Debugf("Blocks occupancy [%.2f%%], inodes occupancy [%.2f%%]", blockLevel*100, iNodeLevel*100)
	if (blockLevel < acceptableBlockRate) && (iNodeLevel < acceptableINodeRate) {
		return 1, nil
	}
	logger.Errorf("The tmp directory is full: [%.2f%%] (bigger than [%.2f%%])", blockLevel*100, acceptableBlockRate*100)
	return -1, nil

}
