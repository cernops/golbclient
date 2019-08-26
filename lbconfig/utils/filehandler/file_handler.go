package filehandler

import (
	"io/ioutil"
	"strings"
)

// ReadAllLinesFromFile : Reads all lines from a file into a string array
func ReadAllLinesFromFile(path string) (lines []string, err error) {
	c, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return strings.Split(string(c), "\n"), nil
}

// ReadFirstLineFromFile : Reads the first line from a file
func ReadFirstLineFromFile(path string) (line string, err error) {
	lines, err := ReadAllLinesFromFile(path)
	if err != nil {
		return line, err
	}
	line = lines[0]
	return line, err
}
