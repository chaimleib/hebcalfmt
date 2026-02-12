package templating

import (
	"strings"

	"github.com/hebcal/hebcal-go/locales"
)

var StringFuncs = map[string]any{
	"contains":        strings.Contains,
	"containsAny":     strings.ContainsAny,
	"count":           strings.Count,
	"equalFold":       strings.EqualFold,
	"hasPrefix":       strings.HasPrefix,
	"hasSuffix":       strings.HasSuffix,
	"stringsIndex":    strings.Index,
	"stringsIndexAny": strings.IndexAny,
	"join":            strings.Join,
	"lastIndex":       strings.LastIndex,
	"lastIndexAny":    strings.LastIndexAny,
	"lines":           strings.Lines,
	"repeat":          strings.Repeat,
	"replace":         strings.Replace,
	"replaceAll":      strings.ReplaceAll,
	"split":           strings.Split,
	"splitAfter":      strings.SplitAfter,
	"splitAfterN":     strings.SplitAfterN,
	"splitAfterSeq":   strings.SplitAfterSeq,
	"splitN":          strings.SplitN,
	"splitSeq":        strings.SplitSeq,
	"toLower":         strings.ToLower,
	"toTitle":         strings.ToTitle,
	"toUpper":         strings.ToUpper,
	"toValidUTF8":     strings.ToValidUTF8,
	"trim":            strings.Trim,
	"trimLeft":        strings.TrimLeft,
	"tripPrefix":      strings.TrimPrefix,
	"trimRight":       strings.TrimRight,
	"trimSpace":       strings.TrimSpace,
	"trimSuffix":      strings.TrimSuffix,

	"translate": Translate,
}

// Translate attempts to translate s into lang
// using hebcal's translation dictionaries.
// If it fails, it returns s.
func Translate(lang, s string) string {
	if got, ok := locales.LookupTranslation(s, lang); ok {
		return got
	}
	return s
}

// Apply calls fn on each of the members of s and returns the results.
func Apply(s []string, fn func(string) string) []string {
	result := make([]string, 0, len(s))
	for _, item := range s {
		result = append(result, fn(item))
	}
	return result
}
