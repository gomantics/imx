package output

import "strings"

// isTimeField checks if a tag name is a time/date field
func isTimeField(name string) bool {
	name = strings.ToLower(name)
	return strings.Contains(name, "date") ||
		strings.Contains(name, "time") ||
		name == "datetime" ||
		name == "datetimeoriginal" ||
		name == "datetimedigitized" ||
		name == "modifydate" ||
		name == "createdate"
}
