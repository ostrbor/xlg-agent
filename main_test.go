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

	t.Run("log last line", func(t *testing.T) {
		data := `+{"msg": "test"}
+{"msg": "test2"}
-{"msg": "test3"}
`
		check(os.WriteFile(filePath, []byte(data), 0644))
		noop := func(_ []byte) error {
			return nil
		}
		offset, err := process(filePath, 0, noop)
		check(err)
		if int(offset) != len(data) {
			t.Error("unexpected number of bytes processed ", offset)
		}

		content, err := os.ReadFile(filePath)
		check(err)

		expected := `+{"msg": "test"}
+{"msg": "test2"}
+{"msg": "test3"}
`
		if string(content) != expected {
			t.Error("unexpected file content")
		}
	})

	t.Run("partial last line", func(t *testing.T) {
		data := `+{"msg": "test"}
-{"msg": "test2"}
`
		partial := `{"msg": "te`
		content := data + partial

		check(os.WriteFile(filePath, []byte(content), 0644))
		noop := func(_ []byte) error {
			return nil
		}
		offset, err := process(filePath, 0, noop)
		check(err)
		if int(offset) != len(data) {
			t.Error("unexpected number of bytes processed ", offset)
		}

		c, err := os.ReadFile(filePath)
		check(err)

		expected := `+{"msg": "test"}
+{"msg": "test2"}
`
		expected += partial
		if string(c) != expected {
			t.Error("unexpected file content")
		}
	})

}
