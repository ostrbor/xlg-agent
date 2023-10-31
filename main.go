package xlg_agent

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"
)

func main() {
	// each line in /etc/xlg-agent is a path to a directory with log files
	//conf, err := os.ReadFile("/etc/xlg-agent.conf")
	conf, err := io.ReadAll(os.Stdin)
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
		dirFiles, err := os.ReadDir(string(dir))
		if err != nil {
			panic(err)
		}

		var logFiles []string
		for _, file := range dirFiles {
			name := file.Name()
			if !strings.HasSuffix(name, ".log") {
				continue
			}
			logFiles = append(logFiles, name)
		}
		sort.Strings(logFiles)

		for _, file := range logFiles {
			err = processFile(file, send)
			if err != nil {
				panic(err)
			}
		}
		time.Sleep(5 * time.Second)
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
