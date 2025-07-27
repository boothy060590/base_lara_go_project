# Smart File Processing & Memory Mapping Implementation Strategy

## Overview

This document outlines the implementation strategy for Phase 3.2: Smart File Processing with Memory Mapping. The system provides intelligent, config-driven file processing with automatic optimization selection while allowing developer override when needed.

## Architecture Philosophy

### Core Principles
- **Smart Defaults**: System automatically chooses optimal processing strategy
- **Developer Control**: Override capabilities when developers know better
- **Extensible**: Custom file types and processing strategies
- **Performance Aware**: Considers system resources and file characteristics
- **Configurable**: Fine-tune behavior per environment and use case
- **Laravel Familiar**: Easy-to-use facade API following Laravel patterns

### Comparison to Existing Solutions
Similar to [League CSV](https://github.com/thephpleague/csv) but with built-in smart memory optimization:
- Automatic strategy selection based on file size, type, and system resources
- Memory mapping for large files without manual configuration
- Streaming processing for medium files
- Chunked processing for small files
- Developer override capabilities

## Core Architecture

### Smart File Processor
```go
// Go Core: Smart file processor with automatic optimization
type SmartFileProcessor[T any] struct {
    config     *FileProcessingConfig
    strategies map[string]ProcessingStrategy
    autoDetect *AutoDetectionEngine
    registry   *FileTypeRegistry
}

// Laravel Core: Developer-friendly facade
type FileProcessor struct {
    processor *go_core.SmartFileProcessor[any]
}
```

### Processing Strategies
```go
type ProcessingStrategy interface {
    Process(filePath string, config *FileProcessingConfig) error
    Supports(filePath string) bool
    GetMemoryUsage() int64
    GetPerformanceMetrics() *PerformanceMetrics
}

// Available Strategies:
// - MemoryMappedStrategy: For large files (>100MB)
// - StreamingStrategy: For medium files (10MB-100MB)
// - ChunkedStrategy: For small files (1MB-10MB)
// - StandardStrategy: For very small files (<1MB)
```

## Smart Auto-Detection Engine

### Decision Tree
```go
type AutoDetectionEngine struct {
    config *FileProcessingConfig
}

func (a *AutoDetectionEngine) SelectStrategy(filePath string, fileInfo os.FileInfo) ProcessingStrategy {
    switch {
    case a.shouldUseMemoryMapping(filePath, fileInfo):
        return &MemoryMappedStrategy{}
    case a.shouldUseStreaming(filePath, fileInfo):
        return &StreamingStrategy{}
    case a.shouldUseChunked(filePath, fileInfo):
        return &ChunkedStrategy{}
    default:
        return &StandardStrategy{}
    }
}
```

### Auto-Detection Criteria

#### Memory Mapping Decision
```go
func (a *AutoDetectionEngine) shouldUseMemoryMapping(filePath string, fileInfo os.FileInfo) bool {
    return fileInfo.Size() > a.config.MemoryMappingThreshold && // File size > threshold
           a.isMemoryMappedType(filePath) &&                    // File type supports memory mapping
           a.hasAvailableMemory(fileInfo.Size()) &&             // Sufficient system memory
           a.isSequentialAccess(filePath)                       // Sequential access pattern
}
```

#### Streaming Decision
```go
func (a *AutoDetectionEngine) shouldUseStreaming(filePath string, fileInfo os.FileInfo) bool {
    return fileInfo.Size() > a.config.StreamingThreshold &&     // File size > threshold
           a.isStreamingType(filePath) &&                       // File type supports streaming
           a.hasModerateMemory(fileInfo.Size())                 // Moderate memory available
}
```

#### Chunked Decision
```go
func (a *AutoDetectionEngine) shouldUseChunked(filePath string, fileInfo os.FileInfo) bool {
    return fileInfo.Size() > a.config.ChunkedThreshold &&       // File size > threshold
           a.isChunkedType(filePath) &&                         // File type supports chunking
           a.hasLimitedMemory(fileInfo.Size())                  // Limited memory available
}
```

## Configuration System

### Configuration Structure
```go
// api/config/file_processing.go
type FileProcessingConfig struct {
    // Smart auto-detection settings
    AutoDetection AutoDetectionConfig `mapstructure:"auto_detection"`
    
    // Memory mapping configuration
    MemoryMapping MemoryMappingConfig `mapstructure:"memory_mapping"`
    
    // File type specific settings
    FileTypes map[string]FileTypeConfig `mapstructure:"file_types"`
    
    // Processing strategies
    Strategies StrategyConfig `mapstructure:"strategies"`
}
```

### Auto-Detection Configuration
```go
type AutoDetectionConfig struct {
    Enabled bool `mapstructure:"enabled"`
    
    // Automatic thresholds (in bytes)
    MemoryMappingThreshold int64 `mapstructure:"memory_mapping_threshold"` // 100MB default
    StreamingThreshold     int64 `mapstructure:"streaming_threshold"`     // 10MB default
    ChunkedThreshold       int64 `mapstructure:"chunked_threshold"`       // 1MB default
    
    // Smart file type detection
    MemoryMappedTypes []string `mapstructure:"memory_mapped_types"` // ["csv", "json", "log"]
    StreamingTypes    []string `mapstructure:"streaming_types"`     // ["xml", "yaml"]
    ChunkedTypes      []string `mapstructure:"chunked_types"`       // ["txt", "md"]
    
    // System resource awareness
    MaxMemoryUsage    float64 `mapstructure:"max_memory_usage"`    // 0.8 (80% of available)
    EnableNUMA        bool    `mapstructure:"enable_numa"`         // true
    EnableCompression bool    `mapstructure:"enable_compression"`  // false
}
```

### Memory Mapping Configuration
```go
type MemoryMappingConfig struct {
    Enabled           bool   `mapstructure:"enabled"`
    ChunkSize         int    `mapstructure:"chunk_size"`         // 1MB chunks
    PrefetchSize      int    `mapstructure:"prefetch_size"`      // 4MB prefetch
    CleanupInterval   int    `mapstructure:"cleanup_interval"`   // 5 minutes
    MaxConcurrentMaps int    `mapstructure:"max_concurrent_maps"` // 10 files
    EnableNUMA        bool   `mapstructure:"enable_numa"`        // true
}
```

### File Type Configuration
```go
type FileTypeConfig struct {
    Strategy          string `mapstructure:"strategy"`           // "auto", "memory_mapped", "streaming", "chunked"
    ChunkSize         int    `mapstructure:"chunk_size"`
    MemoryMapping     bool   `mapstructure:"memory_mapping"`
    EnableCompression bool   `mapstructure:"enable_compression"`
    Delimiter         string `mapstructure:"delimiter"`          // For CSV files
    Encoding          string `mapstructure:"encoding"`           // UTF-8, etc.
}
```

## Processing Strategies Implementation

### Memory-Mapped Strategy
```go
type MemoryMappedStrategy struct {
    config *MemoryMappingConfig
}

func (m *MemoryMappedStrategy) Process(filePath string, config *FileProcessingConfig) error {
    // Memory-mapped file processing
    file, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
    if err != nil {
        return err
    }
    defer file.Close()
    
    // Memory map the file
    mmap, err := mmap.Map(file, mmap.RDONLY, 0)
    if err != nil {
        return err
    }
    defer mmap.Unmap()
    
    // Process in chunks with zero-copy
    return m.processChunks(mmap, config)
}

func (m *MemoryMappedStrategy) processChunks(data []byte, config *FileProcessingConfig) error {
    chunkSize := m.config.ChunkSize
    for offset := 0; offset < len(data); offset += chunkSize {
        end := offset + chunkSize
        if end > len(data) {
            end = len(data)
        }
        
        chunk := data[offset:end]
        if err := m.processChunk(chunk, config); err != nil {
            return err
        }
    }
    return nil
}
```

### Streaming Strategy
```go
type StreamingStrategy struct {
    config *StreamingConfig
}

func (s *StreamingStrategy) Process(filePath string, config *FileProcessingConfig) error {
    // Streaming processing with buffered I/O
    file, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer file.Close()
    
    // Use buffered reader for optimal performance
    reader := bufio.NewReaderSize(file, s.config.BufferSize)
    return s.processStream(reader, config)
}

func (s *StreamingStrategy) processStream(reader *bufio.Reader, config *FileProcessingConfig) error {
    // Process file line by line or record by record
    for {
        line, err := reader.ReadString('\n')
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }
        
        if err := s.processLine(line, config); err != nil {
            return err
        }
    }
    return nil
}
```

### Chunked Strategy
```go
type ChunkedStrategy struct {
    config *ChunkedConfig
}

func (c *ChunkedStrategy) Process(filePath string, config *FileProcessingConfig) error {
    // Chunked processing with controlled memory usage
    file, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer file.Close()
    
    buffer := make([]byte, c.config.ChunkSize)
    for {
        n, err := file.Read(buffer)
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }
        
        if err := c.processChunk(buffer[:n], config); err != nil {
            return err
        }
    }
    return nil
}
```

## Developer Override System

### Flexible API Design
```go
// Laravel Core: Developer-friendly API
type FileProcessor struct {
    processor *go_core.SmartFileProcessor[any]
}

// Let the system decide (smart auto-detection)
func (fp *FileProcessor) Process(filePath string) error {
    return fp.processor.Process(filePath)
}

// Override with specific strategy
func (fp *FileProcessor) ProcessWithStrategy(filePath string, strategy string) error {
    return fp.processor.ProcessWithStrategy(filePath, strategy)
}

// Override with custom configuration
func (fp *FileProcessor) ProcessWithConfig(filePath string, config *FileProcessingConfig) error {
    return fp.processor.ProcessWithConfig(filePath, config)
}

// Force memory mapping regardless of auto-detection
func (fp *FileProcessor) ProcessMemoryMapped(filePath string) error {
    return fp.processor.ProcessWithStrategy(filePath, "memory_mapped")
}

// Force streaming regardless of auto-detection
func (fp *FileProcessor) ProcessStreaming(filePath string) error {
    return fp.processor.ProcessWithStrategy(filePath, "streaming")
}

// Force chunked processing regardless of auto-detection
func (fp *FileProcessor) ProcessChunked(filePath string) error {
    return fp.processor.ProcessWithStrategy(filePath, "chunked")
}
```

## File Type Extensions

### Extensible File Type System
```go
type FileTypeRegistry struct {
    processors map[string]FileTypeProcessor
    config     *FileProcessingConfig
}

func (r *FileTypeRegistry) Register(fileType string, processor FileTypeProcessor) {
    r.processors[fileType] = processor
}

// Custom file type processor
type CustomFileProcessor struct {
    config *FileProcessingConfig
}

func (c *CustomFileProcessor) Process(filePath string) error {
    // Custom processing logic
    // Can still use memory mapping, streaming, etc.
    return nil
}

func (c *CustomFileProcessor) GetRecommendedStrategy(filePath string) string {
    // Custom logic to recommend processing strategy
    return "memory_mapped"
}
```

## Usage Examples

### Automatic Smart Processing
```go
// Developer just calls Process() - system decides
processor := FileProcessor{}
err := processor.Process("large_dataset.csv") // Auto-detects: large file + CSV = memory mapping
err = processor.Process("small_config.json")  // Auto-detects: small file = standard processing
err = processor.Process("medium_log.txt")     // Auto-detects: medium file = streaming
```

### Developer Override
```go
// Developer knows better than auto-detection
processor := FileProcessor{}
err := processor.ProcessMemoryMapped("medium_file.csv") // Force memory mapping
err = processor.ProcessWithStrategy("custom_file.xyz", "streaming") // Custom strategy
err = processor.ProcessChunked("large_file.json") // Force chunked processing
```

### Custom File Types
```go
// Register custom file type with smart processing
registry := FileTypeRegistry{}
registry.Register("xyz", &CustomXYZProcessor{})

// System will use custom processor but still apply smart optimization
processor := FileProcessor{}
err := processor.Process("data.xyz") // Uses custom processor + smart optimization
```

### Configuration Override
```go
// Override configuration for specific processing
config := &FileProcessingConfig{
    AutoDetection: AutoDetectionConfig{
        MemoryMappingThreshold: 50 * 1024 * 1024, // 50MB instead of 100MB
    },
    MemoryMapping: MemoryMappingConfig{
        ChunkSize: 2 * 1024 * 1024, // 2MB chunks instead of 1MB
    },
}

processor := FileProcessor{}
err := processor.ProcessWithConfig("file.csv", config)
```

## Performance Monitoring

### Metrics Collection
```go
type PerformanceMetrics struct {
    ProcessingTime    time.Duration
    MemoryUsage       int64
    StrategyUsed      string
    FileSize          int64
    ChunksProcessed   int
    ErrorsEncountered int
}

type MetricsCollector struct {
    metrics map[string]*PerformanceMetrics
}

func (m *MetricsCollector) Record(filePath string, metrics *PerformanceMetrics) {
    m.metrics[filePath] = metrics
}

func (m *MetricsCollector) GetRecommendations() []string {
    // Analyze metrics and provide optimization recommendations
    return []string{
        "Consider memory mapping for files > 50MB",
        "Streaming strategy shows 30% better performance for XML files",
        "Chunked processing recommended for memory-constrained environments",
    }
}
```

## Implementation Phases

### Phase 3.2.1: Core Architecture
- [ ] Implement `SmartFileProcessor` with auto-detection engine
- [ ] Create processing strategy interfaces and base implementations
- [ ] Add configuration structure and loading
- [ ] Implement file type registry system

### Phase 3.2.2: Processing Strategies
- [ ] Implement `MemoryMappedStrategy` with mmap integration
- [ ] Implement `StreamingStrategy` with buffered I/O
- [ ] Implement `ChunkedStrategy` with controlled memory usage
- [ ] Implement `StandardStrategy` for small files

### Phase 3.2.3: Auto-Detection Engine
- [ ] Implement decision tree logic
- [ ] Add system resource detection
- [ ] Implement file type detection
- [ ] Add performance monitoring and metrics

### Phase 3.2.4: Laravel Integration
- [ ] Create `FileProcessor` facade
- [ ] Implement service provider for auto-registration
- [ ] Add configuration to core app service provider
- [ ] Create comprehensive unit and integration tests

### Phase 3.2.5: Extensibility
- [ ] Implement custom file type registration
- [ ] Add developer override capabilities
- [ ] Create performance monitoring dashboard
- [ ] Add documentation and examples

## Configuration Examples

### Default Configuration
```go
// api/config/file_processing.go
var DefaultFileProcessingConfig = FileProcessingConfig{
    AutoDetection: AutoDetectionConfig{
        Enabled:                 true,
        MemoryMappingThreshold:  100 * 1024 * 1024, // 100MB
        StreamingThreshold:      10 * 1024 * 1024,  // 10MB
        ChunkedThreshold:        1 * 1024 * 1024,   // 1MB
        MemoryMappedTypes:       []string{"csv", "json", "log"},
        StreamingTypes:          []string{"xml", "yaml"},
        ChunkedTypes:            []string{"txt", "md"},
        MaxMemoryUsage:          0.8, // 80%
        EnableNUMA:              true,
        EnableCompression:       false,
    },
    MemoryMapping: MemoryMappingConfig{
        Enabled:           true,
        ChunkSize:         1 * 1024 * 1024, // 1MB
        PrefetchSize:      4 * 1024 * 1024, // 4MB
        CleanupInterval:   300, // 5 minutes
        MaxConcurrentMaps: 10,
        EnableNUMA:        true,
    },
    FileTypes: map[string]FileTypeConfig{
        "csv": {
            Strategy:          "auto",
            ChunkSize:         1 * 1024 * 1024,
            MemoryMapping:     true,
            EnableCompression: false,
            Delimiter:         ",",
            Encoding:          "UTF-8",
        },
        "json": {
            Strategy:          "auto",
            ChunkSize:         1 * 1024 * 1024,
            MemoryMapping:     true,
            EnableCompression: false,
            Encoding:          "UTF-8",
        },
    },
}
```

### Environment-Specific Configuration
```go
// Production environment
var ProductionFileProcessingConfig = FileProcessingConfig{
    AutoDetection: AutoDetectionConfig{
        MemoryMappingThreshold: 50 * 1024 * 1024, // Lower threshold for production
        MaxMemoryUsage:         0.6, // More conservative memory usage
    },
    MemoryMapping: MemoryMappingConfig{
        MaxConcurrentMaps: 5, // Fewer concurrent maps
    },
}

// Development environment
var DevelopmentFileProcessingConfig = FileProcessingConfig{
    AutoDetection: AutoDetectionConfig{
        MemoryMappingThreshold: 200 * 1024 * 1024, // Higher threshold for development
        MaxMemoryUsage:         0.9, // More aggressive memory usage
    },
    MemoryMapping: MemoryMappingConfig{
        MaxConcurrentMaps: 20, // More concurrent maps
    },
}
```

## Benefits and Expected Performance Gains

### Memory Optimization
- **Memory Mapping**: 50-80% reduction in memory usage for large files
- **Streaming**: 30-50% reduction in memory usage for medium files
- **Chunked Processing**: 20-40% reduction in memory usage for small files

### Performance Improvements
- **Zero-Copy Operations**: Memory mapping eliminates data copying
- **OS-Level Optimization**: Leverages OS caching and prefetching
- **NUMA Awareness**: Optimal memory placement on multi-socket systems
- **Concurrent Processing**: Multiple files can be processed simultaneously

### Developer Experience
- **Zero Configuration**: Works out of the box with sensible defaults
- **Automatic Optimization**: No manual strategy selection required
- **Override Capabilities**: Developer control when needed
- **Extensible**: Custom file types and processing strategies
- **Laravel Familiar**: Easy-to-use facade API

## Integration with Existing Systems

### Object Pools Integration
- Memory-mapped files can use object pools for chunk processing
- Streaming strategies can reuse buffer pools
- Chunked processing can use pooled memory buffers

### NUMA Integration
- Memory mapping respects NUMA node placement
- Object pools can be NUMA-aware
- Processing strategies can be optimized for specific NUMA nodes

### Performance Monitoring Integration
- Metrics collection integrates with existing monitoring
- Performance recommendations feed into optimization engine
- Strategy effectiveness tracking for continuous improvement

This implementation strategy provides a comprehensive, intelligent file processing system that automatically optimizes performance while maintaining developer control and extensibility. 