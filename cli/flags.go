package cli

import (
	"fmt"
	"reflect"
	"strings"
)

type Flags struct {
	Help    bool   `alt:"h" desc:"print this help text"`
	Version bool   `alt:"v" desc:"show version number"`
	Config  string `alt:"c"                             descfunc:"HelpConfig"`
	Info    string `alt:"i"                             descfunc:"HelpInfo"`
}

func (*Flags) HelpConfig() string {
	return "select a JSON config file (default $HOME/.config/hebcalfmt/config.json)"
}

func (*Flags) HelpInfo() string {
	return fmt.Sprintf(
		"show data from the internal databases or compiled values. Available options: %q",
		InfoKeys,
	)
}

type FlagSet struct {
	help    []string
	short   map[rune]int
	long    map[string]int
	factory reflect.Type
}

func ParseFlags(args []string) (*Flags, []string, error) {
	var flags Flags
	var rest []string

	flagsVal := reflect.ValueOf(flags)
	typ := flagsVal.Type()

	short := map[string]int{}
	long := map[string]int{}
	help := make([]string, typ.NumField())
	for i := range typ.NumField() {
		field := typ.Field(i)
		long[ToKebab(field.Name)] = i

		tag := field.Tag
		if alt := tag.Get("alt"); alt != "" {
			if len(alt) == 1 {
				short[alt] = i
			} else {
				long[alt] = i
			}
		}

		desc, err := FieldDesc(flagsVal, tag)
		if err != nil {
			return nil, nil, fmt.Errorf("ParseFlags failed: %w", err)
		}
		help[i] = desc
	}

	return &flags, rest, nil
}

func FieldDesc(structVal reflect.Value, tag reflect.StructTag) (string, error) {
	if desc := tag.Get("desc"); desc != "" {
		return desc, nil
	}

	descFuncName := tag.Get("descfunc")
	if descFuncName == "" {
		return "", nil
	}

	structType := structVal.Type()
	typeName := structType.Name()
	if pkgPath := structType.PkgPath(); pkgPath != "" {
		typeName = fmt.Sprintf("%s.%s", pkgPath, typeName)
	}

	methodVal := structVal.MethodByName(descFuncName)
	if methodVal.IsZero() {
		return "", fmt.Errorf("no such method: (%s).%s", typeName, descFuncName)
	}
	methodType := methodVal.Type()
	if methodType.NumIn() != 0 {
		return "", fmt.Errorf(
			"descfunc must not accept any arguments, got %s is %T",
			descFuncName,
			methodVal.Interface(),
		)
	}
	switch descfunc := methodVal.Interface().(type) {
	case func() string:
		return descfunc(), nil

	case func() (string, error):
		return descfunc()

	default:
		return "", fmt.Errorf(
			"descfunc must return a string or (string, error), got %s is %T",
			descFuncName,
			methodVal.Interface(),
		)
	}
}

func ToKebab(name string) string {
	var b strings.Builder
	b.Grow(len(name))
	const lowerDelta = 'a' - 'A'
	for i, r := range name {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteRune('-')
			}
			b.WriteRune(r - lowerDelta)
			continue
		}
		if r == '_' || r == ' ' {
			b.WriteRune('-')
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
