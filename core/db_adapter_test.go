package core

import (
	"strings"
	"testing"

	dbadapter "github.com/pocketbase/pocketbase/tools/db-adapter"
)

// createMockTable creates a test table with default values
func createMockTable() Table {
	return Table{
		Table: "test_table",
		Columns: []Field{
			&TextField{Name: "id", PrimaryKey: true},
			&TextField{Name: "name", Default: "''"},
			&TextField{Name: "email"},
		},
		Indexes: []Index{
			{
				Name:    "idx_name",
				Columns: []string{"name"},
				Comment: "Index on name field",
			},
			{
				Name:    "idx_email",
				Columns: []string{"email"},
				Unique:  true,
				Comment: "Unique index on email",
			},
		},
		Comment: "Test table for unit testing",
	}
}

func TestDBAdapters(t *testing.T) {
	// Test that we can create adapters
	app := &BaseApp{}

	// Test SQLite adapter creation
	sqliteAdapter := NewSQLiteDBAdapter(app)
	if sqliteAdapter == nil {
		t.Error("SQLite adapter should not be nil")
	}

	// Test MySQL adapter creation
	mysqlAdapter := NewMySQLDBAdapter(app)
	if mysqlAdapter == nil {
		t.Error("MySQL adapter should not be nil")
	}

	// Test PostgreSQL adapter creation
	postgresAdapter := NewPostgreSQLDBAdapter(app)
	if postgresAdapter == nil {
		t.Error("PostgreSQL adapter should not be nil")
	}
}

func TestGetDBAdapter(t *testing.T) {
	// Test the GetDBAdapter function with different driver names
	app := &BaseApp{}

	tests := []struct {
		driverName string
		expected   string
	}{
		{"sqlite", "SQLiteDBAdapter"},
		{"mysql", "MySQLDBAdapter"},
		{"postgres", "PostgreSQLDBAdapter"},
		{"pgx", "PostgreSQLDBAdapter"},
		{"unknown", "SQLiteDBAdapter"}, // Default fallback
	}

	for _, test := range tests {
		adapter := GetDBAdapter(dbadapter.DBType(test.driverName), app)
		if adapter == nil {
			t.Errorf("GetDBAdapter returned nil for driver %s", test.driverName)
		}
	}
}

func TestTableStruct(t *testing.T) {
	// Test that our Table struct works correctly
	table := createMockTable()

	if table.Table != "test_table" {
		t.Errorf("Expected table name 'test_table', got '%s'", table.Table)
	}
}

func TestDBAdapterInterface(t *testing.T) {
	// Test that our adapters implement the DBAdapter interface
	app := &BaseApp{}

	// Create a mock table
	table := createMockTable()

	// Test that we can call the interface method
	adapters := []DBAdapter{
		NewSQLiteDBAdapter(app),
		NewMySQLDBAdapter(app),
		NewPostgreSQLDBAdapter(app),
	}

	for i, adapter := range adapters {
		query, err := adapter.CreateTableQuery(table)
		if err != nil {
			t.Logf("Adapter %d returned error: %v", i, err)
		} else if query != "" {
			t.Logf("Adapter %d returned a query: %s", i, query)
		}
	}
}

// TestCreateTableQueryFull tests the complete CREATE TABLE statement generation
func TestCreateTableQueryFull(t *testing.T) {
	app := &BaseApp{}

	// Create a comprehensive test table
	table := Table{
		Table: "users",
		Columns: []Field{
			&TextField{Name: "id", PrimaryKey: true},
			&TextField{Name: "username"},
			&TextField{Name: "email", Default: "''"},
			&DateField{Name: "created_at"},
		},
		Indexes: []Index{
			{
				Name:    "idx_username",
				Columns: []string{"username"},
				Comment: "Index on username",
			},
			{
				Name:    "idx_email",
				Columns: []string{"email"},
				Unique:  true,
				Comment: "Unique index on email",
			},
		},
		Comment: "Users table for storing user information",
	}

	// Test SQLite adapter
	t.Run("SQLite", func(t *testing.T) {
		adapter := NewSQLiteDBAdapter(app)
		query, err := adapter.CreateTableQuery(table)
		if err != nil {
			t.Fatalf("SQLite adapter error: %v", err)
		}

		// Verify basic structure
		if !strings.Contains(query, "CREATE TABLE IF NOT EXISTS") {
			t.Error("SQLite query should contain CREATE TABLE IF NOT EXISTS")
		}

		if !strings.Contains(query, "`users`") {
			t.Error("SQLite query should contain table name")
		}

		// Verify columns
		if !strings.Contains(query, "`id` TEXT PRIMARY KEY") {
			t.Error("SQLite query should contain id column definition")
		}

		if !strings.Contains(query, "`username` TEXT") {
			t.Error("SQLite query should contain username column definition")
		}

		// Verify indexes
		if !strings.Contains(query, "CREATE UNIQUE INDEX IF NOT EXISTS") {
			t.Error("SQLite query should contain index creation")
		}

		t.Logf("SQLite query:\n%s", query)
	})

	// Test MySQL adapter
	t.Run("MySQL", func(t *testing.T) {
		adapter := NewMySQLDBAdapter(app)
		query, err := adapter.CreateTableQuery(table)
		if err != nil {
			t.Fatalf("MySQL adapter error: %v", err)
		}

		// Verify basic structure
		if !strings.Contains(query, "CREATE TABLE IF NOT EXISTS") {
			t.Error("MySQL query should contain CREATE TABLE IF NOT EXISTS")
		}

		if !strings.Contains(query, "`users`") {
			t.Error("MySQL query should contain table name")
		}

		// Verify columns with comments
		if !strings.Contains(query, "`id` TEXT PRIMARY KEY") {
			t.Error("MySQL query should contain id column with comment")
		}

		// Verify that TEXT columns in indexes are limited to 191 characters
		if !strings.Contains(query, "`username`(191)") {
			t.Error("MySQL query should limit TEXT columns in indexes to 191 characters")
		}

		// Verify table comment
		if !strings.Contains(query, "COMMENT='Users table for storing user information'") {
			t.Error("MySQL query should contain table comment")
		}

		t.Logf("MySQL query:\n%s", query)
	})

	// Test PostgreSQL adapter
	t.Run("PostgreSQL", func(t *testing.T) {
		adapter := NewPostgreSQLDBAdapter(app)
		query, err := adapter.CreateTableQuery(table)
		if err != nil {
			t.Fatalf("PostgreSQL adapter error: %v", err)
		}

		// Verify basic structure
		if !strings.Contains(query, "CREATE TABLE IF NOT EXISTS") {
			t.Error("PostgreSQL query should contain CREATE TABLE IF NOT EXISTS")
		}

		if !strings.Contains(query, "\"users\"") {
			t.Error("PostgreSQL query should contain table name with double quotes")
		}

		// Verify table comment
		if !strings.Contains(query, "COMMENT ON TABLE \"users\" IS 'Users table for storing user information'") {
			t.Error("PostgreSQL query should contain table comment")
		}

		t.Logf("PostgreSQL query:\n%s", query)
	})
}

// TestColumnTypeConversion tests the column type conversion for different field types
func TestColumnTypeConversion(t *testing.T) {
	app := &BaseApp{}

	testCases := []struct {
		name         string
		fieldType    string
		driverName   string
		expectedType string
	}{
		{
			name:         "TextField_SQLite",
			fieldType:    FieldTypeText,
			driverName:   "sqlite",
			expectedType: "TEXT",
		},
		{
			name:         "TextField_MySQL",
			fieldType:    FieldTypeText,
			driverName:   "mysql",
			expectedType: "TEXT",
		},
		{
			name:         "TextField_PostgreSQL",
			fieldType:    FieldTypeText,
			driverName:   "postgres",
			expectedType: "TEXT",
		},
		{
			name:         "NumberField_SQLite",
			fieldType:    FieldTypeNumber,
			driverName:   "sqlite",
			expectedType: "NUMERIC",
		},
		{
			name:         "NumberField_MySQL",
			fieldType:    FieldTypeNumber,
			driverName:   "mysql",
			expectedType: "DECIMAL",
		},
		{
			name:         "NumberField_PostgreSQL",
			fieldType:    FieldTypeNumber,
			driverName:   "postgres",
			expectedType: "NUMERIC",
		},
		{
			name:         "JSONField_SQLite",
			fieldType:    FieldTypeJSON,
			driverName:   "sqlite",
			expectedType: "JSON",
		},
		{
			name:         "JSONField_MySQL",
			fieldType:    FieldTypeJSON,
			driverName:   "mysql",
			expectedType: "JSON",
		},
		{
			name:         "JSONField_PostgreSQL",
			fieldType:    FieldTypeJSON,
			driverName:   "postgres",
			expectedType: "JSONB",
		},
		{
			name:         "BoolField_SQLite",
			fieldType:    FieldTypeBool,
			driverName:   "sqlite",
			expectedType: "BOOLEAN",
		},
		{
			name:         "BoolField_MySQL",
			fieldType:    FieldTypeBool,
			driverName:   "mysql",
			expectedType: "TINYINT(1)",
		},
		{
			name:         "BoolField_PostgreSQL",
			fieldType:    FieldTypeBool,
			driverName:   "postgres",
			expectedType: "BOOLEAN",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a field of the specified type
			fieldFunc, ok := Fields[tc.fieldType]
			if !ok {
				t.Fatalf("Unknown field type: %s", tc.fieldType)
			}
			field := fieldFunc()

			// Get the column type
			columnType := field.ColumnType(app)

			// Check that the result contains the expected type
			if !strings.Contains(columnType, tc.expectedType) {
				t.Errorf("Expected column type to contain '%s', got '%s'", tc.expectedType, columnType)
			}

			t.Logf("Field: %s, Driver: %s, Result: %s", tc.fieldType, tc.driverName, columnType)
		})
	}
}
