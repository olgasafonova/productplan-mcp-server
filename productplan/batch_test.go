package productplan

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestExecute_Sequential(t *testing.T) {
	config := BatchConfig{Concurrency: 1, StopOnError: false}

	fns := []func(ctx context.Context) (int, error){
		func(ctx context.Context) (int, error) { return 1, nil },
		func(ctx context.Context) (int, error) { return 2, nil },
		func(ctx context.Context) (int, error) { return 3, nil },
	}

	result := Execute(context.Background(), config, fns)

	if result.ErrorCount() != 0 {
		t.Errorf("Expected 0 errors, got %d", result.ErrorCount())
	}
	if result.SuccessCount() != 3 {
		t.Errorf("Expected 3 results, got %d", result.SuccessCount())
	}
}

func TestExecute_Concurrent(t *testing.T) {
	config := BatchConfig{Concurrency: 3, StopOnError: false}

	var counter int64
	fns := make([]func(ctx context.Context) (int, error), 10)
	for i := 0; i < 10; i++ {
		idx := i
		fns[i] = func(ctx context.Context) (int, error) {
			atomic.AddInt64(&counter, 1)
			time.Sleep(10 * time.Millisecond)
			return idx, nil
		}
	}

	result := Execute(context.Background(), config, fns)

	if result.ErrorCount() != 0 {
		t.Errorf("Expected 0 errors, got %d", result.ErrorCount())
	}
	if result.SuccessCount() != 10 {
		t.Errorf("Expected 10 results, got %d", result.SuccessCount())
	}
	if counter != 10 {
		t.Errorf("Expected 10 executions, got %d", counter)
	}
}

func TestExecute_WithErrors(t *testing.T) {
	config := BatchConfig{Concurrency: 1, StopOnError: false}

	fns := []func(ctx context.Context) (int, error){
		func(ctx context.Context) (int, error) { return 1, nil },
		func(ctx context.Context) (int, error) { return 0, errors.New("test error") },
		func(ctx context.Context) (int, error) { return 3, nil },
	}

	result := Execute(context.Background(), config, fns)

	if result.ErrorCount() != 1 {
		t.Errorf("Expected 1 error, got %d", result.ErrorCount())
	}
	if result.SuccessCount() != 2 {
		t.Errorf("Expected 2 results, got %d", result.SuccessCount())
	}
	if !result.HasErrors() {
		t.Error("Expected HasErrors() to return true")
	}
}

func TestExecute_StopOnError(t *testing.T) {
	config := BatchConfig{Concurrency: 1, StopOnError: true}

	callCount := 0
	fns := []func(ctx context.Context) (int, error){
		func(ctx context.Context) (int, error) { callCount++; return 1, nil },
		func(ctx context.Context) (int, error) { callCount++; return 0, errors.New("stop here") },
		func(ctx context.Context) (int, error) { callCount++; return 3, nil },
	}

	result := Execute(context.Background(), config, fns)

	if callCount != 2 {
		t.Errorf("Expected 2 calls (stop on error), got %d", callCount)
	}
	if result.SuccessCount() != 1 {
		t.Errorf("Expected 1 result, got %d", result.SuccessCount())
	}
}

func TestExecute_ContextCancellation(t *testing.T) {
	config := BatchConfig{Concurrency: 1, StopOnError: false}

	ctx, cancel := context.WithCancel(context.Background())

	callCount := 0
	fns := []func(ctx context.Context) (int, error){
		func(ctx context.Context) (int, error) {
			callCount++
			cancel() // Cancel after first call
			return 1, nil
		},
		func(ctx context.Context) (int, error) { callCount++; return 2, nil },
		func(ctx context.Context) (int, error) { callCount++; return 3, nil },
	}

	result := Execute(ctx, config, fns)

	if callCount != 1 {
		t.Errorf("Expected 1 call before context cancellation, got %d", callCount)
	}
	if result.SuccessCount() != 1 {
		t.Errorf("Expected 1 result, got %d", result.SuccessCount())
	}
	if !errors.Is(result.Errors[0].Err, context.Canceled) {
		t.Error("Expected context.Canceled error")
	}
}

func TestExecuteWithKeys(t *testing.T) {
	config := BatchConfig{Concurrency: 2, StopOnError: false}

	keys := []string{"roadmap-1", "roadmap-2", "roadmap-3"}

	result := ExecuteWithKeys(context.Background(), config, keys, func(ctx context.Context, key string) (string, error) {
		if key == "roadmap-2" {
			return "", errors.New("not found")
		}
		return "bars for " + key, nil
	})

	if result.SuccessCount() != 2 {
		t.Errorf("Expected 2 results, got %d", result.SuccessCount())
	}
	if result.ErrorCount() != 1 {
		t.Errorf("Expected 1 error, got %d", result.ErrorCount())
	}
	if result.Errors[0].Key != "roadmap-2" {
		t.Errorf("Expected error key 'roadmap-2', got '%s'", result.Errors[0].Key)
	}
}

func TestExecute_EmptyInput(t *testing.T) {
	config := DefaultBatchConfig()

	result := Execute(context.Background(), config, []func(ctx context.Context) (int, error){})

	if result.SuccessCount() != 0 {
		t.Errorf("Expected 0 results, got %d", result.SuccessCount())
	}
	if result.ErrorCount() != 0 {
		t.Errorf("Expected 0 errors, got %d", result.ErrorCount())
	}
}

func TestPaginator_FetchAll(t *testing.T) {
	paginator := &Paginator[int]{
		PageSize: 3,
		MaxPages: 0,
	}

	// Simulate 7 total items across 3 pages
	allItems := []int{1, 2, 3, 4, 5, 6, 7}

	result := paginator.FetchAll(context.Background(), func(ctx context.Context, page, pageSize int) ([]int, bool, error) {
		start := (page - 1) * pageSize
		if start >= len(allItems) {
			return nil, false, nil
		}
		end := start + pageSize
		if end > len(allItems) {
			end = len(allItems)
		}
		hasMore := end < len(allItems)
		return allItems[start:end], hasMore, nil
	})

	if result.Error != nil {
		t.Errorf("Unexpected error: %v", result.Error)
	}
	if len(result.Items) != 7 {
		t.Errorf("Expected 7 items, got %d", len(result.Items))
	}
	if result.TotalPages != 3 {
		t.Errorf("Expected 3 pages, got %d", result.TotalPages)
	}
}

func TestPaginator_MaxPages(t *testing.T) {
	paginator := &Paginator[int]{
		PageSize: 2,
		MaxPages: 2,
	}

	result := paginator.FetchAll(context.Background(), func(ctx context.Context, page, pageSize int) ([]int, bool, error) {
		return []int{page * 10, page*10 + 1}, true, nil // Always has more
	})

	if result.Error != nil {
		t.Errorf("Unexpected error: %v", result.Error)
	}
	if len(result.Items) != 4 {
		t.Errorf("Expected 4 items (2 pages * 2), got %d", len(result.Items))
	}
	if result.TotalPages != 2 {
		t.Errorf("Expected 2 pages (max), got %d", result.TotalPages)
	}
}

func TestPaginator_Error(t *testing.T) {
	paginator := &Paginator[int]{
		PageSize: 3, // Match the returned item count
		MaxPages: 0,
	}

	result := paginator.FetchAll(context.Background(), func(ctx context.Context, page, pageSize int) ([]int, bool, error) {
		if page == 2 {
			return nil, false, errors.New("api error")
		}
		return []int{1, 2, 3}, true, nil
	})

	if result.Error == nil {
		t.Error("Expected error")
	}
	// TotalPages is the last successful page (1), since error occurred on page 2
	if result.TotalPages != 1 {
		t.Errorf("Expected last successful page 1, got %d", result.TotalPages)
	}
	// Should have items from page 1
	if len(result.Items) != 3 {
		t.Errorf("Expected 3 items from page 1, got %d", len(result.Items))
	}
}

func TestPipeline(t *testing.T) {
	pipeline := NewPipeline[int]().
		AddStep(func(ctx context.Context, n int) (int, error) {
			return n * 2, nil
		}).
		AddStep(func(ctx context.Context, n int) (int, error) {
			return n + 10, nil
		}).
		AddStep(func(ctx context.Context, n int) (int, error) {
			return n / 2, nil
		})

	result, err := pipeline.Execute(context.Background(), 5)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// (5 * 2 + 10) / 2 = 10
	if result != 10 {
		t.Errorf("Expected 10, got %d", result)
	}
}

func TestPipeline_Error(t *testing.T) {
	pipeline := NewPipeline[int]().
		AddStep(func(ctx context.Context, n int) (int, error) {
			return n * 2, nil
		}).
		AddStep(func(ctx context.Context, n int) (int, error) {
			return 0, errors.New("step failed")
		}).
		AddStep(func(ctx context.Context, n int) (int, error) {
			return n + 100, nil // Should not be called
		})

	_, err := pipeline.Execute(context.Background(), 5)
	if err == nil {
		t.Error("Expected error")
	}
	if err.Error() != "step failed" {
		t.Errorf("Expected 'step failed', got '%s'", err.Error())
	}
}

func TestCollector(t *testing.T) {
	collector := NewCollector[string](0) // Unlimited

	collector.Add("item1")
	collector.Add("item2")
	collector.Add("item3")

	if collector.Count() != 3 {
		t.Errorf("Expected 3 items, got %d", collector.Count())
	}

	items := collector.Items()
	if len(items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(items))
	}

	collector.Clear()
	if collector.Count() != 0 {
		t.Errorf("Expected 0 items after clear, got %d", collector.Count())
	}
}

func TestCollector_MaxSize(t *testing.T) {
	collector := NewCollector[int](3)

	for i := 1; i <= 5; i++ {
		collector.Add(i)
	}

	if collector.Count() != 3 {
		t.Errorf("Expected 3 items (max), got %d", collector.Count())
	}

	items := collector.Items()
	// Should only have first 3 items
	expected := []int{1, 2, 3}
	for i, v := range items {
		if v != expected[i] {
			t.Errorf("Item %d: expected %d, got %d", i, expected[i], v)
		}
	}
}

func TestCollector_Concurrent(t *testing.T) {
	collector := NewCollector[int](0)

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			for j := 0; j < 100; j++ {
				collector.Add(n*100 + j)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	if collector.Count() != 1000 {
		t.Errorf("Expected 1000 items, got %d", collector.Count())
	}
}

func TestDefaultBatchConfig(t *testing.T) {
	config := DefaultBatchConfig()

	if config.Concurrency != 3 {
		t.Errorf("Expected concurrency 3, got %d", config.Concurrency)
	}
	if config.StopOnError != false {
		t.Error("Expected StopOnError to be false")
	}
}
