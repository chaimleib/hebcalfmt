package cli_test

import (
	"fmt"
	"testing"

	"github.com/chaimleib/hebcalfmt/cli"
	"github.com/chaimleib/hebcalfmt/test"
)

func TestDefaultConfigPath(t *testing.T) {
	if cli.ProgName == "" {
		t.Error("cli.ProgName is unexpectedly empty")
	}

	cases := []struct {
		Name string
		Home string
		Want string
	}{
		{
			Name: "HOME is set",
			Home: "/home/user",
			Want: fmt.Sprintf("/home/user/.config/%s/config.json", cli.ProgName),
		},
		{
			Name: "HOME is empty",
			Home: "",
			Want: "",
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			t.Setenv("HOME", c.Home)
			got := cli.DefaultConfigPath()
			test.CheckString(t, "DefaultConfigPath", c.Want, got)
		})
	}
}
