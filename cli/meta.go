package cli

import (
	"fmt"
	"strings"
)

func usage(flagUsages string) string {
	return strings.Join(
		[]string{
			"usage:",
			fmt.Sprintf(
				"  %s [{ --config | -c } config.json ] template.tmpl [[ month [ day ]] year ]",
				ProgName,
			),
			fmt.Sprintf(
				"  %s { --info | -i }[=]{ %s }",
				ProgName,
				strings.Join(InfoKeys, " | "),
			),
			fmt.Sprintf("  %s [ -h | --help | --version ]", ProgName),
			"",
			"OPTIONS:",
			flagUsages,
		},
		"\n",
	)
}

func versionMessage() string {
	return fmt.Sprintf("%s %s", ProgName, Version)
}
