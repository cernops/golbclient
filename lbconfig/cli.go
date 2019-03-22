package lbconfig

// CLI : generic interface for all the functions that run a CLI command
type CLI interface {
	//	Code() int
	Run(...interface{}) interface{}
	//SafeRun(...interface{}) (interface{}, error)
}
