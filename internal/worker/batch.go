package worker

import (
	"sync"

	"github.com/zerodayz7/http-server/internal/shared/logger"
)

type eventBatch struct {
	mu     sync.Mutex
	counts map[string]int64
	ids    []string
	logger logger.Logger
}

func newEventBatch(log logger.Logger) *eventBatch {
	return &eventBatch{
		counts: make(map[string]int64),
		ids:    make([]string, 0, batchSizeLimit),
		logger: log,
	}
}

func (b *eventBatch) add(id string, eventType string) {
	b.mu.Lock()
	b.counts[eventType]++
	b.ids = append(b.ids, id)
	b.mu.Unlock()
}

func (b *eventBatch) snapshot() (map[string]int64, []string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.ids) == 0 {
		return nil, nil
	}

	countsCopy := make(map[string]int64, len(b.counts))
	for k, v := range b.counts {
		countsCopy[k] = v
	}

	idsCopy := make([]string, len(b.ids))
	copy(idsCopy, b.ids)

	return countsCopy, idsCopy
}

func (b *eventBatch) clear() {
	b.mu.Lock()
	b.counts = make(map[string]int64)
	b.ids = b.ids[:0]
	b.mu.Unlock()
}

func (b *eventBatch) size() int {
	b.mu.Lock()
	n := len(b.ids)
	b.mu.Unlock()
	return n
}
