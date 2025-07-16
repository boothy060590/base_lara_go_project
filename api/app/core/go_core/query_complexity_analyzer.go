package go_core

import (
	"fmt"
	"strings"
	"time"
)

// QueryComplexityAnalyzer automatically determines the optimal optimization level
type QueryComplexityAnalyzer struct {
	config *ComplexityAnalyzerConfig
	stats  *ComplexityAnalyzerStats
}

// ComplexityAnalyzerConfig configures the complexity analyzer
type ComplexityAnalyzerConfig struct {
	// Thresholds for auto-escalation
	MaxSimpleConditions     int           `json:"max_simple_conditions"`
	MaxSimpleOrderBy        int           `json:"max_simple_order_by"`
	MaxBalancedConditions   int           `json:"max_balanced_conditions"`
	MaxBalancedOrderBy      int           `json:"max_balanced_order_by"`
	MaxBalancedJoins        int           `json:"max_balanced_joins"`
	MaxBalancedInValues     int           `json:"max_balanced_in_values"`
	
	// Performance thresholds
	MaxSimpleLatency        time.Duration `json:"max_simple_latency"`
	MaxBalancedLatency      time.Duration `json:"max_balanced_latency"`
	MaxSimpleResultSize     int           `json:"max_simple_result_size"`
	MaxBalancedResultSize   int           `json:"max_balanced_result_size"`
	
	// Auto-learning settings
	EnableAutoLearning      bool          `json:"enable_auto_learning"`
	LearningWindowSize      int           `json:"learning_window_size"`
	LearningThreshold       float64       `json:"learning_threshold"`
	
	// Complexity keywords
	ComplexKeywords         []string      `json:"complex_keywords"`
	SimpleOperations        []string      `json:"simple_operations"`
}

// ComplexityAnalyzerStats tracks analyzer performance
type ComplexityAnalyzerStats struct {
	QueriesAnalyzed         int64         `json:"queries_analyzed"`
	SimpleQueriesDetected   int64         `json:"simple_queries_detected"`
	BalancedQueriesDetected int64         `json:"balanced_queries_detected"`
	ComplexQueriesDetected  int64         `json:"complex_queries_detected"`
	AutoEscalations         int64         `json:"auto_escalations"`
	CorrectPredictions      int64         `json:"correct_predictions"`
	IncorrectPredictions    int64         `json:"incorrect_predictions"`
	LastAnalysis            time.Time     `json:"last_analysis"`
}

// QueryAnalysisResult contains the analysis results
type QueryAnalysisResult struct {
	RecommendedComplexity QueryComplexity `json:"recommended_complexity"`
	Confidence            float64         `json:"confidence"`
	Reasons               []string        `json:"reasons"`
	Metrics               map[string]any  `json:"metrics"`
	ShouldEscalate        bool            `json:"should_escalate"`
}

// NewQueryComplexityAnalyzer creates a new complexity analyzer
func NewQueryComplexityAnalyzer(config *ComplexityAnalyzerConfig) *QueryComplexityAnalyzer {
	if config == nil {
		config = DefaultComplexityAnalyzerConfig()
	}
	
	return &QueryComplexityAnalyzer{
		config: config,
		stats:  &ComplexityAnalyzerStats{},
	}
}

// DefaultComplexityAnalyzerConfig returns default configuration
func DefaultComplexityAnalyzerConfig() *ComplexityAnalyzerConfig {
	return &ComplexityAnalyzerConfig{
		MaxSimpleConditions:     1,
		MaxSimpleOrderBy:        1,
		MaxBalancedConditions:   3,
		MaxBalancedOrderBy:      2,
		MaxBalancedJoins:        1,
		MaxBalancedInValues:     10,
		MaxSimpleLatency:        5 * time.Millisecond,
		MaxBalancedLatency:      50 * time.Millisecond,
		MaxSimpleResultSize:     100,
		MaxBalancedResultSize:   1000,
		EnableAutoLearning:      true,
		LearningWindowSize:      1000,
		LearningThreshold:       0.8,
		ComplexKeywords: []string{
			"GROUP BY", "HAVING", "UNION", "SUBQUERY", "EXISTS", "NOT EXISTS",
			"WINDOW", "PARTITION", "RECURSIVE", "CTE", "CROSS JOIN", "FULL OUTER JOIN",
			"REGEXP", "FULLTEXT", "MATCH", "AGAINST", "JSON_EXTRACT", "JSON_UNQUOTE",
		},
		SimpleOperations: []string{
			"SELECT", "INSERT", "UPDATE", "DELETE", "WHERE", "ORDER BY", "LIMIT",
		},
	}
}

// AnalyzeQueryComplexity analyzes a query and recommends optimization level
func (qca *QueryComplexityAnalyzer) AnalyzeQueryComplexity(query QueryContext) QueryAnalysisResult {
	qca.stats.QueriesAnalyzed++
	qca.stats.LastAnalysis = time.Now()
	
	result := QueryAnalysisResult{
		RecommendedComplexity: SimpleQuery,
		Confidence:            0.0,
		Reasons:               []string{},
		Metrics:               make(map[string]any),
	}
	
	// Analyze different aspects of the query
	result = qca.analyzeOperationType(query, result)
	result = qca.analyzeConditions(query, result)
	result = qca.analyzeJoins(query, result)
	result = qca.analyzeAggregations(query, result)
	result = qca.analyzeResultSize(query, result)
	result = qca.analyzePerformanceHistory(query, result)
	result = qca.analyzeComplexKeywords(query, result)
	
	// Calculate final complexity and confidence
	result = qca.calculateFinalComplexity(result)
	
	// Update stats
	qca.updateStats(result)
	
	return result
}

// analyzeOperationType analyzes the type of operation
func (qca *QueryComplexityAnalyzer) analyzeOperationType(query QueryContext, result QueryAnalysisResult) QueryAnalysisResult {
	operation := strings.ToUpper(query.Operation)
	
	switch operation {
	case "FIND", "FINDBY", "EXISTS", "COUNT":
		result.Reasons = append(result.Reasons, "Simple operation: "+operation)
		result.Metrics["operation_complexity"] = 1
		
	case "WHERE", "PAGINATE":
		result.Reasons = append(result.Reasons, "Balanced operation: "+operation)
		result.Metrics["operation_complexity"] = 2
		
	case "RAW", "BULK", "STREAM", "REPORT":
		result.Reasons = append(result.Reasons, "Complex operation: "+operation)
		result.Metrics["operation_complexity"] = 3
		result.RecommendedComplexity = ComplexQueryLevel
		
	default:
		result.Metrics["operation_complexity"] = 2
	}
	
	return result
}

// analyzeConditions analyzes WHERE conditions
func (qca *QueryComplexityAnalyzer) analyzeConditions(query QueryContext, result QueryAnalysisResult) QueryAnalysisResult {
	conditionCount := len(query.Conditions)
	
	if conditionCount == 0 {
		result.Reasons = append(result.Reasons, "No WHERE conditions")
		result.Metrics["condition_complexity"] = 1
		return result
	}
	
	// Analyze condition complexity
	complexConditions := 0
	inOperators := 0
	
	for field, value := range query.Conditions {
		// Check for complex operators
		if strings.Contains(field, "_in") {
			inOperators++
			if values, ok := value.([]any); ok && len(values) > qca.config.MaxBalancedInValues {
				complexConditions++
				result.Reasons = append(result.Reasons, fmt.Sprintf("Large IN operator with %d values", len(values)))
			}
		}
		
		// Check for complex field patterns
		if strings.Contains(field, "_like") || strings.Contains(field, "_regexp") || strings.Contains(field, "_json") {
			complexConditions++
			result.Reasons = append(result.Reasons, "Complex field operator: "+field)
		}
	}
	
	// Determine complexity based on condition count and types
	if conditionCount <= qca.config.MaxSimpleConditions && complexConditions == 0 {
		result.Reasons = append(result.Reasons, fmt.Sprintf("Simple conditions: %d", conditionCount))
		result.Metrics["condition_complexity"] = 1
	} else if conditionCount <= qca.config.MaxBalancedConditions && complexConditions <= 1 {
		result.Reasons = append(result.Reasons, fmt.Sprintf("Balanced conditions: %d", conditionCount))
		result.Metrics["condition_complexity"] = 2
		if result.RecommendedComplexity < MediumQuery {
			result.RecommendedComplexity = MediumQuery
		}
	} else {
		result.Reasons = append(result.Reasons, fmt.Sprintf("Complex conditions: %d (complex: %d)", conditionCount, complexConditions))
		result.Metrics["condition_complexity"] = 3
		result.RecommendedComplexity = ComplexQueryLevel
	}
	
	return result
}

// analyzeJoins analyzes JOIN operations
func (qca *QueryComplexityAnalyzer) analyzeJoins(query QueryContext, result QueryAnalysisResult) QueryAnalysisResult {
	joinCount := len(query.Joins)
	
	if joinCount == 0 {
		result.Metrics["join_complexity"] = 1
		return result
	}
	
	// Analyze join types
	complexJoins := 0
	for _, join := range query.Joins {
		joinUpper := strings.ToUpper(join)
		if strings.Contains(joinUpper, "FULL OUTER") || strings.Contains(joinUpper, "CROSS") || strings.Contains(joinUpper, "SELF") {
			complexJoins++
		}
	}
	
	if joinCount <= qca.config.MaxBalancedJoins && complexJoins == 0 {
		result.Reasons = append(result.Reasons, fmt.Sprintf("Balanced joins: %d", joinCount))
		result.Metrics["join_complexity"] = 2
		if result.RecommendedComplexity < MediumQuery {
			result.RecommendedComplexity = MediumQuery
		}
	} else {
		result.Reasons = append(result.Reasons, fmt.Sprintf("Complex joins: %d (complex: %d)", joinCount, complexJoins))
		result.Metrics["join_complexity"] = 3
		result.RecommendedComplexity = ComplexQueryLevel
	}
	
	return result
}

// analyzeAggregations analyzes GROUP BY, HAVING, aggregation functions
func (qca *QueryComplexityAnalyzer) analyzeAggregations(query QueryContext, result QueryAnalysisResult) QueryAnalysisResult {
	hasGroupBy := len(query.GroupBy) > 0
	hasHaving := len(query.Having) > 0
	hasAggregations := query.HasAggregations
	
	if hasGroupBy || hasHaving || hasAggregations {
		result.Reasons = append(result.Reasons, "Contains aggregations, GROUP BY, or HAVING")
		result.Metrics["aggregation_complexity"] = 3
		result.RecommendedComplexity = ComplexQueryLevel
	} else {
		result.Metrics["aggregation_complexity"] = 1
	}
	
	return result
}

// analyzeResultSize analyzes expected result size
func (qca *QueryComplexityAnalyzer) analyzeResultSize(query QueryContext, result QueryAnalysisResult) QueryAnalysisResult {
	expectedSize := query.ExpectedResultSize
	
	if expectedSize <= qca.config.MaxSimpleResultSize {
		result.Reasons = append(result.Reasons, fmt.Sprintf("Small result set: %d", expectedSize))
		result.Metrics["result_size_complexity"] = 1
	} else if expectedSize <= qca.config.MaxBalancedResultSize {
		result.Reasons = append(result.Reasons, fmt.Sprintf("Medium result set: %d", expectedSize))
		result.Metrics["result_size_complexity"] = 2
		if result.RecommendedComplexity < MediumQuery {
			result.RecommendedComplexity = MediumQuery
		}
	} else {
		result.Reasons = append(result.Reasons, fmt.Sprintf("Large result set: %d", expectedSize))
		result.Metrics["result_size_complexity"] = 3
		result.RecommendedComplexity = ComplexQueryLevel
	}
	
	return result
}

// analyzePerformanceHistory analyzes historical performance
func (qca *QueryComplexityAnalyzer) analyzePerformanceHistory(query QueryContext, result QueryAnalysisResult) QueryAnalysisResult {
	if query.HistoricalLatency > 0 {
		if query.HistoricalLatency <= qca.config.MaxSimpleLatency {
			result.Reasons = append(result.Reasons, fmt.Sprintf("Fast historical performance: %v", query.HistoricalLatency))
			result.Metrics["performance_complexity"] = 1
		} else if query.HistoricalLatency <= qca.config.MaxBalancedLatency {
			result.Reasons = append(result.Reasons, fmt.Sprintf("Moderate historical performance: %v", query.HistoricalLatency))
			result.Metrics["performance_complexity"] = 2
			if result.RecommendedComplexity < MediumQuery {
				result.RecommendedComplexity = MediumQuery
			}
		} else {
			result.Reasons = append(result.Reasons, fmt.Sprintf("Slow historical performance: %v", query.HistoricalLatency))
			result.Metrics["performance_complexity"] = 3
			result.RecommendedComplexity = ComplexQueryLevel
		}
	}
	
	return result
}

// analyzeComplexKeywords analyzes for complex SQL keywords
func (qca *QueryComplexityAnalyzer) analyzeComplexKeywords(query QueryContext, result QueryAnalysisResult) QueryAnalysisResult {
	if query.RawSQL == "" {
		result.Metrics["keyword_complexity"] = 1
		return result
	}
	
	sqlUpper := strings.ToUpper(query.RawSQL)
	complexKeywordCount := 0
	
	for _, keyword := range qca.config.ComplexKeywords {
		if strings.Contains(sqlUpper, keyword) {
			complexKeywordCount++
			result.Reasons = append(result.Reasons, "Contains complex keyword: "+keyword)
		}
	}
	
	if complexKeywordCount == 0 {
		result.Metrics["keyword_complexity"] = 1
	} else if complexKeywordCount <= 2 {
		result.Reasons = append(result.Reasons, fmt.Sprintf("Some complex keywords: %d", complexKeywordCount))
		result.Metrics["keyword_complexity"] = 2
		if result.RecommendedComplexity < MediumQuery {
			result.RecommendedComplexity = MediumQuery
		}
	} else {
		result.Reasons = append(result.Reasons, fmt.Sprintf("Many complex keywords: %d", complexKeywordCount))
		result.Metrics["keyword_complexity"] = 3
		result.RecommendedComplexity = ComplexQueryLevel
	}
	
	return result
}

// calculateFinalComplexity calculates the final complexity and confidence
func (qca *QueryComplexityAnalyzer) calculateFinalComplexity(result QueryAnalysisResult) QueryAnalysisResult {
	// Calculate weighted complexity score
	weights := map[string]float64{
		"operation_complexity":    0.3,
		"condition_complexity":    0.25,
		"join_complexity":         0.2,
		"aggregation_complexity":  0.15,
		"result_size_complexity":  0.05,
		"performance_complexity":  0.03,
		"keyword_complexity":      0.02,
	}
	
	totalScore := 0.0
	maxScore := 0.0
	
	for metric, complexity := range result.Metrics {
		if weight, exists := weights[metric]; exists {
			if complexityVal, ok := complexity.(int); ok {
				totalScore += weight * float64(complexityVal)
				maxScore += weight * 3.0 // Max complexity is 3
			}
		}
	}
	
	// Calculate confidence (0.0 to 1.0)
	if maxScore > 0 {
		result.Confidence = totalScore / maxScore
	}
	
	// Determine final complexity based on score
	normalizedScore := totalScore / maxScore
	
	if normalizedScore <= 0.4 {
		result.RecommendedComplexity = SimpleQuery
	} else if normalizedScore <= 0.7 {
		result.RecommendedComplexity = MediumQuery
	} else {
		result.RecommendedComplexity = ComplexQueryLevel
	}
	
	// High confidence threshold check
	if result.Confidence < 0.6 {
		result.ShouldEscalate = true
		result.Reasons = append(result.Reasons, "Low confidence, consider escalating")
	}
	
	return result
}

// updateStats updates analyzer statistics
func (qca *QueryComplexityAnalyzer) updateStats(result QueryAnalysisResult) {
	switch result.RecommendedComplexity {
	case SimpleQuery:
		qca.stats.SimpleQueriesDetected++
	case MediumQuery:
		qca.stats.BalancedQueriesDetected++
	case ComplexQueryLevel:
		qca.stats.ComplexQueriesDetected++
	}
	
	if result.ShouldEscalate {
		qca.stats.AutoEscalations++
	}
}

// GetStats returns analyzer statistics
func (qca *QueryComplexityAnalyzer) GetStats() *ComplexityAnalyzerStats {
	return qca.stats
}

// UpdateConfig updates analyzer configuration
func (qca *QueryComplexityAnalyzer) UpdateConfig(config *ComplexityAnalyzerConfig) {
	qca.config = config
}

// QueryContext provides context for query analysis
type QueryContext struct {
	Operation            string            `json:"operation"`
	Conditions           map[string]any    `json:"conditions"`
	Joins                []string          `json:"joins"`
	GroupBy              []string          `json:"group_by"`
	Having               []string          `json:"having"`
	OrderBy              []string          `json:"order_by"`
	Limit                int               `json:"limit"`
	Offset               int               `json:"offset"`
	RawSQL               string            `json:"raw_sql"`
	ExpectedResultSize   int               `json:"expected_result_size"`
	HistoricalLatency    time.Duration     `json:"historical_latency"`
	HasAggregations      bool              `json:"has_aggregations"`
	TableName            string            `json:"table_name"`
	IsTransaction        bool              `json:"is_transaction"`
	IsBulkOperation      bool              `json:"is_bulk_operation"`
	AdditionalContext    map[string]any    `json:"additional_context"`
}

// Smart escalation functions
func (qca *QueryComplexityAnalyzer) ShouldEscalateToBalanced(query QueryContext) bool {
	result := qca.AnalyzeQueryComplexity(query)
	return result.RecommendedComplexity >= MediumQuery
}

func (qca *QueryComplexityAnalyzer) ShouldEscalateToComplex(query QueryContext) bool {
	result := qca.AnalyzeQueryComplexity(query)
	return result.RecommendedComplexity >= ComplexQueryLevel
}

func (qca *QueryComplexityAnalyzer) GetOptimalComplexity(query QueryContext) (QueryComplexity, float64) {
	result := qca.AnalyzeQueryComplexity(query)
	return result.RecommendedComplexity, result.Confidence
}