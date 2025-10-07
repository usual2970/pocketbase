package core

import (
	"fmt"
	"strings"
)

// SQLiteDBAdapter implements DBAdapter for SQLite
type SQLiteDBAdapter struct {
	app App
}

func NewSQLiteDBAdapter(app App) *SQLiteDBAdapter {
	return &SQLiteDBAdapter{app: app}
}

func (a *SQLiteDBAdapter) CreateTableQuery(table Table) (string, error) {
	tableName := table.Table
	columns := table.Columns
	indexes := table.Indexes
	comment := table.Comment

	var columnDefs []string
	for _, f := range columns {
		columnDefs = append(columnDefs, fmt.Sprintf("`%s` %s", f.GetName(), f.ColumnType(a.app)))
	}

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (\n  %s\n)",
		tableName,
		strings.Join(columnDefs, ",\n  "))

	if comment != "" {
		query += fmt.Sprintf(" -- %s", comment)
	}

	for _, idx := range indexes {
		uniqueKeyword := ""
		if idx.Unique {
			uniqueKeyword = "UNIQUE "
		}
		indexQuery := fmt.Sprintf("CREATE %sINDEX IF NOT EXISTS `%s` ON `%s` (%s)",
			uniqueKeyword, idx.Name, tableName, strings.Join(idx.Columns, ", "))
		if idx.Comment != "" {
			indexQuery += fmt.Sprintf(" -- %s", idx.Comment)
		}
		query += ";\n" + indexQuery
	}

	return query, nil
}

func (a *SQLiteDBAdapter) CreateIndexQuery(index Index) (string, error) {
	uniqueKeyword := ""
	if index.Unique {
		uniqueKeyword = "UNIQUE "
	}
	return fmt.Sprintf("CREATE %sINDEX IF NOT EXISTS `%s` ON `%s` (%s)",
		uniqueKeyword, index.Name, index.TableName, strings.Join(index.Columns, ", ")), nil
}

func (a *SQLiteDBAdapter) DropIndexQuery(schemaName string, indexName string) (string, error) {
	return fmt.Sprintf("DROP INDEX IF EXISTS `%s`", indexName), nil
}
	