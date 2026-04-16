package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zerodayz7/http-server/internal/service"
	"github.com/zerodayz7/http-server/internal/service/mocks"
)

func TestProcessInitialVisit_ShouldNotPublishIfAlreadyRecorded(t *testing.T) {
	// 1. GEST (Setup) - Tworzymy mocki dla zależności
	mockRepo := new(mocks.InteractionRepository)
	mockCache := new(mocks.InteractionCache)
	mockPublisher := new(mocks.EventPublisher)
	mockIdentity := new(mocks.IdentityService)
	// Loggera możemy zostawić jako nil lub dodać prosty mock,
	// jeśli serwis go nie używa intensywnie do logiki.

	s := service.NewInteractionService(mockRepo, mockCache, mockPublisher, mockIdentity, nil)

	ctx := context.Background()
	fp := "fake-fingerprint"

	// 2. PROGRAMOWANIE (Expectations)
	// Mówimy: TryRecordVisit ma zwrócić FALSE (użytkownik już tu był)
	mockCache.On("TryRecordVisit", ctx, fp, service.Cooldown).Return(false, nil)

	// Mówimy: Pobieranie statystyk ma zwrócić jakieś dane
	mockCache.On("GetGlobalCount", ctx, mock.Anything).Return(int64(100), true)
	mockCache.On("GetUserChoice", ctx, fp).Return("", false, nil)

	// 3. DZIAŁANIE (Execution)
	res, err := s.ProcessInitialVisit(ctx, fp)

	// 4. SPRAWDZENIE (Assertions)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, int64(100), res.Visits)

	// KLUCZOWE: Sprawdzamy, czy PublishInteraction NIGDY nie został wywołany
	mockPublisher.AssertNotCalled(t, "PublishInteraction", mock.Anything, mock.Anything, mock.Anything)

	// Sprawdzamy czy wszystkie zaprogramowane metody zostały zawołane
	mockCache.AssertExpectations(t)
}
