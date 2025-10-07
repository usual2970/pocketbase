package core

import (
	"database/sql"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql" // register mysql driver
	_ "github.com/jackc/pgx/v5/stdlib" // register postgres driver for database/sql
	"github.com/pocketbase/dbx"
)

// BuildDBConnectFromEnv returns a DBConnectFunc that chooses the driver based on env vars.
// PB_DB_TYPE: sqlite | mysql | postgres
// PB_DB_DSN: full DSN for the selected driver (required for mysql/postgres)
func BuildDBConnectFromEnv() DBConnectFunc {
	dbType := os.Getenv("PB_DB_TYPE")
	dsn := os.Getenv("PB_DB_DSN")

	// Fallback to default sqlite behavior when unset or explicitly sqlite
	if dbType == "" || dbType == "sqlite" {
		return DefaultDBConnect
	}

	return func(_ string) (*dbx.DB, error) {
		driverName := dbType
		db, err := dbx.Open(driverName, dsn)
		if err != nil {
			return nil, err
		}

		applyPoolSettings(db.DB())
		if err := enforceMinimumVersion(driverName, db); err != nil {
			_ = db.Close()
			return nil, err
		}

		return db, nil
	}
}

func applyPoolSettings(sqldb *sql.DB) {
	maxOpen := 50
	maxIdle := 25
	lifetime := 30 * time.Minute

	if v := os.Getenv("PB_DB_MAX_OPEN"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxOpen = n
		}
	}
	if v := os.Getenv("PB_DB_MAX_IDLE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			maxIdle = n
		}
	}
	if v := os.Getenv("PB_DB_CONN_MAX_LIFETIME"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			lifetime = d
		}
	}

	sqldb.SetMaxOpenConns(maxOpen)
	sqldb.SetMaxIdleConns(maxIdle)
	sqldb.SetConnMaxLifetime(lifetime)
}

func enforceMinimumVersion(driverName string, db *dbx.DB) error {
	switch driverName {
	case "mysql":
		var version string
		if err := db.NewQuery("SELECT VERSION()").Row(&version); err != nil {
			return err
		}
		if !versionAtLeast(version, 8) {
			return ErrUnsupportedDBVersion("mysql", "8.0+", version)
		}
	case "postgres", "pgx":
		var version string
		if err := db.NewQuery("SHOW server_version").Row(&version); err != nil {
			return err
		}
		if !versionAtLeast(version, 13) {
			return ErrUnsupportedDBVersion("postgres", "13+", version)
		}
	}
	return nil
}

func versionAtLeast(version string, minMajor int) bool {
	v := strings.TrimSpace(version)
	if v == "" {
		return false
	}
	tok := v
	if i := strings.IndexAny(v, ". "); i >= 0 {
		tok = v[:i]
	}
	major, err := strconv.Atoi(strings.TrimLeft(tok, "vV"))
	if err != nil {
		return false
	}
	return major >= minMajor
}

type unsupportedVersionError struct {
	driver   string
	required string
	actual   string
}

func (e unsupportedVersionError) Error() string {
	return "database server version not supported for " + e.driver + ": requires " + e.required + ", got " + e.actual
}

func ErrUnsupportedDBVersion(driver, required, actual string) error {
	return unsupportedVersionError{driver: driver, required: required, actual: actual}
}
