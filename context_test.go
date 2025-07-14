package sqlcomment

import (
	"context"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestWithTagThreadSafety(t *testing.T) {
	ctx := context.Background()
	var wg sync.WaitGroup
	const numGoroutines = 100
	results := make(map[int]Tags)
	mu := sync.Mutex{}
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// Add a unique tag for each goroutine
			newCtx := WithTag(ctx, "goroutine", string(rune('A'+i)))
			tags := FromContext(newCtx)
			// Store the result in the map
			mu.Lock()
			results[i] = tags
			mu.Unlock()
		}(i)
	}

	wg.Wait()
	// Verify that each goroutine has its own independent tags
	assert.Equal(t, numGoroutines, len(results))
	for i := 0; i < numGoroutines; i++ {
		assert.Contains(t, results[i], "goroutine")
		assert.Equal(t, string(rune('A'+i)), results[i]["goroutine"])
	}
}
