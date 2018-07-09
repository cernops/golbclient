package filehandler

import (
	"bufio"
	"lbalias/utils/logger"
	"os"
)

// ReadAllLinesFromFile : Reads all lines from a file into a string array
func ReadAllLinesFromFile(path string) (lines []string, err error) {
	logger.Trace("Attempting to read file [%s]", path)
	file, err := os.Open(path)
	if err != nil {
		logger.Error("Error when attempting to read the file [%s]. Error [%s]", path, err.Error())
		return nil, err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	err = scanner.Err()
	return lines, err
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
