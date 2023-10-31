package xlg_agent

import (
	"os"
	"testing"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func Test_processFile(t *testing.T) {
	dir := t.TempDir()
	filePath := dir + "/test.log"
	data := `-{"msg": "test"}
+{"msg": "test2"}
-{"msg": "test3"}
`
	check(os.WriteFile(filePath, []byte(data), 0644))

	noop := func(_ []byte) error {
		return nil
	}
	check(processFile(filePath, noop))

	content, err := os.ReadFile(filePath)
	check(err)

	expected := `+{"msg": "test"}
+{"msg": "test2"}
+{"msg": "test3"}
`
	if string(content) != expected {
		t.Error("unexpected file content")
	}
}
