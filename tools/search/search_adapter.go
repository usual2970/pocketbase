package search

import (
	"fmt"
	"strings"

	dbadapter "github.com/pocketbase/pocketbase/tools/db-adapter"
	"github.com/pocketbase/pocketbase/tools/inflector"
)

type SearchAdapter interface {
	BuildSortExpr(sortField SortField, fieldResolver FieldResolver) (string, error)
	PrefixTable(table string, expr string) string
}

func GetSearchAdapter() SearchAdapter {
	switch dbadapter.GetDriverNameFromDB() {
	case dbadapter.DBTypeMySQL:
		return &MySQLSearchAdapter{}
	default:
		return &SQLiteSearchAdapter{}
	}
}

type MySQLSearchAdapter struct {
}

func (a *MySQLSearchAdapter) BuildSortExpr(sortField SortField, fieldResolver FieldResolver) (string, error) {
	s := sortField
	if s.Name == randomSortKey {
		return "RANDOM()", nil
	}

	// special case for the builtin SQLite rowid column
	if s.Name == rowidSortKey {
		return fmt.Sprintf("created %s", s.Direction), nil
	}

	result, err := fieldResolver.Resolve(s.Name)

	// invalidate empty fields and non-column identifiers
	if err != nil || len(result.Params) > 0 || result.Identifier == "" || strings.ToLower(result.Identifier) == "null" {
		return "", fmt.Errorf("invalid sort field %q", s.Name)
	}

	return fmt.Sprintf("%s %s", result.Identifier, s.Direction), nil
}

func (a *MySQLSearchAdapter) PrefixTable(table string, expr string) string {
	return expr
}

type SQLiteSearchAdapter struct {
}

func (a *SQLiteSearchAdapter) BuildSortExpr(sortField SortField, fieldResolver FieldResolver) (string, error) {
	return sortField.BuildExpr(fieldResolver)
}

func (a *SQLiteSearchAdapter) PrefixTable(table string, expr string) string {
	return "[[" + inflector.Columnify(table) + "]]." + expr
}
