package xlg_agent

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"time"
)

func main() {
	// each line in /etc/xlg-agent is a path to a directory with log files
	conf, err := os.ReadFile("/etc/xlg-agent.conf")
	if err != nil {
		panic(err)
	}

	dirs := bytes.Split(conf, []byte("\n"))
	for _, dir := range dirs {
		// read content in dir
		// sort all files
		// find file to send logs
		// send logs
		// update file
		files, err := os.ReadDir(string(dir))
		if err != nil {
			panic(err)
		}

		var filenames []string
		for _, file := range files {
			filenames = append(filenames, file.Name())
		}
		sort.Strings(filenames)

		for _, file := range filenames {
			err = handleFile(file)
			if err != nil {
				panic(err)
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func handleFile(file string) error {
	fd, err := os.OpenFile(file, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer fd.Close()

	f := new(bytes.Buffer)
	_, err = f.ReadFrom(fd)
	if err != nil {
		return err
	}

	delim := '\n'
	lineStart := 0
	lineEnd := 0

	for {
		line, err := f.ReadBytes(byte(delim))
		if err != nil {
			if err.Error() == "EOF" {
				return nil
			}
			return err
		}

		lineEnd += len(line) + len(string(delim))
		if line[0] == '+' {
			lineStart += len(line) + len(string(delim))
			continue
		}

		json := line[1:]
		err = send(json)
		if err != nil {
			return err
		}

		f.WriteByte('+')

	}
}

func send(json []byte) error {
	fmt.Println(string(json))
	return nil
}

func processFile(filePath string, processLine func([]byte) error) error {
	fd, err := os.OpenFile(filePath, os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	defer fd.Close()

	scanner := bufio.NewScanner(fd)
	lineStart := 0
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) > 0 && line[0] == '-' {
			if err := processLine(line); err != nil {
				return err
			}
			// mark line as successfully processed
			_, err = fd.WriteAt([]byte{'+'}, int64(lineStart))
			if err != nil {
				return err
			}
		}
		lineStart += len(line) + len(string('\n'))
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func replaceMinusWithPlus(filename string) error {
	file, err := os.OpenFile(filename, os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()
	bufio.NewReader(file)

	// Create a buffer to read and modify the content.
	buffer := make([]byte, 1)
	for {
		n, err := file.Read(buffer)
		if err != nil {
			break // End of file reached.
		}
		if n != 1 {
			return nil
		}

		// Check if the read byte is '-'.
		if buffer[0] == '-' {
			// Seek back to the position of the '-' character.
			_, err := file.Seek(-1, io.SeekCurrent)
			if err != nil {
				return err
			}

			// Replace '-' with '+' and write the modified byte.
			_, err = file.Write([]byte{byte('+')})
			if err != nil {
				return err
			}

			// Move the file pointer one byte forward.
			_, err = file.Seek(1, io.SeekCurrent)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func readByteByByte(pathname string) {
	f, err := os.Open(pathname)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	b := make([]byte, 1)
	for {
		n, err := f.Read(b)
		if err != nil {
			break
		}
		if n != 1 {
			break
		}
	}
}

func readWithBuffer(pathname string) {
	f, err := os.Open(pathname)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	r := bufio.NewReader(f)
	for {
		_, err := r.ReadBytes('\n')
		if err != nil {
			break
		}
	}
}

func readWithScanner(pathname string) {
	f, err := os.Open(pathname)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		_ = scanner.Bytes()
	}
}
