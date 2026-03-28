package postgres

import "fmt"

// QualifiedTable returns a quoted schema.table reference for SQL (public uses unqualified table only).
func QualifiedTable(schema, table string) string {
	if schema != "" && schema != "public" {
		return fmt.Sprintf("%q.%q", schema, table)
	}
	return fmt.Sprintf("%q", table)
}
