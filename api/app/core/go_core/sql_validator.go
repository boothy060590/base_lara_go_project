package go_core

import (
	"fmt"
	"regexp"
	"strings"
)

// SQLValidator validates SQL queries to prevent injection attacks
type SQLValidator struct {
	allowedKeywords   map[string]bool
	dangerousPatterns []*regexp.Regexp
	queryPattern      *regexp.Regexp
}

// NewSQLValidator creates a new SQL validator
func NewSQLValidator() *SQLValidator {
	// Compile dangerous patterns
	dangerousPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(union\s+select|select\s+union)`),
		regexp.MustCompile(`(?i)(drop\s+table|create\s+table|alter\s+table)`),
		regexp.MustCompile(`(?i)(delete\s+from|truncate\s+table)`),
		regexp.MustCompile(`(?i)(exec\s*\(|execute\s*\(|sp_\w+)`),
		regexp.MustCompile(`(?i)(waitfor\s+delay|sleep\s*\(|benchmark\s*\()`),
		regexp.MustCompile(`(?i)(information_schema|sys\.tables|sys\.columns)`),
		regexp.MustCompile(`(?i)(xp_cmdshell|sp_configure|sp_executesql)`),
		regexp.MustCompile(`(?i)(load_file|into\s+outfile|into\s+dumpfile)`),
		regexp.MustCompile(`(?i)(<script|javascript:|vbscript:)`),
		regexp.MustCompile(`(?i)(onload\s*=|onerror\s*=|onclick\s*=)`),
		regexp.MustCompile(`(?i)(--|/\*|\*/)`),                             // Comments
		regexp.MustCompile(`(?i)(0x[0-9a-f]+|0b[01]+|0o[0-7]+)`),           // Hex, binary, octal
		regexp.MustCompile(`(?i)(\\x[0-9a-f]+|\\u[0-9a-f]+|\\U[0-9a-f]+)`), // Unicode escapes
	}

	// Create regex pattern for valid query structure
	queryPattern := regexp.MustCompile(`(?i)^\s*(select|insert|update|delete)\s+.+$`)

	// Allowed SQL keywords for safe queries
	allowedKeywords := map[string]bool{
		"select": true, "from": true, "where": true, "and": true, "or": true,
		"insert": true, "update": true, "delete": true, "values": true, "set": true,
		"order": true, "group": true, "by": true, "having": true, "limit": true,
		"offset": true, "join": true, "inner": true, "left": true, "right": true,
		"outer": true, "on": true, "as": true, "in": true, "like": true,
		"between": true, "exists": true, "not": true, "null": true, "is": true,
		"distinct": true, "count": true, "sum": true, "avg": true, "max": true,
		"min": true, "case": true, "when": true, "then": true, "else": true,
		"end": true, "if": true, "coalesce": true, "nvl": true, "ifnull": true,
		"now": true, "current_timestamp": true, "current_date": true,
		"date": true, "time": true, "datetime": true, "timestamp": true,
		"year": true, "month": true, "day": true, "hour": true, "minute": true,
		"second": true, "concat": true, "substring": true, "length": true,
		"trim": true, "ltrim": true, "rtrim": true, "upper": true, "lower": true,
		"round": true, "floor": true, "ceil": true, "abs": true, "mod": true,
		"power": true, "sqrt": true, "exp": true, "log": true, "ln": true,
		"sin": true, "cos": true, "tan": true, "asin": true, "acos": true, "atan": true,
	}

	return &SQLValidator{
		allowedKeywords:   allowedKeywords,
		dangerousPatterns: dangerousPatterns,
		queryPattern:      queryPattern,
	}
}

// ValidateQuery validates a SQL query to prevent injection
func (sv *SQLValidator) ValidateQuery(query string) error {
	if query == "" {
		return fmt.Errorf("query cannot be empty")
	}

	// Check for dangerous patterns
	if sv.containsDangerousPatterns(query) {
		return fmt.Errorf("query contains dangerous patterns")
	}

	// Check query structure
	if !sv.isValidQueryStructure(query) {
		return fmt.Errorf("invalid query structure")
	}

	// Check for multiple statements
	if sv.containsMultipleStatements(query) {
		return fmt.Errorf("query contains multiple statements")
	}

	// Check for excessive length
	if len(query) > 10000 {
		return fmt.Errorf("query is too long")
	}

	// Check for balanced parentheses
	if !sv.hasBalancedParentheses(query) {
		return fmt.Errorf("query has unbalanced parentheses")
	}

	// Check for proper parameterization
	if sv.containsUnsafeStringConcatenation(query) {
		return fmt.Errorf("query contains unsafe string concatenation")
	}

	return nil
}

// containsDangerousPatterns checks for dangerous SQL patterns
func (sv *SQLValidator) containsDangerousPatterns(query string) bool {
	for _, pattern := range sv.dangerousPatterns {
		if pattern.MatchString(query) {
			return true
		}
	}
	return false
}

// isValidQueryStructure checks if query has valid structure
func (sv *SQLValidator) isValidQueryStructure(query string) bool {
	// Check if query starts with a valid SQL command
	if !sv.queryPattern.MatchString(query) {
		return false
	}

	// Check for balanced quotes
	if !sv.hasBalancedQuotes(query) {
		return false
	}

	return true
}

// containsMultipleStatements checks for multiple SQL statements
func (sv *SQLValidator) containsMultipleStatements(query string) bool {
	// Check for semicolons (statement separators)
	semicolonCount := strings.Count(query, ";")
	if semicolonCount > 1 {
		return true
	}

	// Check for multiple SELECT, INSERT, UPDATE, DELETE statements
	selectCount := len(regexp.MustCompile(`(?i)\bselect\b`).FindAllString(query, -1))
	insertCount := len(regexp.MustCompile(`(?i)\binsert\b`).FindAllString(query, -1))
	updateCount := len(regexp.MustCompile(`(?i)\bupdate\b`).FindAllString(query, -1))
	deleteCount := len(regexp.MustCompile(`(?i)\bdelete\b`).FindAllString(query, -1))

	totalStatements := selectCount + insertCount + updateCount + deleteCount
	return totalStatements > 1
}

// hasBalancedParentheses checks for balanced parentheses
func (sv *SQLValidator) hasBalancedParentheses(query string) bool {
	count := 0
	for _, char := range query {
		if char == '(' {
			count++
		} else if char == ')' {
			count--
			if count < 0 {
				return false
			}
		}
	}
	return count == 0
}

// hasBalancedQuotes checks for balanced quotes
func (sv *SQLValidator) hasBalancedQuotes(query string) bool {
	singleQuotes := strings.Count(query, "'")
	doubleQuotes := strings.Count(query, "\"")

	// Check for balanced single quotes
	if singleQuotes%2 != 0 {
		return false
	}

	// Check for balanced double quotes
	if doubleQuotes%2 != 0 {
		return false
	}

	return true
}

// containsUnsafeStringConcatenation checks for unsafe string concatenation
func (sv *SQLValidator) containsUnsafeStringConcatenation(query string) bool {
	// Check for string concatenation with variables
	unsafePatterns := []string{
		"'+", "+'", // String concatenation
		"\"+", "+\"", // String concatenation
		"||",      // Oracle concatenation
		"concat(", // Function concatenation
	}

	for _, pattern := range unsafePatterns {
		if strings.Contains(strings.ToLower(query), pattern) {
			return true
		}
	}

	return false
}

// ValidateParameterizedQuery validates a parameterized query
func (sv *SQLValidator) ValidateParameterizedQuery(query string, params []interface{}) error {
	// Validate the base query
	if err := sv.ValidateQuery(query); err != nil {
		return err
	}

	// Count placeholders in query
	placeholderCount := strings.Count(query, "?")

	// Check if parameter count matches placeholder count
	if placeholderCount != len(params) {
		return fmt.Errorf("parameter count (%d) does not match placeholder count (%d)", len(params), placeholderCount)
	}

	// Validate each parameter
	for i, param := range params {
		if err := sv.validateParameter(param); err != nil {
			return fmt.Errorf("parameter %d validation failed: %w", i+1, err)
		}
	}

	return nil
}

// validateParameter validates a single parameter
func (sv *SQLValidator) validateParameter(param interface{}) error {
	switch v := param.(type) {
	case string:
		// Check for dangerous patterns in string parameters
		if sv.containsDangerousPatterns(v) {
			return fmt.Errorf("string parameter contains dangerous patterns")
		}
		// Check for excessive length
		if len(v) > 1000 {
			return fmt.Errorf("string parameter is too long")
		}
	case []byte:
		// Check for dangerous patterns in byte parameters
		if sv.containsDangerousPatterns(string(v)) {
			return fmt.Errorf("byte parameter contains dangerous patterns")
		}
		// Check for excessive length
		if len(v) > 1000 {
			return fmt.Errorf("byte parameter is too long")
		}
	case nil:
		// NULL values are allowed
		return nil
	default:
		// Other types are generally safe
		return nil
	}

	return nil
}

// SanitizeQuery sanitizes a query for logging (removes sensitive data)
func (sv *SQLValidator) SanitizeQuery(query string) string {
	// Remove or mask sensitive patterns
	sanitized := query

	// Mask passwords
	passwordPattern := regexp.MustCompile(`(?i)(password\s*=\s*['"][^'"]*['"])`)
	sanitized = passwordPattern.ReplaceAllString(sanitized, "password=***")

	// Mask API keys
	apiKeyPattern := regexp.MustCompile(`(?i)(api_key\s*=\s*['"][^'"]*['"])`)
	sanitized = apiKeyPattern.ReplaceAllString(sanitized, "api_key=***")

	// Mask tokens
	tokenPattern := regexp.MustCompile(`(?i)(token\s*=\s*['"][^'"]*['"])`)
	sanitized = tokenPattern.ReplaceAllString(sanitized, "token=***")

	// Mask secrets
	secretPattern := regexp.MustCompile(`(?i)(secret\s*=\s*['"][^'"]*['"])`)
	sanitized = secretPattern.ReplaceAllString(sanitized, "secret=***")

	return sanitized
}

// IsAllowedKeyword checks if a keyword is allowed
func (sv *SQLValidator) IsAllowedKeyword(keyword string) bool {
	return sv.allowedKeywords[strings.ToLower(keyword)]
}

// GetAllowedKeywords returns all allowed keywords
func (sv *SQLValidator) GetAllowedKeywords() []string {
	keywords := make([]string, 0, len(sv.allowedKeywords))
	for keyword := range sv.allowedKeywords {
		keywords = append(keywords, keyword)
	}
	return keywords
}
