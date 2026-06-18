package tool

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

const maxRows = 200

type SQL struct {
	db *sql.DB
}

func NewSQL(db *sql.DB) *SQL {
	return &SQL{db: db}
}

func (t *SQL) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:        "run_sql",
		Description: "Виконує SELECT-запит до бази даних транзакцій і повертає рядки у форматі JSON.",
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"query": {
					Type:        genai.TypeString,
					Description: "Один SELECT-запит PostgreSQL. Без крапки з комою в кінці, без мутацій.",
				},
			},
			Required: []string{"query"},
		},
	}
}

func (t *SQL) Execute(ctx context.Context, args map[string]any) (map[string]any, error) {
	query, ok := args["query"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'query' argument")
	}
	if err := validateSelect(query); err != nil {
		return map[string]any{
			"error": fmt.Sprintf("invalid query: %s", err),
		}, nil
	}

	rows, err := t.db.QueryContext(ctx, query)
	if err != nil {
		return map[string]any{
			"error": fmt.Sprintf("query failed: %s", err),
		}, nil
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return map[string]any{
			"error": err.Error(),
		}, nil
	}

	var out []map[string]any
	for rows.Next() && len(out) < maxRows {
		values := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return map[string]any{
				"error": err.Error(),
			}, nil
		}
		row := make(map[string]any, len(cols))
		for i, c := range cols {
			row[c] = normalize(values[i])
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return map[string]any{
			"error": err.Error(),
		}, nil
	}

	payload, err := json.Marshal(out)
	if err != nil {
		return map[string]any{
			"error": err.Error(),
		}, nil
	}
	return map[string]any{
		"payload": string(payload),
	}, nil
}

func normalize(v any) any {
	if b, ok := v.([]byte); ok {
		return string(b)
	}
	return v
}

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
		return r < 'a' || r > 'z'
	}) {
		if field == word {
			return true
		}
	}
	return false
}
