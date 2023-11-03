package xlg_agent

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

var (
	collectorUrl   = os.Getenv("COLLECTOR_URL")
	collectorToken = os.Getenv("COLLECTOR_TOKEN")
)

func main() {
	// each line in /etc/xlg-agent is a path to a directory with log files
	//conf, err := os.ReadFile("/etc/xlg-agent.conf")

	// todo print env vars to stdout
	// todo print config dirs to stdout
	conf, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	dirs := strings.Split(string(conf), "\n")
	for _, dir := range dirs {
		if _, err := os.Stat(dir); err != nil {
			if os.IsNotExist(err) {
				// todo format err
				panic(err)
			} else {
				// eg permission err
				// todo format err
				panic(err)
			}
		}
	}

	for {
		for _, dir := range dirs {
			forward(dir)
		}
		time.Sleep(5 * time.Second)
	}
}

// forward logs from dir to collector
func forward(dir string) {
	// read content in dir
	// sort all files
	// find file to send logs
	// send logs
	// update file
	files, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	var logFiles []string
	for _, file := range files {
		name := file.Name()
		if !strings.HasSuffix(name, ".log") {
			continue
		}
		logFiles = append(logFiles, name)
	}
	sort.Strings(logFiles)

	for _, file := range logFiles {
		err = process(file, send)
		if err != nil {
			panic(err)
		}
	}
}

func send(json []byte) error {
	c := http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(http.MethodPost, collectorUrl, bytes.NewReader(json))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+collectorToken)
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func process(filePath string, handleLine func([]byte) error) error {
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
			if err := handleLine(line); err != nil {
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
