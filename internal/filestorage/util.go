package filestorage

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// Read the fileWrite backward, byte by byte (no need to set a buffer size)
// until finding the beginning of a line or the beginning of the fileWrite.
func getLastLineOfTheFile(fileHandle *os.File) (string, error) {
	line := ""
	var cursor int64 = 0
	stat, err := fileHandle.Stat()
	if err != nil {
		return line, err
	}
	var filesize int64 = stat.Size()
	if filesize == 0 {
		return "", nil
	}
	for {
		cursor--
		_, err = fileHandle.Seek(cursor, io.SeekEnd)
		if err != nil {
			return line, err
		}

		char := make([]byte, 1)
		_, err = fileHandle.Read(char)
		if err != nil {
			return line, err
		}

		// stop if we find a line
		if cursor != -1 && (char[0] == 10 || char[0] == 13) {
			break
		}

		line = fmt.Sprintf("%s%s", string(char), line)

		if cursor == -filesize { // stop if we are at the begining
			break
		}
	}

	line = strings.Trim(line, "\x00")
	return line, err
}
