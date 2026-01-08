package test_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/chaimleib/hebcalfmt/test"
)

var (
	ErrTestFailed = errors.New("test failed")
	ErrTestErr    = errors.New("test err")
)

// MockT is a mock for testing.T.
type MockT struct {
	buf      bytes.Buffer
	failed   bool
	cleanups []func()
}

var (
	_ test.Test = (*MockT)(nil)
	_ io.Closer = (*MockT)(nil)
)

func NewMockT(t *testing.T) *MockT {
	mt := new(MockT)
	t.Cleanup(func() { mt.Close() })
	return mt
}

func (mockT *MockT) Errorf(format string, args ...any) {
	mockT.failed = true
	fmt.Fprintf(&mockT.buf, format, args...)
	fmt.Fprintln(&mockT.buf)
}

func (mockT *MockT) Log(args ...any) {
	fmt.Fprintln(&mockT.buf, args...)
}
func (mockT *MockT) Helper()      {}
func (mockT *MockT) Fail()        { mockT.failed = true }
func (mockT *MockT) Failed() bool { return mockT.failed }

func (mockT *MockT) Cleanup(f func()) {
	mockT.cleanups = append(mockT.cleanups, f)
}

func (mockT *MockT) Close() error {
	for _, cleanup := range mockT.cleanups {
		cleanup()
	}
	return nil
}

func TestCheckErr(t *testing.T) {
	cases := []struct {
		Name      string
		GotInput  error
		WantInput string
		Want      string
		Failed    bool
	}{
		{Name: "ok nil", GotInput: nil, WantInput: "", Failed: false},
		{
			Name:      "unexpected nil",
			GotInput:  nil,
			WantInput: "never seen",
			Failed:    true,
			Want: `got nil err, want:
never seen
`,
		},
		{
			Name:      "unexpected err",
			GotInput:  ErrTestErr,
			WantInput: "",
			Failed:    true,
			Want: `want nil err, got:
test err
`,
		},
		{
			Name:      "wrong err",
			GotInput:  ErrTestErr,
			WantInput: "never seen",
			Failed:    true,
			Want: `got wrong err:
test err
want:
never seen
`,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			// DUT = Device Under Test, a term from electronics engineering
			mockT := NewMockT(t)
			test.CheckErr(mockT, c.GotInput, c.WantInput)
			if c.Failed != mockT.Failed() {
				t.Errorf(
					"c.Failed is %v, but dut.Failed() is %v",
					c.Failed,
					mockT.Failed(),
				)
			}
			if gotLogs := mockT.buf.String(); c.Want != gotLogs {
				t.Errorf("logs do not match - want:\n%s\ngot:\n%s", c.Want, gotLogs)
			}
		})
	}
}

// pointer returns a pointer to a copy of its input.
// Go 1.26 will allow new(expr), but not yet.
func pointer[T any](o T) *T { return &o }

func TestNilPtrThen(t *testing.T) {
	const fieldName = "Field"

	cases := []struct {
		Name      string
		GotInput  any
		WantInput any
		Want      string
		Failed    bool
	}{
		{
			Name:      "nils",
			GotInput:  new(any),
			WantInput: new(any),
			Failed:    false,
		},
		{
			Name:      "empty strings",
			GotInput:  pointer(""),
			WantInput: pointer(""),
			Failed:    false,
		},
		{
			Name:      "equal strings",
			GotInput:  pointer("hi"),
			WantInput: pointer("hi"),
			Failed:    false,
		},
		{
			Name:      "unequal strings",
			GotInput:  pointer("bye"),
			WantInput: pointer("hi"),
			Want: `Field's did not match - want:
"hi"
got:
"bye"
`,
			Failed: true,
		},
		{
			Name:      "nil vs. string",
			GotInput:  (*string)(nil),
			WantInput: pointer("hi"),
			Want: `Field's did not match - want pointer to:
"hi"
got:
(*string)(nil)
`,
			Failed: true,
		},
		{
			Name:      "string vs. nil",
			GotInput:  pointer("hi"),
			WantInput: (*string)(nil),
			Want: `Field's did not match - want:
(*string)(nil)
got pointer to:
"hi"
`,
			Failed: true,
		},
		{
			Name:      "equal ints",
			GotInput:  pointer(42),
			WantInput: pointer(42),
			Failed:    false,
		},
		{
			Name:      "unequal ints",
			GotInput:  pointer(-1),
			WantInput: pointer(42),
			Want: `Field's did not match - want:
42
got:
-1
`,
			Failed: true,
		},
		{
			Name:      "unequal types",
			GotInput:  pointer("hi"),
			WantInput: pointer(42),
			Want: `Field's types did not match - want:
*int
got:
*string
`,
			Failed: true,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			mockT := NewMockT(t)
			switch typedWant := c.WantInput.(type) {
			case *int:
				test.CheckNilPtrThen(mockT, test.CheckComparable, fieldName, typedWant, c.GotInput)

			case *string:
				test.CheckNilPtrThen(mockT, test.CheckComparable, fieldName, typedWant, c.GotInput)

			case *any:
				test.CheckNilPtrThen(mockT, test.CheckComparable, fieldName, typedWant, c.GotInput)

			default:
				t.Fatalf("unknown input type: %T", c.WantInput)
			}
			if c.Failed != mockT.Failed() {
				t.Errorf(
					"c.Failed is %v, but dut.Failed() is %v",
					c.Failed,
					mockT.Failed(),
				)
			}
			if gotLogs := mockT.buf.String(); c.Want != gotLogs {
				t.Errorf("logs do not match - want:\n%s\ngot:\n%s", c.Want, gotLogs)
			}
		})
	}
}

func TestCheckComparable(t *testing.T) {
	const fieldName = "Field"
	cases := []struct {
		Name      string
		GotInput  any
		WantInput any
		Want      string
		Failed    bool
	}{
		{
			Name:   "nils",
			Failed: false,
		},
		{
			Name:      "empty strings",
			GotInput:  "",
			WantInput: "",
			Failed:    false,
		},
		{
			Name:      "equal strings",
			GotInput:  "hi",
			WantInput: "hi",
			Failed:    false,
		},
		{
			Name:      "unequal strings",
			GotInput:  "bye",
			WantInput: "hi",
			Want: `Field's did not match - want:
"hi"
got:
"bye"
`,
			Failed: true,
		},
		{
			Name:      "equal ints",
			GotInput:  42,
			WantInput: 42,
			Failed:    false,
		},
		{
			Name:      "unequal ints",
			GotInput:  -1,
			WantInput: 42,
			Want: `Field's did not match - want:
42
got:
-1
`,
			Failed: true,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			mockT := NewMockT(t)
			switch typedWant := c.WantInput.(type) {
			case int:
				typedGot := c.GotInput.(int)
				test.CheckComparable(mockT, fieldName, typedWant, typedGot)

			case string:
				typedGot := c.GotInput.(string)
				test.CheckComparable(mockT, fieldName, typedWant, typedGot)

			default:
				test.CheckComparable(mockT, fieldName, c.WantInput, c.GotInput)
			}
			if c.Failed != mockT.Failed() {
				t.Errorf(
					"c.Failed is %v, but dut.Failed() is %v",
					c.Failed,
					mockT.Failed(),
				)
			}
			if gotLogs := mockT.buf.String(); c.Want != gotLogs {
				t.Errorf("logs do not match - want:\n%s\ngot:\n%s", c.Want, gotLogs)
			}
		})
	}
}
