package worker

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zerodayz7/http-server/internal/service"
	"github.com/zerodayz7/http-server/internal/shared/logger"
)

type MockInteractionRepo struct {
	mock.Mock
}

func (m *MockInteractionRepo) IncrementBy(ctx context.Context, typ string, count int64) error {
	args := m.Called(ctx, typ, count)
	return args.Error(0)
}

func (m *MockInteractionRepo) Increment(ctx context.Context, typ string) error {
	args := m.Called(ctx, typ)
	return args.Error(0)
}

func (m *MockInteractionRepo) GetStats(ctx context.Context) (service.InteractionStatsDTO, error) {
	args := m.Called(ctx)
	return args.Get(0).(service.InteractionStatsDTO), args.Error(1)
}

func TestInteractionWorker_FullFlow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	rdb := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "devpassword",
		DB:       0,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis unavailable: %v", err)
	}

	// Czyścimy środowisko
	rdb.Del(ctx, streamName)
	rdb.Del(ctx, dlqStream)

	mockRepo := new(MockInteractionRepo)
	log := logger.NewNop()

	// FIX: Dodano 5-ty argument (flushInterval).
	// 500ms shutdown timeout, 200ms flush interval dla szybkich testów.
	w := NewInteractionWorker(rdb, mockRepo, log, 500*time.Millisecond, 200*time.Millisecond)

	mockRepo.On("IncrementBy",
		mock.Anything,
		mock.MatchedBy(func(s string) bool { return s == "like" || s == "view" }),
		mock.AnythingOfType("int64"),
	).Return(nil)

	workerCtx, stopWorker := context.WithCancel(ctx)
	errChan := make(chan error, 1)

	go func() {
		errChan <- w.Start(workerCtx)
	}()

	// Czekamy na gotowość
	time.Sleep(200 * time.Millisecond)

	// Produkcja eventów
	events := []string{"like", "like", "view"}
	for _, e := range events {
		rdb.XAdd(ctx, &redis.XAddArgs{
			Stream: streamName,
			Values: map[string]interface{}{"type": e},
		})
	}

	// Czekamy na cykl flush (ustawiliśmy go na 200ms, więc 500ms wystarczy z zapasem)
	time.Sleep(500 * time.Millisecond)
	stopWorker()

	select {
	case err := <-errChan:
		if err != nil && !strings.Contains(err.Error(), "context canceled") {
			assert.NoError(t, err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Worker did not stop in time")
	}

	mockRepo.AssertExpectations(t)

	assert.Eventually(t, func() bool {
		res, err := rdb.XPending(ctx, streamName, groupName).Result()
		return err == nil && res.Count == 0
	}, 2*time.Second, 100*time.Millisecond, "Wiadomości powinny zostać ACKnięte")
}

func TestInteractionWorker_DLQ_Flow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rdb := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "devpassword",
		DB:       0,
	})

	mockRepo := new(MockInteractionRepo)

	// FIX: Dodano 5-ty argument (flushInterval)
	w := NewInteractionWorker(rdb, mockRepo, logger.NewNop(), 500*time.Millisecond, 200*time.Millisecond)

	rdb.Del(ctx, streamName)
	rdb.Del(ctx, dlqStream) // Czyścimy też DLQ przed testem

	rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: streamName,
		Values: map[string]interface{}{"garbage": "data"}, // błędne dane - trafią do DLQ
	})

	workerCtx, stop := context.WithCancel(ctx)
	go func() { _ = w.Start(workerCtx) }()

	// Poczekaj chwilę na przetworzenie
	time.Sleep(500 * time.Millisecond)
	stop()

	assert.Eventually(t, func() bool {
		res, err := rdb.XPending(ctx, streamName, groupName).Result()
		return err == nil && res.Count == 0
	}, 2*time.Second, 100*time.Millisecond)
}
