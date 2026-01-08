package test

import (
	"bytes"
	"fmt"
	"log"
	"os"
)

// Logger captures the output sent to the log package.
// If the test fails, the logs are printed.
// Otherwise, they are suppressed.
//
// It returns the buffer in case the logged output needs to be checked.
func Logger(t Test) fmt.Stringer {
	oldLogFlags := log.Default().Flags()

	var buf bytes.Buffer
	log.SetFlags(0) // suppress timestamps
	log.SetOutput(&buf)

	t.Cleanup(func() {
		log.SetFlags(oldLogFlags)
		log.SetOutput(os.Stderr)

		if t.Failed() && buf.Len() != 0 {
			t.Log("log output:")
			t.Log(buf.String())
		}
	})
	return &buf
}
