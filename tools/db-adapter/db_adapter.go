package dbadapter

import "os"

type DBType string

const (
	DBTypeMySQL      DBType = "mysql"
	DBTypePostgreSQL DBType = "postgres"
	DBTypeSQLite     DBType = "sqlite"
)

// GetDriverNameFromDB extracts the driver name from env for now
func GetDriverNameFromDB() DBType {
	// Simplified: use env until richer detection is required
	dbType := os.Getenv("PB_DB_TYPE")
	switch dbType {
	case string(DBTypeMySQL):
		return DBTypeMySQL
	case string(DBTypePostgreSQL):
		return DBTypePostgreSQL
	case string(DBTypeSQLite):
		return DBTypeSQLite
	default:
		return DBTypeSQLite
	}
}

func IsMySQL() bool {
	return GetDriverNameFromDB() == DBTypeMySQL
}

func IsPostgreSQL() bool {
	return GetDriverNameFromDB() == DBTypePostgreSQL
}

func IsSQLite() bool {
	return GetDriverNameFromDB() == DBTypeSQLite
}
