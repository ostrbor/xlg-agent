package xlg_agent

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

var (
	collectorUrl   = os.Getenv("COLLECTOR_URL")
	collectorToken = os.Getenv("COLLECTOR_TOKEN")
	cache          = make(map[string]int64)
	rootDir        = flag.String("dir", "", "root directory for log subdirectories")
)

// /rootDir/logDir/logFile
func main() {
	flag.Parse()
	if *rootDir == "" {
		panic("dir flag is required")
	}
	s, err := os.Stat(*rootDir)
	if err != nil {
		if os.IsNotExist(err) {
			// todo format err
			panic(err)
		} else {
			// eg permission err
			// todo format err
			panic(err)
		}
	}
	if !s.IsDir() {
		panic("not a directory")
	}
	// todo print env vars to stdout

	rootEntries, err := os.ReadDir(*rootDir)
	if err != nil {
		panic(err)
	}

	for {
		for _, entry := range rootEntries {
			if !entry.IsDir() {
				continue
			}
			p := path.Join(*rootDir, entry.Name())
			forward(p)
		}
		time.Sleep(5 * time.Second)
	}
}

// forward logs from dirs in rootDir to collector
func forward(dir string) {
	// read content in rootDir
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
		filepath := path.Join(dir, file)
		start := cache[filepath]
		end, err := process(filepath, start, send)
		if err != nil {
			panic(err)
		} else {
			cache[filepath] = end
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

// todo use start offset to compare with file size to avoid reading file
func process(filePath string, start int64, handleLine func([]byte) error) (offset int64, err error) {
	fd, err := os.OpenFile(filePath, os.O_RDWR, os.ModePerm)
	if err != nil {
		return
	}
	defer func() {
		if closeErr := fd.Close(); closeErr != nil {
			// log err
			panic(closeErr)
		}
	}()
	if _, err = fd.Seek(start, io.SeekStart); err != nil {
		return
	}

	// prefer bufio.Reader over bufio.Scanner because Scanner returns last line even if it doesn't end with newline.
	rd := bufio.NewReader(fd)
	offset = start
	for {
		line, err := rd.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return 0, err
			}
		}
		if len(line) > 0 && line[0] == '-' {
			if err = handleLine(line); err != nil {
				return 0, err
			}
			// mark line as successfully processed
			_, err = fd.WriteAt([]byte{'+'}, offset)
			if err != nil {
				return 0, err
			}
		}
		offset += int64(len(line))
	}

	return offset, nil
}
