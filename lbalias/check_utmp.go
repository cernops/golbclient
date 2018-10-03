package lbalias

import (
	"encoding/binary"
	"lbalias/utils/logger"
	"os"
)

// utmpFilePath : path used to access [utmp] type metrics
const utmpFilePath = "/var/run/utmp"

// ExitStatus : struct used to represent the exit status detected in the [utmp] entry
type ExitStatus struct {
	X__e_termination int16
	X__e_exit        int16
}

// TimeVal : struct used to represent the time value used in the [utmp] entry
type TimeVal struct {
	Sec  int32
	Usec int32
}

// Type : type of process detected in the [utmp] entry
type Type int16

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

// Supported Types
const (
	Empty Type = iota
	RunLevel
	BootTime
	NewTime
	OldTime
	InitProcess
	LoginProcess
	UserProcess
	DeadProcess
	Accounting
	Unknown
)

// readAllUtmpEntries : given an array of optional Type objects, the [utmp] file will be read and filtered. Returns the array of entries read from the [utmp] file that comply with the imposed desired Type objects
func readAllUtmpEntries(types ...Type) ([]Utmp, error) {
	var uSlice []Utmp
	var uEntry Utmp

	file, err := os.Open(utmpFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	for {
		err := binary.Read(file, binary.LittleEndian, &uEntry)
		if err != nil {
			logger.Error("An error occurred when attempting to read the [utmp] file at [%s]. Error [%s]", utmpFilePath, err.Error())
			break
		}
		if len(types) > 0 {
			for _, t := range types {
				if uEntry.Type == int16(t) {
					uSlice = append(uSlice, uEntry)
					break
				}
			}
		} else {
			uSlice = append(uSlice, uEntry)
		}
	}
	return uSlice, err
}
