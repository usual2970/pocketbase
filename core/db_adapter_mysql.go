package core

import (
	"fmt"
	"strings"
)

// MySQLDBAdapter implements DBAdapter for MySQL
type MySQLDBAdapter struct {
	app App
}

func NewMySQLDBAdapter(app App) *MySQLDBAdapter { return &MySQLDBAdapter{app: app} }

func (a *MySQLDBAdapter) CreateTableQuery(table Table) (string, error) {
	tableName := table.Table
	columns := table.Columns
	indexes := table.Indexes
	comment := table.Comment

	var columnDefs []string
	for _, f := range columns {
		columnDefs = append(columnDefs, fmt.Sprintf("`%s` %s", f.GetName(), f.ColumnType(a.app)))
	}

	var indexDefs []string
	for _, idx := range indexes {
		uniqueKeyword := ""
		if idx.Unique {
			uniqueKeyword = "UNIQUE "
		}

		indexColumns := make([]string, len(idx.Columns))
		for i, colName := range idx.Columns {
			var isText bool
			for _, f := range columns {
				if f.GetName() == colName {
					colType := f.ColumnType(a.app)
					if strings.Contains(colType, "TEXT") || strings.Contains(colType, "VARCHAR") {
						indexColumns[i] = fmt.Sprintf("`%s`(191)", colName)
						isText = true
						break
					}
				}
			}
			if !isText {
				indexColumns[i] = fmt.Sprintf("`%s`", colName)
			}
		}

		def := fmt.Sprintf("%sKEY `%s` (%s)", uniqueKeyword, idx.Name, strings.Join(indexColumns, ", "))
		if idx.Comment != "" {
			def += fmt.Sprintf(" COMMENT '%s'", idx.Comment)
		}
		indexDefs = append(indexDefs, def)
	}

	allDefs := append(columnDefs, indexDefs...)

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (\n  %s\n)", tableName, strings.Join(allDefs, ",\n  "))
	if comment != "" {
		query += fmt.Sprintf(" COMMENT='%s'", comment)
	}
	return query, nil
}

func (a *MySQLDBAdapter) CreateIndexQuery(index Index) (string, error) {
	cols := make([]string, 0, len(index.Columns))
	for _, c := range index.Columns {
		cols = append(cols, fmt.Sprintf("`%s`(191)", c))
	}
	uniq := ""
	if index.Unique {
		uniq = "UNIQUE "
	}
	return fmt.Sprintf("CREATE %sINDEX `%s` ON `%s` (%s)", uniq, index.Name, index.TableName, strings.Join(cols, ", ")), nil
}

func (a *MySQLDBAdapter) DropIndexQuery(schemaName string, indexName string) (string, error) {
	return fmt.Sprintf("ALTER TABLE `%s` DROP INDEX `%s`", schemaName, indexName), nil
}

func (a *MySQLDBAdapter) FormatDateQuery(fieleName, format, as string) string {
	return fmt.Sprintf("DATE_FORMAT(`%s`, '%s') as `%s`", fieleName, format, as)
}
