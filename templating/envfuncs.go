package templating

import "os"

var EnvFuncs = map[string]any{
	"getenv": os.Getenv,
}
