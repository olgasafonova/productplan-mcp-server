package productplan

import (
	"context"
	"sync"
)

// BatchConfig configures batch operation behavior.
type BatchConfig struct {
	Concurrency int  // Max concurrent API calls (0 = sequential)
	StopOnError bool // Stop all operations on first error
}

// DefaultBatchConfig returns sensible defaults for batch operations.
func DefaultBatchConfig() BatchConfig {
	return BatchConfig{
		Concurrency: 3, // Reasonable parallelism without overwhelming API
		StopOnError: false,
	}
}

// BatchResult holds results from a batch operation.
type BatchResult[T any] struct {
	Results []T
	Errors  []BatchError
}

// BatchError associates an error with its index in the batch.
type BatchError struct {
	Index int
	Key   string // Optional identifier (e.g., roadmap ID)
	Err   error
}

// HasErrors returns true if any errors occurred.
func (br *BatchResult[T]) HasErrors() bool {
	return len(br.Errors) > 0
}

// SuccessCount returns the number of successful operations.
func (br *BatchResult[T]) SuccessCount() int {
	return len(br.Results)
}

// ErrorCount returns the number of failed operations.
func (br *BatchResult[T]) ErrorCount() int {
	return len(br.Errors)
}

// BatchExecutor runs multiple operations with configurable concurrency.
type BatchExecutor struct {
	config BatchConfig
}

// NewBatchExecutor creates a new batch executor.
func NewBatchExecutor(config BatchConfig) *BatchExecutor {
	return &BatchExecutor{config: config}
}

// Execute runs the given functions with configured concurrency.
// Each function receives its index and should return a result or error.
func Execute[T any](ctx context.Context, config BatchConfig, fns []func(ctx context.Context) (T, error)) *BatchResult[T] {
	result := &BatchResult[T]{
		Results: make([]T, 0, len(fns)),
		Errors:  make([]BatchError, 0),
	}

	if len(fns) == 0 {
		return result
	}

	// Sequential execution
	if config.Concurrency <= 1 {
		for i, fn := range fns {
			if ctx.Err() != nil {
				result.Errors = append(result.Errors, BatchError{
					Index: i,
					Err:   ctx.Err(),
				})
				break
			}

			res, err := fn(ctx)
			if err != nil {
				result.Errors = append(result.Errors, BatchError{
					Index: i,
					Err:   err,
				})
				if config.StopOnError {
					break
				}
				continue
			}
			result.Results = append(result.Results, res)
		}
		return result
	}

	// Concurrent execution with semaphore
	type indexedResult struct {
		index  int
		result T
		err    error
	}

	resultChan := make(chan indexedResult, len(fns))
	sem := make(chan struct{}, config.Concurrency)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	for i, fn := range fns {
		wg.Add(1)
		go func(idx int, f func(ctx context.Context) (T, error)) {
			defer wg.Done()

			// Acquire semaphore
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				resultChan <- indexedResult{index: idx, err: ctx.Err()}
				return
			}

			res, err := f(ctx)
			if err != nil && config.StopOnError {
				cancel()
			}
			resultChan <- indexedResult{index: idx, result: res, err: err}
		}(i, fn)
	}

	// Close channel when all done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results maintaining order info
	indexed := make(map[int]indexedResult)
	for ir := range resultChan {
		indexed[ir.index] = ir
	}

	// Process in order
	for i := 0; i < len(fns); i++ {
		ir := indexed[i]
		if ir.err != nil {
			result.Errors = append(result.Errors, BatchError{
				Index: i,
				Err:   ir.err,
			})
		} else {
			result.Results = append(result.Results, ir.result)
		}
	}

	return result
}

// ExecuteWithKeys runs operations associated with string keys.
// Useful for operations like "get bars for each roadmap ID".
func ExecuteWithKeys[T any](ctx context.Context, config BatchConfig, keys []string, fn func(ctx context.Context, key string) (T, error)) *BatchResult[T] {
	fns := make([]func(ctx context.Context) (T, error), len(keys))
	for i, key := range keys {
		k := key // Capture for closure
		fns[i] = func(ctx context.Context) (T, error) {
			return fn(ctx, k)
		}
	}

	result := Execute(ctx, config, fns)

	// Add keys to errors for better debugging
	for i := range result.Errors {
		if result.Errors[i].Index < len(keys) {
			result.Errors[i].Key = keys[result.Errors[i].Index]
		}
	}

	return result
}

// Paginator handles paginated API responses.
type Paginator[T any] struct {
	PageSize int
	MaxPages int // 0 = unlimited
}

// NewPaginator creates a paginator with default settings.
func NewPaginator[T any]() *Paginator[T] {
	return &Paginator[T]{
		PageSize: 100,
		MaxPages: 0,
	}
}

// PaginatedResult holds all items from paginated fetches.
type PaginatedResult[T any] struct {
	Items      []T
	TotalPages int
	Error      error
}

// FetchAll retrieves all pages using the provided fetch function.
// The fetch function receives page number (1-indexed) and page size.
func (p *Paginator[T]) FetchAll(ctx context.Context, fetch func(ctx context.Context, page, pageSize int) ([]T, bool, error)) *PaginatedResult[T] {
	result := &PaginatedResult[T]{
		Items: make([]T, 0),
	}

	page := 1
	for {
		if ctx.Err() != nil {
			result.Error = ctx.Err()
			break
		}

		if p.MaxPages > 0 && page > p.MaxPages {
			break
		}

		items, hasMore, err := fetch(ctx, page, p.PageSize)
		if err != nil {
			result.Error = err
			break
		}

		result.Items = append(result.Items, items...)
		result.TotalPages = page

		if !hasMore || len(items) < p.PageSize {
			break
		}

		page++
	}

	return result
}

// Pipeline chains multiple batch operations together.
type Pipeline[T any] struct {
	steps []func(ctx context.Context, input T) (T, error)
}

// NewPipeline creates a new processing pipeline.
func NewPipeline[T any]() *Pipeline[T] {
	return &Pipeline[T]{
		steps: make([]func(ctx context.Context, input T) (T, error), 0),
	}
}

// AddStep adds a processing step to the pipeline.
func (p *Pipeline[T]) AddStep(step func(ctx context.Context, input T) (T, error)) *Pipeline[T] {
	p.steps = append(p.steps, step)
	return p
}

// Execute runs the pipeline on the input.
func (p *Pipeline[T]) Execute(ctx context.Context, input T) (T, error) {
	current := input
	for _, step := range p.steps {
		if ctx.Err() != nil {
			return current, ctx.Err()
		}
		result, err := step(ctx, current)
		if err != nil {
			return current, err
		}
		current = result
	}
	return current, nil
}

// Collector accumulates results from streaming operations.
type Collector[T any] struct {
	mu      sync.Mutex
	items   []T
	maxSize int
}

// NewCollector creates a collector with optional max size.
func NewCollector[T any](maxSize int) *Collector[T] {
	return &Collector[T]{
		items:   make([]T, 0),
		maxSize: maxSize,
	}
}

// Add adds an item to the collector (thread-safe).
func (c *Collector[T]) Add(item T) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.maxSize > 0 && len(c.items) >= c.maxSize {
		return false
	}
	c.items = append(c.items, item)
	return true
}

// Items returns all collected items.
func (c *Collector[T]) Items() []T {
	c.mu.Lock()
	defer c.mu.Unlock()

	result := make([]T, len(c.items))
	copy(result, c.items)
	return result
}

// Count returns the number of collected items.
func (c *Collector[T]) Count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.items)
}

// Clear removes all items from the collector.
func (c *Collector[T]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = c.items[:0]
}
