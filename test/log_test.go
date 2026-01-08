package test_test

import (
	"log"
	"os"
	"testing"

	"github.com/chaimleib/hebcalfmt/test"
)

func TestLogger(t *testing.T) {
	cases := []struct {
		Name    string
		Actions func(t test.Test)
		Want    string
		Failed  bool
	}{
		{Name: "nothing", Actions: func(test.Test) {}},
		{
			Name:    "log hi",
			Actions: func(test.Test) { log.Println("hi") },
			Want:    "hi\n",
		},
		{
			Name:    "error hi",
			Actions: func(t test.Test) { log.Println("hi"); t.Fail() },
			Want:    "hi\n",
			Failed:  true,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			mockT := NewMockT(t)

			oldLogFlags := log.Default().Flags()
			oldLogPrefix := log.Default().Prefix()
			defer func() {
				log.SetFlags(oldLogFlags)
				log.SetPrefix(oldLogPrefix)
				log.SetOutput(os.Stderr)
			}()

			got := test.Logger(mockT)
			c.Actions(mockT)

			if c.Failed != mockT.Failed() {
				t.Errorf(
					"c.Failed was %v, but t.Failed() was %v",
					c.Failed,
					mockT.Failed(),
				)
			}
			if c.Want != got.String() {
				t.Errorf(
					"logs do not match - want:\n%s\ngot:\n%s",
					c.Want,
					got.String(),
				)
			}
		})
	}
}
