package xlg_agent

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"testing"
)

const (
	filename   = "sample.txt"
	start_data = `- {"msg": "test"}
+ {"msg": "test2"}
- {"msg": "test3"}
`
)

func printContents(filepath string) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
}

func TestX2(t *testing.T) {
	dir := t.TempDir()
	filepath := path.Join(dir, filename)
	err := os.WriteFile(filepath, []byte(start_data), 0644)
	if err != nil {
		panic(err)
	}

	printContents(filepath)

	f, err := os.OpenFile(filepath, os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 && line[0] == '-' {
			if _, err := f.WriteAt([]byte("+"), 0); err != nil {
				panic(err)
			}
		}
	}
	fmt.Println()

	printContents(filepath)
}
