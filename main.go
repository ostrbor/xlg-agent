package xlg_agent

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"github.com/ostrbor/xlg"
	"io"
	"net/http"
	"os"
	"path"
	"slices"
	"time"
)

var (
	collectorUrl   = os.Getenv("COLLECTOR_URL")
	collectorToken = os.Getenv("COLLECTOR_TOKEN")

	// todo memory leak as cache grows indefinitely
	cache = make(map[string]int64)

	rootDir = flag.String("dir", "", "root directory for log subdirectories")
)

const (
	sentMark = '+'
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

	for {
		// todo each time iteration of map has different order
		for dir, files := range searchLogs(*rootDir) {
			for _, f := range files {
				handleFile(path.Join(dir, f))
				// todo log rotation, move old file to archive
			}
		}
		time.Sleep(5 * time.Second)
	}
}

// res key is directory path, value is slice of log file names
func searchLogs(rootDir string) (res map[string][]string) {
	dirNames, err := subDirs(rootDir)
	if err != nil {
		panic(err)
	}
	res = make(map[string][]string)
	for _, name := range dirNames {
		dirpath := path.Join(rootDir, name)
		entries, err := os.ReadDir(dirpath)
		if err != nil {
			panic(err)
		}
		// filenames and match both iterate same entries, not efficient but readable
		fnames := filenames(entries)
		logFiles := match(fnames, xlg.FileFormat)
		slices.Sort(logFiles)
		res[dirpath] = logFiles
	}
	return res
}

func filenames(entries []os.DirEntry) (res []string) {
	res = make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		res = append(res, entry.Name())
	}
	return res
}

func match(filenames []string, format string) (res []string) {
	res = make([]string, 0, len(filenames))
	for _, n := range filenames {
		if !isLogFile(n, format) {
			continue
		}
		res = append(res, n)
	}
	return res
}

func subDirs(dir string) (names []string, err error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	names = make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		names = append(names, entry.Name())
	}
	return names, nil
}

// handleFile logs from dirs in rootDir to collector
func handleFile(filepath string) {
	// read content in rootDir
	// sort all files
	// find file to send logs
	// send logs
	// update file
	fd, err := os.OpenFile(filepath, os.O_RDWR, os.ModePerm)
	if err != nil {
		return
	}
	defer func() {
		if closeErr := fd.Close(); closeErr != nil {
			// log err
			panic(closeErr)
		}
	}()
	nextOffset := cache[filepath]
	if !updated(fd, nextOffset) {
		return
	}
	end, err := handleLines(fd, nextOffset, send)
	if err != nil {
		panic(err)
	}
	cache[filepath] = end
}

func updated(f *os.File, nextOf int64) bool {
	s, err := f.Stat()
	if err != nil {
		panic(err)
	}
	if nextOf == s.Size() {
		return false
	}
	return true
}

func filterLogs(entries []os.DirEntry) []string {
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		n := entry.Name()
		if !isLogFile(n, xlg.FileFormat) {
			continue
		}
		names = append(names, n)
	}
	return names
}

func isLogFile(fileName, fileFormat string) bool {
	_, err := time.Parse(fileFormat, fileName)
	if err != nil {
		return false
	}
	return true
}

func send(json []byte) error {
	c := http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(http.MethodPost, collectorUrl, bytes.NewReader(json))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+collectorToken)
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	// todo handle err
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

// handleLines reads a file line by line, processes logs if a line starts with the '-' character, and marks the line as sent by replacing '-' with '+'.
// It returns the offset of the next byte to be processed, which corresponds to the beginning of the line following the last processed line.
// To resume reading from a specific offset and avoid starting from the beginning each time, it accepts the 'resumeOffset' as an argument.
func handleLines(fd *os.File, resumeOffset int64, send func([]byte) error) (nextLineOffset int64, err error) {
	if _, err = fd.Seek(resumeOffset, io.SeekStart); err != nil {
		return
	}
	// prefer bufio.Reader over bufio.Scanner because Scanner returns last line even if it doesn't end with newline.
	rd := bufio.NewReader(fd)
	nextLineOffset = resumeOffset
	for {
		line, err := rd.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nextLineOffset, err
			}
		}
		if len(line) > 0 && line[0] == xlg.NotSentMark {
			if err = send(line); err != nil {
				return nextLineOffset, err
			}
			_, err = fd.WriteAt([]byte{sentMark}, nextLineOffset)
			if err != nil {
				return nextLineOffset, err
			}
		}
		nextLineOffset += int64(len(line))
	}

	return nextLineOffset, nil
}
