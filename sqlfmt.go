// Package sqlfmt provides SQL formatting functionality using the sql-formatter JavaScript library.
// It supports multiple SQL dialects and offers multiple formatting options for consistent SQL code style.
package sqlfmt

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sync"

	"github.com/rosbit/go-quickjs"
)

//go:embed assets/sql-formatter.min.js
var jsCode []byte

// Common errors returned by the package.
var (
	ErrEmptySQL        = errors.New("empty SQL string")
	ErrSQLTooLarge     = errors.New("SQL string too large")
	ErrFormatterClosed = errors.New("formatter is closed")
)

// spaceBeforeParenRegex matches a space before ( that is not at the start of a line
// It's a hacky workaround for https://github.com/sql-formatter-org/sql-formatter/issues/444
// (?m) enables multiline mode
// ([^\n\s]) captures any character that's not a newline or space
// \s matches the space we want to remove
// \( matches the opening parenthesis
var spaceBeforeParenRegex = regexp.MustCompile(`(?m)([^\n\s])\s\(`)

// CaseOption defines the possible values for case-related formatting options.
type CaseOption string

const (
	// CaseOptionPreserve preserves the original case.
	CaseOptionPreserve CaseOption = "preserve"
	// CaseOptionUpper converts to uppercase.
	CaseOptionUpper CaseOption = "upper"
	// CaseOptionLower converts to lowercase.
	CaseOptionLower CaseOption = "lower"
)

// LogicalOperatorNewlineOption defines the possible values for logical operator newline placement.
type LogicalOperatorNewlineOption string

const (
	// LogicalOperatorNewlineBefore adds newline before the operator.
	LogicalOperatorNewlineBefore LogicalOperatorNewlineOption = "before"
	// LogicalOperatorNewlineAfter adds newline after the operator.
	LogicalOperatorNewlineAfter LogicalOperatorNewlineOption = "after"
)

// IndentStyleOption defines the possible values for indentation style.
type IndentStyleOption string

const (
	// IndentStyleStandard indents code by the amount specified by tabWidth option.
	IndentStyleStandard IndentStyleOption = "standard"
	// IndentStyleTabularLeft indents in tabular style with 10 spaces, aligning keywords to left.
	IndentStyleTabularLeft IndentStyleOption = "tabularLeft"
	// IndentStyleTabularRight indents in tabular style with 10 spaces, aligning keywords to right.
	IndentStyleTabularRight IndentStyleOption = "tabularRight"
)

// LanguageOption defines the possible SQL dialects.
type LanguageOption string

const (
	LanguageSQL           LanguageOption = "sql"
	LanguageBigQuery      LanguageOption = "bigquery"
	LanguageDB2           LanguageOption = "db2"
	LanguageDB2i          LanguageOption = "db2i"
	LanguageDuckDB        LanguageOption = "duckdb"
	LanguageHive          LanguageOption = "hive"
	LanguageMariaDB       LanguageOption = "mariadb"
	LanguageMySQL         LanguageOption = "mysql"
	LanguageTiDB          LanguageOption = "tidb"
	LanguageN1QL          LanguageOption = "n1ql"
	LanguagePLSQL         LanguageOption = "plsql"
	LanguagePostgreSQL    LanguageOption = "postgresql"
	LanguageRedshift      LanguageOption = "redshift"
	LanguageSingleStoreDB LanguageOption = "singlestoredb"
	LanguageSnowflake     LanguageOption = "snowflake"
	LanguageSpark         LanguageOption = "spark"
	LanguageSQLite        LanguageOption = "sqlite"
	LanguageTransactSQL   LanguageOption = "transactsql"
	LanguageTSQL          LanguageOption = "tsql"
	LanguageTrino         LanguageOption = "trino"
)

// FormatOptions configures how SQL queries should be formatted.
// It mirrors the options available in the sql-formatter JavaScript library.
// For detailed documentation, see: https://github.com/sql-formatter-org/sql-formatter/tree/master/docs
type FormatOptions struct {
	// Case of data types (e.g., INT, VARCHAR)
	DataTypeCase CaseOption `json:"dataTypeCase,omitempty"`
	// Whether to pack operators densely without spaces
	DenseOperators bool `json:"denseOperators,omitempty"`
	// Maximum length of parenthesized expressions
	ExpressionWidth int `json:"expressionWidth,omitempty"`
	// Case of function names (e.g., COUNT, SUM)
	FunctionCase CaseOption `json:"functionCase,omitempty"`
	// Case of identifiers (e.g., column names, table names)
	IdentifierCase CaseOption `json:"identifierCase,omitempty"`
	// Indentation style (deprecated in sql-examples, but still supported)
	IndentStyle IndentStyleOption `json:"indentStyle,omitempty"`
	// Case of reserved keywords (e.g., SELECT, FROM)
	KeywordCase CaseOption `json:"keywordCase,omitempty"`
	// SQL dialect to use (e.g., "mysql", "postgresql")
	Language LanguageOption `json:"language,omitempty"`
	// Number of empty lines between SQL statements
	LinesBetweenQueries int `json:"linesBetweenQueries,omitempty"`
	// Newline placement for logical operators (AND, OR, XOR)
	LogicalOperatorNewline LogicalOperatorNewlineOption `json:"logicalOperatorNewline,omitempty"`
	// Whether to place query separator (;) on a separate line
	NewlineBeforeSemicolon bool `json:"newlineBeforeSemicolon,omitempty"`
	// Number of spaces to be used for indentation (ignored if UseTabs is true)
	TabWidth int `json:"tabWidth,omitempty"`
	// Whether to use TAB characters for indentation instead of spaces
	UseTabs bool `json:"useTabs,omitempty"`
}

// DefaultFormatOptions provides a default configuration for SQL formatting.
var DefaultFormatOptions = FormatOptions{
	DataTypeCase:           CaseOptionUpper,
	DenseOperators:         false,
	ExpressionWidth:        80,
	FunctionCase:           CaseOptionUpper,
	IdentifierCase:         CaseOptionPreserve,
	IndentStyle:            IndentStyleStandard,
	KeywordCase:            CaseOptionUpper,
	Language:               LanguageSQL,
	LinesBetweenQueries:    2,
	LogicalOperatorNewline: LogicalOperatorNewlineAfter,
	NewlineBeforeSemicolon: true,
	TabWidth:               4,
	UseTabs:                false,
}

// Formatter provides SQL formatting functionality with a reusable JavaScript context.
type Formatter struct {
	ctx    *quickjs.JsContext
	mu     sync.Mutex
	closed bool
}

// NewFormatter creates a new SQL formatter instance.
// The returned Formatter must be closed when no longer needed to free resources.
func NewFormatter() (*Formatter, error) {
	ctx, err := quickjs.NewContext()
	if err != nil {
		return nil, fmt.Errorf("creating QuickJS context: %w", err)
	}

	f := &Formatter{ctx: ctx}

	if err := f.initialize(); err != nil {
		// QuickJS contexts don't have explicit close in this library
		return nil, err
	}

	return f, nil
}

// initialize sets up the JavaScript environment.
func (f *Formatter) initialize() error {
	// Evaluate the embedded JavaScript code
	_, err := f.ctx.Eval(string(jsCode), nil)
	if err != nil {
		return fmt.Errorf("evaluating sql-formatter.min.js: %w", err)
	}

	// Set up the formatting function
	setupCode := `
		function formatSql(sql, optionsJson) {
			const options = JSON.parse(optionsJson);
			return sqlFormatter.format(sql, options);
		}
	`
	_, err = f.ctx.Eval(setupCode, nil)
	if err != nil {
		return fmt.Errorf("setting up formatSql function: %w", err)
	}

	return nil
}

// Format formats a SQL query string according to the provided formatting options.
func (f *Formatter) Format(sql string, options FormatOptions) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return "", ErrFormatterClosed
	}

	// Marshal options to JSON
	optionsJSON, err := json.Marshal(options)
	if err != nil {
		return "", fmt.Errorf("marshaling options: %w", err)
	}

	// Call the JavaScript function
	res, err := f.ctx.CallFunc("formatSql", sql, string(optionsJSON))
	if err != nil {
		return "", fmt.Errorf("calling formatSql: %w", err)
	}

	// Convert result to string
	formatted, ok := res.(string)
	if !ok {
		return "", fmt.Errorf("unexpected result type: %T", res)
	}

	// Remove spaces before ( except at the start of lines
	formatted = spaceBeforeParenRegex.ReplaceAllString(formatted, "$1(")

	return formatted, nil
}

// Close releases the resources associated with the formatter.
// Note: The QuickJS library doesn't provide explicit close for contexts,
// but this method ensures the formatter is marked as closed.
func (f *Formatter) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return nil
	}

	f.closed = true
	return nil
}

// Format formats a SQL query string according to the provided formatting options.
// It creates a new formatter instance for each call, which may be inefficient for
// multiple formatting operations. Consider using NewFormatter for better performance.
//
// Parameters:
//   - sql: The SQL query string to format
//   - options: Formatting options to customize the output style
//
// Returns:
//   - The formatted SQL string
//   - An error if the formatting fails
//
// Example:
//
//	formatted, err := Format("SELECT * FROM users WHERE id = 1", DefaultFormatOptions)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(formatted)
func Format(sql string, options FormatOptions) (string, error) {
	f, err := NewFormatter()
	if err != nil {
		return "", err
	}
	defer func() {
		_ = f.Close() // Error is intentionally ignored as cleanup is best-effort
	}()

	return f.Format(sql, options)
}
