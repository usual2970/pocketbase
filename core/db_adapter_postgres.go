package core

import (
	"fmt"
	"strings"
)

// PostgreSQLDBAdapter implements DBAdapter for PostgreSQL
type PostgreSQLDBAdapter struct{ app App }

func NewPostgreSQLDBAdapter(app App) *PostgreSQLDBAdapter { return &PostgreSQLDBAdapter{app: app} }

func (a *PostgreSQLDBAdapter) CreateTableQuery(table Table) (string, error) {
	tableName := table.Table
	columns := table.Columns
	indexes := table.Indexes
	comment := table.Comment

	var columnDefs []string
	for _, f := range columns {
		columnDefs = append(columnDefs, fmt.Sprintf("\"%s\" %s", f.GetName(), f.ColumnType(a.app)))
	}

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS \"%s\" (\n  %s\n)", tableName, strings.Join(columnDefs, ",\n  "))
	if comment != "" {
		query += fmt.Sprintf(";\nCOMMENT ON TABLE \"%s\" IS '%s'", tableName, comment)
	}

	for _, idx := range indexes {
		uniq := ""
		if idx.Unique {
			uniq = "UNIQUE "
		}
		indexQuery := fmt.Sprintf("CREATE %sINDEX IF NOT EXISTS \"%s\" ON \"%s\" (%s)", uniq, idx.Name, tableName, strings.Join(idx.Columns, ", "))
		query += ";\n" + indexQuery
		if idx.Comment != "" {
			query += fmt.Sprintf(";\nCOMMENT ON INDEX \"%s\" IS '%s'", idx.Name, idx.Comment)
		}
	}
	return query, nil
}

func (a *PostgreSQLDBAdapter) CreateIndexQuery(index Index) (string, error) {
	uniq := ""
	if index.Unique {
		uniq = "UNIQUE "
	}
	return fmt.Sprintf("CREATE %sINDEX IF NOT EXISTS \"%s\" ON \"%s\" (%s)", uniq, index.Name, index.TableName, strings.Join(index.Columns, ", ")), nil
}

func (a *PostgreSQLDBAdapter) DropIndexQuery(schemaName string, indexName string) (string, error) {
	return fmt.Sprintf("DROP INDEX IF EXISTS \"%s\"", indexName), nil
}

func (a *PostgreSQLDBAdapter) FormatDateQuery(fieleName, format, as string) string {
	return fmt.Sprintf("DATE_TRUNC('%s', \"%s\") as \"%s\"", format, fieleName, as)
}

func (a *PostgreSQLDBAdapter) DBOptimizeQuery() (string, error) {
	return "", fmt.Errorf("PostgreSQL does not require manual table optimization")
}
