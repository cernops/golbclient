package lbalias

import (
	"encoding/binary"
	"os"
)

type ExitStatus struct {
	X__e_termination int16
	X__e_exit        int16
}

type TimeVal struct {
	Sec  int32
	Usec int32
}

type Utmp struct {
	Type              int16
	Pad_cgo_0         [2]byte
	Pid               int32
	Line              [32]byte
	Id                [4]byte
	User              [32]byte
	Host              [256]byte
	Exit              ExitStatus
	Session           int32
	Tv                TimeVal
	Addr_v6           [4]int32
	X__glibc_reserved [20]byte
}

const (
	Empty        = 0x0
	RunLevel     = 0x1
	BootTime     = 0x2
	NewTime      = 0x3
	OldTime      = 0x4
	InitProcess  = 0x5
	LoginProcess = 0x6
	UserProcess  = 0x7
	DeadProcess  = 0x8
	Accounting   = 0x9
	Unknown      = 0x0
)

func readAllUtmpEntries() ([]Utmp, error) {
	var uSlice []Utmp
	var uEntry Utmp

	file, err := os.Open("/var/run/utmp")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	for {
		err := binary.Read(file, binary.LittleEndian, &uEntry)
		if err != nil {
			break
		}
		uSlice = append(uSlice, uEntry)
	}
	return uSlice, err
}
