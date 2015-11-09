package cache

import "sort"

// sort sorts a map[string]string and skips any empty values
func sortMap(m map[string]string, removeEmpty bool) (keys []string, values []string) {
	for k, v := range m {
		if removeEmpty && v == "" {
			continue
		}

		keys = append(keys, k)
	}

	sort.Strings(keys)
	for _, k := range keys {
		values = append(values, m[k])
	}

	return keys, values
}
