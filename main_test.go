package xlg_agent

import (
	"bufio"
	"bytes"
	"math/rand"
	"os"
	"path"
	"testing"
	"time"
)

func TestX(t *testing.T) {
	x := []byte("a\nb\nc")
	lines := bytes.Split(x, []byte("\n"))

	for _, l := range lines {
		println(string(l))
	}
}

func TestCmpRead(t *testing.T) {
	dir := t.TempDir()
	filepath := path.Join(dir, "/test.txt")
	createTestFile(filepath)
	start := time.Now()
	readByteByByte(filepath)
	println("readByteByByte", time.Since(start).Milliseconds())

	start = time.Now()
	readWithBuffer(filepath)
	println("readWithBuffer", time.Since(start).Milliseconds())
}

func createTestFile(fileName string) {
	file, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Set the desired file size (10 megabytes)
	fileSize := 10 * 1024 * 1024

	// Create a random source for generating content
	rand.Seed(time.Now().UnixNano())

	// Create a buffer writer for improved performance
	writer := bufio.NewWriter(file)

	// Keep writing lines until the file reaches the desired size
	for size(file) < int64(fileSize) {
		// Generate a random line or use any content generation logic you prefer
		line := generateRandomLine()

		// Write the line to the file
		_, err := writer.WriteString(line)
		if err != nil {
			panic(err)
		}
	}

	// Flush the writer and ensure all data is written to the file
	writer.Flush()

}

func generateRandomLine() string {
	// You can modify this function to generate your desired content for each line
	// For example, here, we are generating a random line of characters
	lineLength := 100 // Adjust the desired line length
	letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	line := make([]byte, lineLength)
	for i := range line {
		line[i] = letters[rand.Intn(len(letters))]
	}

	return string(line) + "\n"
}

func size(f *os.File) int64 {
	info, err := f.Stat()
	if err != nil {
		panic(err)
	}
	return info.Size()
}
