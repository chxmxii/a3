// Package metautil provides safe accessors for the raw metadata maps
// (map[string]any, unmarshaled from Steampipe JSON) attached to resources.
package metautil

// GetString returns the first non-empty string value found under keys.
// Non-string values are ignored. Returns "" if m is nil or no key matches.
func GetString(m map[string]any, keys ...string) string {
	if m == nil {
		return ""
	}
	for _, key := range keys {
		if v, ok := m[key]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}
