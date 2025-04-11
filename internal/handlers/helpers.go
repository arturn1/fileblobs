package handlers

import "strings"

func filterByQuery(items []string, query string) []string {
	var filtered []string
	for _, item := range items {
		if strings.Contains(strings.ToLower(item), strings.ToLower(query)) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func baseName(path string) string {
	split := strings.Split(strings.TrimSuffix(path, "/"), "/")
	return split[len(split)-1]
}
