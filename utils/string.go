package utils

import (
	"fmt"
	"strings"
)

func SliceToSQLString(slice []string) string {
	// Wrap each element in single quotes
	quoted := make([]string, len(slice))
	for i, s := range slice {
		quoted[i] = fmt.Sprintf("'%s'", s)
	}
	// Join with commas and wrap in parentheses
	return fmt.Sprintf("(%s)", strings.Join(quoted, ", "))
}
