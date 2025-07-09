package go_core

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// ValidationRule represents a single validation rule
type ValidationRule interface {
	Validate(value any, data map[string]any) error
	GetMessage() string
}

// RuleFunc is a function-based validation rule (like Laravel closures)
type RuleFunc func(value any, data map[string]any) error

// Rule represents a validation rule with optional parameters
type Rule struct {
	Name       string
	Parameters []string
	Func       RuleFunc
	Message    string
}

// Validate implements ValidationRule interface
func (r *Rule) Validate(value any, data map[string]any) error {
	if r.Func != nil {
		return r.Func(value, data)
	}
	return fmt.Errorf("rule %s not implemented", r.Name)
}

// GetMessage returns the custom message for this rule
func (r *Rule) GetMessage() string {
	return r.Message
}

// Validator provides Laravel-style validation with performance optimizations
type Validator[T any] struct {
	rules    map[string][]ValidationRule
	messages map[string]string
	data     T
	dataMap  map[string]any
	// Performance optimizations (safe for validation operations)
	atomicCounter     *AtomicCounter
	rulePool          *ObjectPool[Rule]
	performanceFacade *PerformanceFacade
}

// NewValidator creates a new validator for type T with performance optimizations
func NewValidator[T any](data T) *Validator[T] {
	// Create performance optimizations
	atomicCounter := NewAtomicCounter()
	performanceFacade := NewPerformanceFacade()

	// Create object pool for rule objects (safe - no database state)
	rulePool := NewObjectPool[Rule](100,
		func() Rule { return Rule{} },
		func(rule Rule) Rule { return Rule{} },
	)

	validator := &Validator[T]{
		rules:             make(map[string][]ValidationRule),
		messages:          make(map[string]string),
		data:              data,
		atomicCounter:     atomicCounter,
		rulePool:          rulePool,
		performanceFacade: performanceFacade,
	}
	validator.dataMap = validator.structToMap(data)
	return validator
}

// Rules sets the validation rules with performance tracking and atomic counter
func (v *Validator[T]) Rules(rules map[string]any) *Validator[T] {
	// Track operation count atomically
	v.atomicCounter.Increment()

	v.performanceFacade.Track("validation.rules", func() error {
		for field, rule := range rules {
			v.rules[field] = v.parseRules(rule)
		}
		return nil
	})

	return v
}

// Messages sets custom validation messages
func (v *Validator[T]) Messages(messages map[string]string) *Validator[T] {
	v.messages = messages
	return v
}

// Validate validates the data against the rules with performance tracking and atomic counter
func (v *Validator[T]) Validate() (bool, map[string][]string) {
	// Track operation count atomically
	v.atomicCounter.Increment()

	var errors map[string][]string

	v.performanceFacade.Track("validation.validate", func() error {
		errors = make(map[string][]string)

		for field, rules := range v.rules {
			value := v.dataMap[field]

			for _, rule := range rules {
				if err := rule.Validate(value, v.dataMap); err != nil {
					message := rule.GetMessage()
					if message == "" {
						message = err.Error()
					}

					if errors[field] == nil {
						errors[field] = []string{}
					}
					errors[field] = append(errors[field], message)
				}
			}
		}

		return nil
	})

	return len(errors) == 0, errors
}

// Request Input Access Methods (Laravel-style)

// Get retrieves a value from the request data
func (v *Validator[T]) Get(key string) any {
	return v.getNestedValue(key)
}

// GetString retrieves a string value from the request data
func (v *Validator[T]) GetString(key string) string {
	value := v.Get(key)
	if str, ok := value.(string); ok {
		return str
	}
	return ""
}

// GetInt retrieves an integer value from the request data
func (v *Validator[T]) GetInt(key string) int {
	value := v.Get(key)
	switch v := value.(type) {
	case int:
		return v
	case int8:
		return int(v)
	case int16:
		return int(v)
	case int32:
		return int(v)
	case int64:
		return int(v)
	case uint:
		return int(v)
	case uint8:
		return int(v)
	case uint16:
		return int(v)
	case uint32:
		return int(v)
	case uint64:
		return int(v)
	case float32:
		return int(v)
	case float64:
		return int(v)
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return 0
}

// GetFloat retrieves a float value from the request data
func (v *Validator[T]) GetFloat(key string) float64 {
	value := v.Get(key)
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int8:
		return float64(v)
	case int16:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case uint:
		return float64(v)
	case uint8:
		return float64(v)
	case uint16:
		return float64(v)
	case uint32:
		return float64(v)
	case uint64:
		return float64(v)
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return 0.0
}

// GetBool retrieves a boolean value from the request data
func (v *Validator[T]) GetBool(key string) bool {
	value := v.Get(key)
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return v == "true" || v == "1" || v == "yes" || v == "on"
	case int:
		return v != 0
	case float64:
		return v != 0
	}
	return false
}

// Has checks if a key exists in the request data
func (v *Validator[T]) Has(key string) bool {
	value := v.Get(key)
	return value != nil
}

// Input retrieves a value from the request data (alias for Get)
func (v *Validator[T]) Input(key string) any {
	return v.Get(key)
}

// All returns all request data
func (v *Validator[T]) All() map[string]any {
	return v.dataMap
}

// Only returns only the specified keys
func (v *Validator[T]) Only(keys []string) map[string]any {
	result := make(map[string]any)
	for _, key := range keys {
		if value := v.Get(key); value != nil {
			result[key] = value
		}
	}
	return result
}

// Except returns all except the specified keys
func (v *Validator[T]) Except(keys []string) map[string]any {
	result := make(map[string]any)
	excluded := make(map[string]bool)
	for _, key := range keys {
		excluded[key] = true
	}

	for key, value := range v.dataMap {
		if !excluded[key] {
			result[key] = value
		}
	}
	return result
}

// Validated returns only the validated data (keys that have validation rules)
func (v *Validator[T]) Validated() map[string]any {
	result := make(map[string]any)
	for key := range v.rules {
		if value := v.Get(key); value != nil {
			result[key] = value
		}
	}
	return result
}

// GetPerformanceStats returns validation performance statistics
func (v *Validator[T]) GetPerformanceStats() map[string]interface{} {
	stats := v.performanceFacade.GetStats()

	// Add validation-specific stats
	stats["validation"] = map[string]interface{}{
		"operations_count": v.atomicCounter.Get(),
		"rule_pool_size":   len(v.rulePool.pool),
		"rules_count":      len(v.rules),
	}

	return stats
}

// GetOptimizationStats returns validation optimization statistics
func (v *Validator[T]) GetOptimizationStats() map[string]interface{} {
	return map[string]interface{}{
		"atomic_operations": v.atomicCounter.Get(),
		"rule_pool_usage":   len(v.rulePool.pool),
		"rules_count":       len(v.rules),
	}
}

// getNestedValue retrieves a nested value using dot notation (e.g., "data.jobs.0")
func (v *Validator[T]) getNestedValue(key string) any {
	parts := strings.Split(key, ".")
	var current any = v.dataMap

	for i, part := range parts {
		if current == nil {
			return nil
		}

		// Handle array access (e.g., "jobs.0")
		if strings.HasPrefix(part, "[") && strings.HasSuffix(part, "]") {
			// Array access like "jobs[0]"
			indexStr := part[1 : len(part)-1]
			if index, err := strconv.Atoi(indexStr); err == nil {
				if arr, ok := current.([]any); ok && index >= 0 && index < len(arr) {
					current = arr[index]
				} else {
					return nil
				}
			} else {
				return nil
			}
		} else if strings.Contains(part, "[") {
			// Mixed access like "jobs[0].name"
			bracketIndex := strings.Index(part, "[")
			fieldName := part[:bracketIndex]
			indexStr := part[bracketIndex+1 : strings.Index(part, "]")]

			if index, err := strconv.Atoi(indexStr); err == nil {
				if mapData, ok := current.(map[string]any); ok {
					if arr, ok := mapData[fieldName].([]any); ok && index >= 0 && index < len(arr) {
						current = arr[index]
					} else {
						return nil
					}
				} else {
					return nil
				}
			} else {
				return nil
			}
		} else {
			// Regular field access
			if mapData, ok := current.(map[string]any); ok {
				if value, exists := mapData[part]; exists {
					current = value
				} else {
					return nil
				}
			} else {
				return nil
			}
		}

		// If this is the last part, return the value
		if i == len(parts)-1 {
			return current
		}
	}

	return current
}

// parseRules parses Laravel-style rule definitions
func (v *Validator[T]) parseRules(rule any) []ValidationRule {
	var rules []ValidationRule

	switch r := rule.(type) {
	case string:
		// Parse pipe-separated rules: "required|string|max:100"
		ruleStrings := strings.Split(r, "|")
		for _, ruleStr := range ruleStrings {
			if rule := v.parseRuleString(ruleStr); rule != nil {
				rules = append(rules, rule)
			}
		}
	case []any:
		// Parse array of rules: ["required", "string", Rule::exists(...)]
		for _, item := range r {
			switch item := item.(type) {
			case string:
				if rule := v.parseRuleString(item); rule != nil {
					rules = append(rules, rule)
				}
			case ValidationRule:
				rules = append(rules, item)
			case RuleFunc:
				rules = append(rules, &Rule{Func: item})
			}
		}
	case ValidationRule:
		rules = append(rules, r)
	case RuleFunc:
		rules = append(rules, &Rule{Func: r})
	}

	return rules
}

// parseRuleString parses a single rule string like "max:100" or "unique:users,email"
func (v *Validator[T]) parseRuleString(ruleStr string) ValidationRule {
	parts := strings.SplitN(ruleStr, ":", 2)
	ruleName := parts[0]

	var parameters []string
	if len(parts) > 1 {
		parameters = strings.Split(parts[1], ",")
	}

	// Return built-in rule or custom rule
	return v.getBuiltInRule(ruleName, parameters)
}

// getBuiltInRule returns a built-in validation rule
func (v *Validator[T]) getBuiltInRule(name string, parameters []string) ValidationRule {
	switch name {
	case "required":
		return &Rule{
			Name: name,
			Func: func(value any, data map[string]any) error {
				if value == nil || value == "" {
					return fmt.Errorf("the field is required")
				}
				return nil
			},
		}
	case "string":
		return &Rule{
			Name: name,
			Func: func(value any, data map[string]any) error {
				if value == nil {
					return nil // Skip if not required
				}
				if _, ok := value.(string); !ok {
					return fmt.Errorf("the field must be a string")
				}
				return nil
			},
		}
	case "int":
		return &Rule{
			Name: name,
			Func: func(value any, data map[string]any) error {
				if value == nil {
					return nil // Skip if not required
				}
				switch v := value.(type) {
				case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
					return nil
				case string:
					if _, err := strconv.Atoi(v); err != nil {
						return fmt.Errorf("the field must be an integer")
					}
					return nil
				default:
					return fmt.Errorf("the field must be an integer")
				}
			},
		}
	case "email":
		return &Rule{
			Name: name,
			Func: func(value any, data map[string]any) error {
				if value == nil {
					return nil // Skip if not required
				}
				email, ok := value.(string)
				if !ok {
					return fmt.Errorf("the field must be a string")
				}

				emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
				if !emailRegex.MatchString(email) {
					return fmt.Errorf("the field must be a valid email address")
				}
				return nil
			},
		}
	case "max":
		if len(parameters) == 0 {
			return nil
		}
		max, err := strconv.Atoi(parameters[0])
		if err != nil {
			return nil
		}
		return &Rule{
			Name:       name,
			Parameters: parameters,
			Func: func(value any, data map[string]any) error {
				if value == nil {
					return nil // Skip if not required
				}

				switch v := value.(type) {
				case string:
					if len(v) > max {
						return fmt.Errorf("the field may not be greater than %d characters", max)
					}
				case []any:
					if len(v) > max {
						return fmt.Errorf("the field may not have more than %d items", max)
					}
				}
				return nil
			},
		}
	case "min":
		if len(parameters) == 0 {
			return nil
		}
		min, err := strconv.Atoi(parameters[0])
		if err != nil {
			return nil
		}
		return &Rule{
			Name:       name,
			Parameters: parameters,
			Func: func(value any, data map[string]any) error {
				if value == nil {
					return nil // Skip if not required
				}

				switch v := value.(type) {
				case string:
					if len(v) < min {
						return fmt.Errorf("the field must be at least %d characters", min)
					}
				case []any:
					if len(v) < min {
						return fmt.Errorf("the field must have at least %d items", min)
					}
				}
				return nil
			},
		}
	case "unique":
		if len(parameters) < 2 {
			return nil
		}
		_ = parameters[0] // table
		_ = parameters[1] // column
		return &Rule{
			Name:       name,
			Parameters: parameters,
			Func: func(value any, data map[string]any) error {
				// TODO: Implement database uniqueness check
				// This would need access to the database connection
				return nil
			},
		}
	case "exists":
		if len(parameters) < 2 {
			return nil
		}
		_ = parameters[0] // table
		_ = parameters[1] // column
		return &Rule{
			Name:       name,
			Parameters: parameters,
			Func: func(value any, data map[string]any) error {
				// TODO: Implement database existence check
				// This would need access to the database connection
				return nil
			},
		}
	}

	return nil
}

// structToMap converts a struct to a map for validation
func (v *Validator[T]) structToMap(data T) map[string]any {
	result := make(map[string]any)

	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return result
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Get field name (support json tags)
		fieldName := fieldType.Name
		if jsonTag := fieldType.Tag.Get("json"); jsonTag != "" {
			if parts := strings.Split(jsonTag, ","); len(parts) > 0 {
				fieldName = parts[0]
			}
		}

		result[fieldName] = field.Interface()
	}

	return result
}

// RuleBuilder provides fluent API for building custom rules
type RuleBuilder struct {
	rule *Rule
}

// NewRule creates a new rule builder
func NewRule(name string) *RuleBuilder {
	return &RuleBuilder{
		rule: &Rule{Name: name},
	}
}

// WithMessage sets a custom message for the rule
func (rb *RuleBuilder) WithMessage(message string) *RuleBuilder {
	rb.rule.Message = message
	return rb
}

// WithFunc sets the validation function
func (rb *RuleBuilder) WithFunc(fn RuleFunc) *RuleBuilder {
	rb.rule.Func = fn
	return rb
}

// Build returns the built rule
func (rb *RuleBuilder) Build() ValidationRule {
	return rb.rule
}

// Rule helper functions for common Laravel rules
func Required() ValidationRule {
	return &Rule{
		Name: "required",
		Func: func(value any, data map[string]any) error {
			if value == nil || value == "" {
				return fmt.Errorf("the field is required")
			}
			return nil
		},
	}
}

func String() ValidationRule {
	return &Rule{
		Name: "string",
		Func: func(value any, data map[string]any) error {
			if value == nil {
				return nil
			}
			if _, ok := value.(string); !ok {
				return fmt.Errorf("the field must be a string")
			}
			return nil
		},
	}
}

func Max(length int) ValidationRule {
	return &Rule{
		Name: "max",
		Func: func(value any, data map[string]any) error {
			if value == nil {
				return nil
			}

			switch v := value.(type) {
			case string:
				if len(v) > length {
					return fmt.Errorf("the field may not be greater than %d characters", length)
				}
			case []any:
				if len(v) > length {
					return fmt.Errorf("the field may not have more than %d items", length)
				}
			}
			return nil
		},
	}
}

func Min(length int) ValidationRule {
	return &Rule{
		Name: "min",
		Func: func(value any, data map[string]any) error {
			if value == nil {
				return nil
			}

			switch v := value.(type) {
			case string:
				if len(v) < length {
					return fmt.Errorf("the field must be at least %d characters", length)
				}
			case []any:
				if len(v) < length {
					return fmt.Errorf("the field must have at least %d items", length)
				}
			}
			return nil
		},
	}
}

// Exists creates a database existence rule (placeholder for now)
func Exists(table, column string) ValidationRule {
	return &Rule{
		Name: "exists",
		Func: func(value any, data map[string]any) error {
			// TODO: Implement database existence check
			return nil
		},
	}
}

// Unique creates a database uniqueness rule (placeholder for now)
func Unique(table, column string) ValidationRule {
	return &Rule{
		Name: "unique",
		Func: func(value any, data map[string]any) error {
			// TODO: Implement database uniqueness check
			return nil
		},
	}
}
