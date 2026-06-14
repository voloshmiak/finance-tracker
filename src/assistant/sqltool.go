package assistant

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

// maxRows caps how many rows a single tool call may return, to keep the model's
// context bounded.
const maxRows = 200

// SQLTool executes read-only SELECT queries against the transactions database.
// It is deliberately backed by a read-only connection AND validates the query,
// so a hallucinated or injected mutation is rejected at two layers.
type SQLTool struct {
	db *sql.DB
}

func NewSQLTool(db *sql.DB) *SQLTool {
	return &SQLTool{db: db}
}

// Run validates and executes a SELECT, returning the rows as JSON text suitable
// for feeding back to the model.
func (t *SQLTool) Run(ctx context.Context, query string) (string, error) {
	if err := validateSelect(query); err != nil {
		return "", err
	}

	rows, err := t.db.QueryContext(ctx, query)
	if err != nil {
		return "", fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return "", err
	}

	var out []map[string]any
	for rows.Next() && len(out) < maxRows {
		values := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return "", err
		}
		row := make(map[string]any, len(cols))
		for i, c := range cols {
			row[c] = normalize(values[i])
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return "", err
	}

	payload, err := json.Marshal(out)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

// normalize converts driver byte slices to strings so JSON output is readable.
func normalize(v any) any {
	if b, ok := v.([]byte); ok {
		return string(b)
	}
	return v
}

// validateSelect is the in-code guard layer: only a single SELECT/WITH query is
// allowed, no statement chaining, no mutating keywords.
func validateSelect(query string) error {
	q := strings.TrimSpace(query)
	q = strings.TrimSuffix(q, ";")
	if q == "" {
		return fmt.Errorf("empty query")
	}
	if strings.Contains(q, ";") {
		return fmt.Errorf("multiple statements are not allowed")
	}

	lower := strings.ToLower(q)
	if !strings.HasPrefix(lower, "select") && !strings.HasPrefix(lower, "with") {
		return fmt.Errorf("only SELECT queries are allowed")
	}

	forbidden := []string{
		"insert", "update", "delete", "drop", "alter", "create",
		"truncate", "grant", "revoke", "copy", "merge",
	}
	for _, kw := range forbidden {
		// match as a whole word to avoid false positives (e.g. "created_at")
		if containsWord(lower, kw) {
			return fmt.Errorf("forbidden keyword %q in query", kw)
		}
	}
	return nil
}

func containsWord(s, word string) bool {
	for _, field := range strings.FieldsFunc(s, func(r rune) bool {
		return !(r >= 'a' && r <= 'z')
	}) {
		if field == word {
			return true
		}
	}
	return false
}
