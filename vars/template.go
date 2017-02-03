package vars

import (
	"bytes"
	"strings"
)

func ExpandAll(v Vars, templates []string) []string {
	results := make([]string, len(templates))
	for i, template := range templates {
		results[i] = Expand(v, template)
	}
	return results
}

func Expand(v Vars, current string) string {
	// XXX: hack

	final := new(bytes.Buffer)

	for lidx := strings.Index(current, "{{"); lidx >= 0; lidx = strings.Index(current, "{{") {

		final.WriteString(current[0:lidx])

		current = current[lidx+2:]

		if ridx := strings.Index(current, "}}"); ridx > 0 {
			key := current[0:ridx]
			if v.Has(key) {
				final.WriteString(v.Get(key))
				current = current[ridx+2:]
				continue
			}
		}

		final.WriteString("{{")
	}

	final.WriteString(current)

	return final.String()
}

func Extract(current string) []string {
	var varnames []string

	for lidx := strings.Index(current, "{{"); lidx >= 0; lidx = strings.Index(current, "{{") {
		current = current[lidx+2:]
		if ridx := strings.Index(current, "}}"); ridx > 0 {
			key := current[0:ridx]
			varnames = append(varnames, key)
			current = current[ridx+2:]
		}
	}

	return varnames
}
