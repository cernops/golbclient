package checks

import (
	"gitlab.cern.ch/lb-experts/golbclient/utils/logger"
	"syscall"
)

const ACCEPTABLE_BLOCK_RATE = 0.90
const ACCEPTABLE_INODE_RATE = 0.95

type TmpFull struct {
}

func (tmpFull TmpFull) Run(...interface{}) interface{} {
	var stat syscall.Statfs_t
	syscall.Statfs("/tmp", &stat)
	blockLevel := 1 - (float64(stat.Bavail) / float64(stat.Blocks))
	inodeLevel := 1 - (float64(stat.Ffree) / float64(stat.Files))

	logger.Debug("Blocks occupancy: %.2f%% inodes occupancy: %.2f%%", blockLevel*100, inodeLevel*100)
	return ((blockLevel < ACCEPTABLE_BLOCK_RATE) && (inodeLevel < ACCEPTABLE_INODE_RATE))
}
