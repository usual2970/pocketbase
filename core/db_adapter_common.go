package core

import dbadapter "github.com/pocketbase/pocketbase/tools/db-adapter"

// Index represents a database index definition
type Index struct {
	Name      string
	TableName string
	Columns   []string
	Unique    bool
	Comment   string
}

// Table represents a database table definition
type Table struct {
	Table   string
	Columns []Field
	Indexes []Index
	Comment string
}

type DBAdapter interface {
	CreateTableQuery(table Table) (string, error)
	CreateIndexQuery(index Index) (string, error)
	DropIndexQuery(schemaName string, indexName string) (string, error)
	FormatDateQuery(fieleName, format, as string) string
	DBOptimizeQuery() (string, error)
}

// GetDBAdapter returns the appropriate database adapter based on the driver name
func GetDBAdapter(driverName dbadapter.DBType, app App) DBAdapter {
	switch driverName {
	case dbadapter.DBTypeMySQL:
		return NewMySQLDBAdapter(app)
	case dbadapter.DBTypePostgreSQL:
		return NewPostgreSQLDBAdapter(app)
	default: // SQLite
		return NewSQLiteDBAdapter(app)
	}
}
