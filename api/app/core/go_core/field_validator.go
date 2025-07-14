package go_core

import (
	"fmt"
	"regexp"
	"strings"
)

// FieldValidator validates field names to prevent SQL injection
type FieldValidator struct {
	tableName     string
	allowedFields map[string]bool
	fieldPattern  *regexp.Regexp
}

// NewFieldValidator creates a new field validator
func NewFieldValidator(tableName string) *FieldValidator {
	// Create regex pattern for valid field names
	// Allow alphanumeric characters, underscores, and dots
	fieldPattern := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*(\.[a-zA-Z_][a-zA-Z0-9_]*)*$`)

	return &FieldValidator{
		tableName:     tableName,
		allowedFields: make(map[string]bool),
		fieldPattern:  fieldPattern,
	}
}

// ValidateField validates a field name to prevent SQL injection
func (fv *FieldValidator) ValidateField(field string) error {
	if field == "" {
		return fmt.Errorf("field name cannot be empty")
	}

	// Check for SQL injection patterns
	if fv.containsSQLInjection(field) {
		return fmt.Errorf("field name contains potential SQL injection: %s", field)
	}

	// Check field name pattern
	if !fv.fieldPattern.MatchString(field) {
		return fmt.Errorf("invalid field name format: %s", field)
	}

	// Check for reserved SQL keywords
	if fv.isReservedKeyword(field) {
		return fmt.Errorf("field name is a reserved SQL keyword: %s", field)
	}

	// Check for dangerous patterns
	if fv.containsDangerousPattern(field) {
		return fmt.Errorf("field name contains dangerous pattern: %s", field)
	}

	return nil
}

// containsSQLInjection checks for SQL injection patterns
func (fv *FieldValidator) containsSQLInjection(field string) bool {
	// Convert to lowercase for case-insensitive matching
	lowerField := strings.ToLower(field)

	// Check for SQL injection patterns
	dangerousPatterns := []string{
		"select", "insert", "update", "delete", "drop", "create", "alter",
		"union", "exec", "execute", "script", "javascript", "vbscript",
		"onload", "onerror", "onclick", "onmouseover",
		"--", "/*", "*/", "xp_", "sp_",
		"waitfor", "delay", "sleep",
		"information_schema", "sys.tables", "sys.columns",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerField, pattern) {
			return true
		}
	}

	// Check for comment patterns
	if strings.Contains(field, "--") || strings.Contains(field, "/*") || strings.Contains(field, "*/") {
		return true
	}

	// Check for semicolon (statement separator)
	if strings.Contains(field, ";") {
		return true
	}

	// Check for quotes
	if strings.Contains(field, "'") || strings.Contains(field, "\"") {
		return true
	}

	return false
}

// isReservedKeyword checks if field is a reserved SQL keyword
func (fv *FieldValidator) isReservedKeyword(field string) bool {
	// Convert to lowercase for case-insensitive matching
	lowerField := strings.ToLower(field)

	// Common SQL reserved keywords
	reservedKeywords := map[string]bool{
		"select": true, "from": true, "where": true, "and": true, "or": true,
		"insert": true, "update": true, "delete": true, "drop": true, "create": true,
		"alter": true, "table": true, "database": true, "index": true, "view": true,
		"procedure": true, "function": true, "trigger": true, "constraint": true,
		"primary": true, "foreign": true, "key": true, "unique": true, "check": true,
		"default": true, "null": true, "not": true, "in": true, "like": true,
		"between": true, "exists": true, "all": true, "any": true, "some": true,
		"union": true, "intersect": true, "except": true, "order": true, "group": true,
		"by": true, "having": true, "limit": true, "offset": true, "top": true,
		"distinct": true, "as": true, "join": true, "inner": true, "left": true,
		"right": true, "outer": true, "full": true, "cross": true, "natural": true,
		"on": true, "using": true, "case": true, "when": true, "then": true,
		"else": true, "end": true, "if": true, "elseif": true, "while": true,
		"for": true, "loop": true, "repeat": true, "until": true, "leave": true,
		"iterate": true, "return": true, "call": true, "declare": true, "set": true,
		"begin": true, "transaction": true, "commit": true, "rollback": true,
		"savepoint": true, "grant": true, "revoke": true, "deny": true, "use": true,
		"backup": true, "restore": true, "load": true, "dump": true, "lock": true,
		"unlock": true, "kill": true, "waitfor": true, "delay": true, "sleep": true,
		"user": true, "password": true, "admin": true, "root": true, "system": true,
		"information_schema": true, "sys": true, "mysql": true, "performance_schema": true,
	}

	return reservedKeywords[lowerField]
}

// containsDangerousPattern checks for dangerous patterns in field names
func (fv *FieldValidator) containsDangerousPattern(field string) bool {
	// Check for patterns that could be used for injection
	dangerousPatterns := []string{
		"1=1", "1=0", "true", "false",
		"0x", "0b", "0o", // Hex, binary, octal prefixes
		"\\x", "\\u", "\\U", // Unicode escapes
		"\\n", "\\r", "\\t", // Control characters
		"<script", "</script", // Script tags
		"javascript:", "vbscript:", // Script protocols
		"data:", "file:", // Data protocols
		"onload", "onerror", "onclick", // Event handlers
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(strings.ToLower(field), pattern) {
			return true
		}
	}

	// Check for excessive length
	if len(field) > 255 {
		return true
	}

	// Check for excessive dots (could indicate path traversal)
	if strings.Count(field, ".") > 5 {
		return true
	}

	return false
}

// AddAllowedField adds a field to the allowed fields list
func (fv *FieldValidator) AddAllowedField(field string) {
	fv.allowedFields[strings.ToLower(field)] = true
}

// RemoveAllowedField removes a field from the allowed fields list
func (fv *FieldValidator) RemoveAllowedField(field string) {
	delete(fv.allowedFields, strings.ToLower(field))
}

// IsAllowedField checks if a field is in the allowed fields list
func (fv *FieldValidator) IsAllowedField(field string) bool {
	return fv.allowedFields[strings.ToLower(field)]
}

// GetAllowedFields returns all allowed fields
func (fv *FieldValidator) GetAllowedFields() []string {
	fields := make([]string, 0, len(fv.allowedFields))
	for field := range fv.allowedFields {
		fields = append(fields, field)
	}
	return fields
}

// ValidateFields validates multiple field names
func (fv *FieldValidator) ValidateFields(fields []string) error {
	for _, field := range fields {
		if err := fv.ValidateField(field); err != nil {
			return fmt.Errorf("field validation failed for '%s': %w", field, err)
		}
	}
	return nil
}

// GetTableName returns the table name
func (fv *FieldValidator) GetTableName() string {
	return fv.tableName
}
