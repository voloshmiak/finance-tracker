package parser

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Each bank exports its statement in a distinct file format, so the file
// extension unambiguously selects the parser.
type Format string

const (
	FormatCSV  Format = "csv"  // monobank
	FormatXLSX Format = "xlsx" // credit agricole
	FormatPDF  Format = "pdf"  // abank
)

// Registry maps a statement format to the parser that handles it.
type Registry struct {
	parsers map[Format]BankParser
}

func NewRegistry() *Registry {
	return &Registry{
		parsers: map[Format]BankParser{
			FormatCSV:  NewMonobank(),
			FormatXLSX: NewCreditAgricole(),
			FormatPDF:  NewAbank(),
		},
	}
}

// For returns the parser for a format, or an error if the format is unsupported.
func (r *Registry) For(format Format) (BankParser, error) {
	p, ok := r.parsers[format]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnsupportedFormat, format)
	}
	return p, nil
}

// FormatFromFilename derives the statement format from a file name's extension.
func FormatFromFilename(name string) (Format, error) {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(name), "."))
	switch Format(ext) {
	case FormatCSV, FormatXLSX, FormatPDF:
		return Format(ext), nil
	default:
		return "", fmt.Errorf("%w: %q", ErrUnsupportedFormat, ext)
	}
}
