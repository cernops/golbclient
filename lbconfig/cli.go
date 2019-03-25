package lbconfig

// CLI : generic interface for all the functions that run a CLI command
type CLI interface {
	Run(...interface{}) (interface{}, error)
}
