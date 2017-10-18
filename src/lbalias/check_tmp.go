package lbalias

import ("syscall")

func checkTmpFull(lbalias *LBalias, line string) bool {
	var stat syscall.Statfs_t

	syscall.Statfs("/tmp", &stat)
	blockLevel := 1 - (float64(stat.Bavail) / float64(stat.Blocks))
	inodeLevel := 1 - (float64(stat.Ffree) / float64(stat.Files))
	lbalias.DebugMessage("[check_tmpfull] blocks occupancy: %.2f%% inodes occupancy: %.2f%%\n", blockLevel*100, inodeLevel*100)

	return ((blockLevel > ACCEPTABLE_BLOCK_RATE) || (inodeLevel > ACCEPTABLE_INODE_RATE))
}
