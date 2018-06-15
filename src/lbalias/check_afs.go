package lbalias

import (
	"fmt"
	"io/ioutil"
)

const AFS_DIR = "/afs/cern.ch/user/"

func checkAFS(lbalias *LBalias, line string) interface{} {
	_, err := ioutil.ReadDir(AFS_DIR)
	if err != nil {
		fmt.Println("Error checking afs", err)
		return true
	}

	return false
}
