package filehandler

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// ReadAllLinesFromFile : Reads all lines from a file into a string array
func ReadAllLinesFromFile(path string) (lines []string, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer func() {err = file.Close()}()
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

// ReadAllLinesFromFileAsString : Reads all the lines from a file into a single string joined by the given separator 
func ReadAllLinesFromFileAsString(path string, separator string) (content string, err error) {
	lines, err := ReadAllLinesFromFile(path)
	if err != nil {
		return content, err
	}
	content = strings.Join(lines, separator)
	return content, err
}

// CreateFileInDir : Creates a file with all the required parent directories with the given permissions. If an issue is
// detected during this process, an error will be returned. Otherwise, a pointer to @see os.File the instance is
// returned
func CreateFileInDir(file string, mode os.FileMode) (fHandle *os.File, err error) {
	_, err = os.Stat(file)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(filepath.Dir(file), mode); err != nil {
			return fHandle, err
		}
	}
	return
}

