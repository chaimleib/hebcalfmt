package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
)

func usage(fs *pflag.FlagSet) string {
	return strings.Join(
		[]string{
			"usage:",
			fmt.Sprintf(
				"  %s [{ --config | -c } config.json ] template.tmpl [[ month [ day ]] year ]",
				ProgName,
			),
			fmt.Sprintf("  %s --list { cities | languages }", ProgName),
			fmt.Sprintf("  %s [ -h | --help | --version ]", ProgName),
			"",
			"OPTIONS:",
			fs.FlagUsages(),
		},
		"\n",
	)
}

func versionMessage() string {
	return fmt.Sprintf("%s version %s\n", ProgName, Version)
}
