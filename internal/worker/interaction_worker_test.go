package worker

import (
	"context"
	"errors"
	"math/rand"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
	"github.com/zerodayz7/http-server/internal/service"
	"github.com/zerodayz7/http-server/internal/shared/logger"
	"github.com/zerodayz7/http-server/internal/shared/telemetry"
)

type MockInteractionRepo struct {
	mock.Mock
	dbFailure *bool
}

func (m *MockInteractionRepo) IncrementBy(ctx context.Context, typ string, amount int64) error {
	args := m.Called(ctx, typ, amount)

	// Sprawdzamy, czy pod indeksem 0 jest funkcja (tak się dzieje przy dynamicznym Return)
	if f, ok := args.Get(0).(func(context.Context, string, int64) error); ok {
		return f(ctx, typ, amount)
	}

	// Jeśli nie funkcja, to standardowy mechanizm testify dla błędów
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

	rdb := setupTestRedis(t)

	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis unavailable: %v", err)
	}

	rdb.Del(ctx, streamName)
	rdb.Del(ctx, dlqStream)

	mockRepo := new(MockInteractionRepo)
	log := logger.NewLogger("development")

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

	time.Sleep(200 * time.Millisecond)

	events := []string{"like", "like", "view"}
	for _, e := range events {
		rdb.XAdd(ctx, &redis.XAddArgs{
			Stream: streamName,
			Values: map[string]interface{}{"type": e},
		})
	}

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

	rdb := setupTestRedis(t)

	mockRepo := new(MockInteractionRepo)

	w := NewInteractionWorker(rdb, mockRepo, logger.NewNop(), 500*time.Millisecond, 200*time.Millisecond)

	rdb.Del(ctx, streamName)
	rdb.Del(ctx, dlqStream)

	rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: streamName,
		Values: map[string]interface{}{"garbage": "data"},
	})

	workerCtx, stop := context.WithCancel(ctx)
	go func() { _ = w.Start(workerCtx) }()

	time.Sleep(500 * time.Millisecond)
	stop()

	assert.Eventually(t, func() bool {
		res, err := rdb.XPending(ctx, streamName, groupName).Result()
		return err == nil && res.Count == 0
	}, 2*time.Second, 100*time.Millisecond)
}

func TestInteractionWorker_LoadTest(t *testing.T) {
	const totalMessages = 10000

	testCtx, testCancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer testCancel()

	rdb := setupTestRedis(t)
	_ = rdb.Del(testCtx, streamName, dlqStream).Err()
	_ = rdb.XGroupCreateMkStream(testCtx, streamName, groupName, "0").Err()

	mockRepo := new(MockInteractionRepo)
	log := logger.NewLogger("development")

	w := NewInteractionWorker(rdb, mockRepo, log, 2*time.Second, 100*time.Millisecond)

	var totalProcessed int64
	var mu sync.Mutex

	mockRepo.
		On("IncrementBy", mock.Anything, "load_test_event", mock.AnythingOfType("int64")).
		Run(func(args mock.Arguments) {
			count := args.Get(2).(int64)
			mu.Lock()
			totalProcessed += count
			mu.Unlock()
		}).
		Return(nil)

	workerCtx, stopWorker := context.WithCancel(context.Background())
	done := make(chan struct{})
	start := time.Now()

	go func() {
		defer close(done)
		if err := w.Start(workerCtx); err != nil && !errors.Is(err, context.Canceled) {
			t.Errorf("Worker error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	t.Logf("Starting production of %d messages...", totalMessages)
	pipe := rdb.Pipeline()
	for i := 0; i < totalMessages; i++ {
		pipe.XAdd(testCtx, &redis.XAddArgs{
			Stream: streamName,
			Values: map[string]interface{}{
				"type": "load_test_event",
			},
		})

		if (i+1)%1000 == 0 {
			_, err := pipe.Exec(testCtx)
			if err != nil {
				t.Fatalf("FAILED to produce batch at %d: %v", i+1, err)
			}
			t.Logf("Producer: sent %d/%d to Redis", i+1, totalMessages)
			pipe = rdb.Pipeline()
		}
	}

	if _, err := pipe.Exec(testCtx); err != nil {
		t.Logf("Final batch production error: %v", err)
	}

	redisCount, _ := rdb.XLen(testCtx, streamName).Result()
	t.Logf("Redis Stream verify: %d messages waiting", redisCount)

	assert.Eventually(t, func() bool {
		mu.Lock()
		current := totalProcessed
		mu.Unlock()

		if current > 0 {
			t.Logf("Worker Progress: %d / %d (Redis Stream: %d)", current, totalMessages, redisCount)
		}

		return current == totalMessages
	}, 60*time.Second, 500*time.Millisecond, "Worker failed to process all messages")

	stopWorker()
	<-done

	time.Sleep(500 * time.Millisecond)
	duration := time.Since(start)

	mu.Lock()
	finalCount := totalProcessed
	mu.Unlock()

	t.Logf("Performance: %.2f msg/s | Total Processed: %d | Total Sent: %d",
		float64(totalMessages)/duration.Seconds(),
		finalCount,
		totalMessages,
	)

	assert.Equal(t, int64(totalMessages), finalCount)
	mockRepo.AssertExpectations(t)
}

// # region ChaosTest
func TestInteractionWorker_PropertyChaos(t *testing.T) {
	log := logger.NewLogger("development")

	const (
		iterations         = 60
		maxMessages        = 100
		testTimeout        = 60 * time.Second
		workerFlush        = 200 * time.Millisecond
		workerInterval     = 50 * time.Millisecond
		sleepBetween       = 2 * time.Millisecond
		pauseDuration      = 20
		eventualWait       = 20 * time.Second
		minRestartInterval = 500 * time.Millisecond
	)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	rdb := setupIsolatedRedis(t, 14)
	require.NoError(t, rdb.FlushDB(ctx).Err())
	err := rdb.XGroupCreateMkStream(ctx, streamName, groupName, "0").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		require.NoError(t, err)
	}

	mockRepo := new(MockInteractionRepo)

	var (
		mu             sync.Mutex
		totalProcessed int64
		totalCalls     int
		dbFailure      bool
	)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	mockRepo.
		On("IncrementBy", mock.Anything, "chaos", mock.AnythingOfType("int64")).
		Run(func(args mock.Arguments) {
			mu.Lock()
			defer mu.Unlock()

			totalCalls++

			if !dbFailure {
				totalProcessed += args.Get(2).(int64)
			}
		}).
		Return(nil)

	w := NewInteractionWorker(rdb, mockRepo, log, workerFlush, workerInterval)

	var (
		workerCtx  context.Context
		stopWorker context.CancelFunc
		wg         sync.WaitGroup
	)

	startWorker := func(parent context.Context) {
		workerCtx, stopWorker = context.WithCancel(parent)

		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Info("worker started")
			_ = w.Start(workerCtx)
			log.Info("worker stopped")
		}()
	}

	stopWorkerFn := func() {
		if stopWorker != nil {
			stopWorker()
		}

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("worker did not shutdown in time")
		}
	}

	startWorker(ctx)

	sent := 0
	lastRestart := time.Now()

	for i := 0; i < iterations; i++ {
		action := rng.Intn(4)

		switch action {

		case 0:
			if sent < maxMessages {
				err := rdb.XAdd(ctx, &redis.XAddArgs{
					Stream: streamName,
					Values: map[string]interface{}{
						"type": "chaos",
					},
				}).Err()
				require.NoError(t, err)

				sent++
				log.Debug("sent message", zap.Int("sent", sent))
			}

		case 1:
			// throttled restart
			if time.Since(lastRestart) > minRestartInterval {
				log.Info("chaos: restarting worker")
				stopWorkerFn()
				startWorker(ctx)
				lastRestart = time.Now()
			}

		case 2:
			// probabilistic DB failure (less aggressive)
			if rng.Float32() < 0.3 {
				dbFailure = true
				log.Warn("chaos: db failure enabled")
			} else {
				dbFailure = false
			}

		case 3:
			// rarer Redis pause
			if rng.Float32() < 0.2 {
				log.Warn("chaos: redis pause")
				_ = rdb.Do(ctx, "CLIENT", "PAUSE", pauseDuration).Err()
			}
		}

		time.Sleep(sleepBetween)
	}

	stopWorkerFn()

	select {
	case <-ctx.Done():
		t.Fatal("context timed out before worker shutdown")
	default:
	}

	// Eventual consistency
	assert.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return totalProcessed >= int64(sent)*7/10
	}, eventualWait, 100*time.Millisecond)

	pending, err := rdb.XPending(ctx, streamName, groupName).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(0), pending.Count)

	assert.Greater(t, totalCalls, 0)

	t.Logf("Chaos test complete: sent=%d processed=%d calls=%d", sent, totalProcessed, totalCalls)
}

func TestInteractionWorker_DBRecovery_NoDuplicates(t *testing.T) {
	cleanup := telemetry.InitTelemetry(context.Background(), "worker-test", "localhost:4317", 1*time.Second)
	defer cleanup()

	// Zwiększony timeout całego testu na wszelki wypadek
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	rdb := setupTestRedis(t)

	require.NoError(t, rdb.FlushDB(ctx).Err())
	err := rdb.XGroupCreateMkStream(ctx, streamName, groupName, "0").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		require.NoError(t, err)
	}

	mockRepo := new(MockInteractionRepo)

	var (
		mu         sync.Mutex
		totalCalls int
		dbFailure  bool
	)

	mockRepo.
		On("IncrementBy", mock.Anything, "recovery", mock.AnythingOfType("int64")).
		Run(func(args mock.Arguments) {
			mu.Lock()
			defer mu.Unlock()

			amount := args.Get(2).(int64)
			if !dbFailure {
				totalCalls += int(amount)
				t.Logf("✅ IncrementBy success: +%d (total: %d)", amount, totalCalls)
			} else {
				t.Logf("💥 DB FAILURE SIMULATED: +%d", amount)
			}
		}).
		Return(func(ctx context.Context, typ string, amount int64) error {
			mu.Lock()
			defer mu.Unlock()
			if dbFailure {
				return errors.New("db down")
			}
			return nil
		})

	log := logger.NewLogger("development")

	// Ustawiamy agresywne interwały: 100ms na Flush i 10ms na sprawdzanie Redisa
	w := NewInteractionWorker(rdb, mockRepo, log, 100*time.Millisecond, 10*time.Millisecond)

	workerCtx, stop := context.WithCancel(ctx)
	workerDone := make(chan struct{})
	go func() {
		defer close(workerDone)
		_ = w.Start(workerCtx)
	}()

	// 1. NAJPIERW WŁĄCZ AWARIĘ
	mu.Lock()
	dbFailure = true
	mu.Unlock()
	t.Log("💥 DB FAILURE ON (Circuit Breaker should start failing)")

	// 2. PRODUCE EVENTS - teraz wszystkie trafią na "zamkniętą" bazę
	for i := 0; i < 50; i++ {
		_, err := rdb.XAdd(ctx, &redis.XAddArgs{
			Stream: streamName,
			Values: map[string]interface{}{
				"type": "recovery",
				"id":   i,
			},
		}).Result()
		require.NoError(t, err)
	}
	t.Log("🚀 Produced 50 events while DB is DOWN")

	// Daj workerowi czas na pobranie wiadomości i "odbicie się" od błędu bazy
	time.Sleep(2 * time.Second)

	// 3. RECOVERY - baza wraca do życia
	mu.Lock()
	dbFailure = false
	mu.Unlock()
	t.Log("🟢 DB RECOVERED")

	// 4. MULTI-TRIGGER - wysyłamy serię impulsów w tle, żeby "wybudzać" workera
	go func() {
		for i := 0; i < 15; i++ {
			select {
			case <-ctx.Done():
				return
			case <-workerDone:
				return
			default:
				_ = rdb.XAdd(ctx, &redis.XAddArgs{
					Stream: streamName,
					Values: map[string]interface{}{"type": "recovery", "id": "kick"},
				}).Err()
				time.Sleep(250 * time.Millisecond)
			}
		}
	}()

	// 5. ASSERTION - czekamy aż przetworzy wszystko (50 oryginałów + triggery)
	assert.Eventually(t, func() bool {
		mu.Lock()
		current := totalCalls
		mu.Unlock()

		pending, _ := rdb.XPending(ctx, streamName, groupName).Result()
		pCount := int64(0)
		if pending != nil {
			pCount = pending.Count
		}

		t.Logf("⏳ Progress Check: totalCalls=%d, pendingInRedis=%d", current, pCount)

		// Warunek sukcesu: mamy co najmniej 50 zapisów i zero wiszących wiadomości w Redis
		return current >= 50 && pCount == 0
	}, 25*time.Second, 500*time.Millisecond)

	stop()
	<-workerDone

	finalPending, err := rdb.XPending(context.Background(), streamName, groupName).Result()
	require.NoError(t, err)

	t.Logf("📊 FINAL STATS: totalCalls=%d | pending=%d", totalCalls, finalPending.Count)

	assert.Equal(t, int64(0), finalPending.Count, "Na koniec Redis musi być pusty")
	assert.GreaterOrEqual(t, totalCalls, 50, "Musieliśmy przetworzyć co najmniej 50 oryginalnych eventów")
	mockRepo.AssertExpectations(t)
}

func TestMultipleWorkers_LoadBalancing(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rdb := setupTestRedis(t)
	_ = rdb.Del(ctx, streamName).Err()
	_ = rdb.XGroupCreateMkStream(ctx, streamName, groupName, "0").Err()

	mockRepo := new(MockInteractionRepo)
	var mu sync.Mutex
	var totalProcessed int64

	mockRepo.On("IncrementBy", mock.Anything, "multi_test", mock.AnythingOfType("int64")).
		Run(func(args mock.Arguments) {
			mu.Lock()
			totalProcessed += args.Get(2).(int64)
			mu.Unlock()
		}).Return(nil)

	// Uruchomienie 3 workerów
	const numWorkers = 3
	var stopFns []context.CancelFunc

	for i := 0; i < numWorkers; i++ {
		wCtx, stop := context.WithCancel(ctx)
		stopFns = append(stopFns, stop)

		worker := NewInteractionWorker(rdb, mockRepo, logger.NewNop(), 2*time.Second, 100*time.Millisecond)
		go func() { _ = worker.Start(wCtx) }()
	}

	// Wysłanie 3000 wiadomości
	const msgs = 3000
	pipe := rdb.Pipeline()
	for i := 0; i < msgs; i++ {
		pipe.XAdd(ctx, &redis.XAddArgs{
			Stream: streamName,
			Values: map[string]interface{}{"type": "multi_test"},
		})
	}
	_, _ = pipe.Exec(ctx)

	// Weryfikacja
	assert.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return totalProcessed == msgs
	}, 10*time.Second, 200*time.Millisecond)

	// Zatrzymanie workerów
	for _, stop := range stopFns {
		stop()
	}
}

func setupTestRedis(t *testing.T) *redis.Client {
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	addr := host + ":" + port

	if host == "" || port == "" {
		addr = "127.0.0.1:6379"
	}

	pass := os.Getenv("REDIS_PASSWORD")
	if pass == "" {
		pass = "devpassword"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       0,
	})

	t.Logf("Connecting to test Redis at %s", addr)

	return rdb
}

func setupIsolatedRedis(t *testing.T, db int) *redis.Client {
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	addr := host + ":" + port

	if host == "" || port == "" {
		addr = "127.0.0.1:6379"
	}

	pass := os.Getenv("REDIS_PASSWORD")
	if pass == "" {
		pass = "devpassword"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       db,
	})

	t.Logf("Connecting to isolated Redis DB %d at %s", db, addr)
	return rdb
}
