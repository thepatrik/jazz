package jazz

import (
	"fmt"
	"sort"
	"strings"
)

// JazzDict is a first-class dictionary/map value. It is backed by a native Go
// map keyed on string, so only string keys are supported (mirroring cjazz,
// whose backing hash table is keyed on C strings).
type JazzDict struct {
	Entries map[string]interface{}
}

func NewJazzDict(entries map[string]interface{}) *JazzDict {
	return &JazzDict{Entries: entries}
}

func (d *JazzDict) String() string {
	// Go map iteration order is randomised; sort keys for stable output.
	keys := make([]string, 0, len(d.Entries))
	for k := range d.Entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = fmt.Sprintf("%q: %v", k, d.Entries[k])
	}
	return "{" + strings.Join(parts, ", ") + "}"
}
