package application

import (
	"context"
	"testing"

	"github.com/Carlvalencia1/streamhub-backend/internal/notifications/domain"
)

// MockNotificationRepository implementa domain.NotificationRepository para tests
type MockNotificationRepository struct {
	tokens map[string][]*domain.DeviceToken
	err    error
}

func NewMockNotificationRepository() *MockNotificationRepository {
	return &MockNotificationRepository{
		tokens: make(map[string][]*domain.DeviceToken),
	}
}

func (m *MockNotificationRepository) SaveDeviceToken(ctx context.Context, token *domain.DeviceToken) error {
	return m.err
}

func (m *MockNotificationRepository) RemoveDeviceToken(ctx context.Context, userID, token string) error {
	return m.err
}

func (m *MockNotificationRepository) GetDeviceTokensByUser(ctx context.Context, userID string) ([]*domain.DeviceToken, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.tokens[userID], nil
}

func (m *MockNotificationRepository) GetDeviceTokensByUsersExcept(ctx context.Context, excludeUserID string) ([]*domain.DeviceToken, error) {
	if m.err != nil {
		return nil, m.err
	}

	var result []*domain.DeviceToken
	for uid, tokens := range m.tokens {
		if uid != excludeUserID {
			result = append(result, tokens...)
		}
	}
	return result, nil
}

func (m *MockNotificationRepository) MarkTokenAsInvalid(ctx context.Context, token string) error {
	return m.err
}

func (m *MockNotificationRepository) RemoveInvalidTokens(ctx context.Context) error {
	return m.err
}

func (m *MockNotificationRepository) UpdateTokenLastUsed(ctx context.Context, token string) error {
	return m.err
}

// MockPushProvider implementa domain.PushProvider para tests
type MockPushProvider struct {
	lastTokens  []string
	lastPayload *domain.PushPayload
	err         error
	calledCount int
}

func NewMockPushProvider() *MockPushProvider {
	return &MockPushProvider{}
}

func (m *MockPushProvider) SendMulticast(ctx context.Context, tokens []string, payload *domain.PushPayload) error {
	m.calledCount++
	m.lastTokens = tokens
	m.lastPayload = payload
	return m.err
}

func (m *MockPushProvider) IsTokenInvalid(err error) bool {
	return false
}

// Tests
func TestNotifyStreamLive_Execute_Success(t *testing.T) {
	mockRepo := NewMockNotificationRepository()
	mockProvider := NewMockPushProvider()

	// Setup: agregar tokens para usuarios no-owner
	user2Token := domain.NewDeviceToken("id1", "user2", "token_user2_1", "android", "device1", "1.0")
	user3Token := domain.NewDeviceToken("id2", "user3", "token_user3_1", "android", "device2", "1.0")

	mockRepo.tokens["user2"] = []*domain.DeviceToken{user2Token}
	mockRepo.tokens["user3"] = []*domain.DeviceToken{user3Token}

	usecase := NewNotifyStreamLive(mockRepo, mockProvider)

	// Execute
	input := NotifyStreamLiveInput{
		StreamID:    "stream_123",
		StreamTitle: "My Awesome Stream",
		OwnerUserID: "user1",
	}

	err := usecase.Execute(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if mockProvider.calledCount != 1 {
		t.Fatalf("expected SendMulticast to be called once, got %d", mockProvider.calledCount)
	}

	if len(mockProvider.lastTokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(mockProvider.lastTokens))
	}

	// Verificar payload
	if mockProvider.lastPayload.Data["stream_id"] != "stream_123" {
		t.Fatalf("expected stream_id to be stream_123, got %s", mockProvider.lastPayload.Data["stream_id"])
	}

	if mockProvider.lastPayload.Data["type"] != "stream_live" {
		t.Fatalf("expected type to be stream_live, got %s", mockProvider.lastPayload.Data["type"])
	}
}

func TestNotifyStreamLive_Execute_NoTokens(t *testing.T) {
	mockRepo := NewMockNotificationRepository()
	mockProvider := NewMockPushProvider()

	// No tokens registered
	usecase := NewNotifyStreamLive(mockRepo, mockProvider)

	input := NotifyStreamLiveInput{
		StreamID:    "stream_123",
		StreamTitle: "My Awesome Stream",
		OwnerUserID: "user1",
	}

	err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if mockProvider.calledCount != 0 {
		t.Fatalf("expected SendMulticast NOT to be called, got %d calls", mockProvider.calledCount)
	}
}

func TestNotifyStreamLive_Execute_InvalidInput(t *testing.T) {
	mockRepo := NewMockNotificationRepository()
	mockProvider := NewMockPushProvider()
	usecase := NewNotifyStreamLive(mockRepo, mockProvider)

	// Empty streamID
	input := NotifyStreamLiveInput{
		StreamID:    "",
		StreamTitle: "Title",
		OwnerUserID: "user1",
	}

	err := usecase.Execute(context.Background(), input)

	if err != ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestNotifyStreamLive_Execute_RepositoryError(t *testing.T) {
	mockRepo := NewMockNotificationRepository()
	mockProvider := NewMockPushProvider()

	// Simulate repository error
	mockRepo.err = ErrInternal

	usecase := NewNotifyStreamLive(mockRepo, mockProvider)

	input := NotifyStreamLiveInput{
		StreamID:    "stream_123",
		StreamTitle: "Title",
		OwnerUserID: "user1",
	}

	err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if mockProvider.calledCount != 0 {
		t.Fatalf("expected SendMulticast NOT to be called on error")
	}
}

func TestNotifyStreamLive_Execute_GenericMapInput(t *testing.T) {
	mockRepo := NewMockNotificationRepository()
	mockProvider := NewMockPushProvider()

	// Setup tokens
	userToken := domain.NewDeviceToken("id1", "user2", "token_user2", "android", "device1", "1.0")
	mockRepo.tokens["user2"] = []*domain.DeviceToken{userToken}

	usecase := NewNotifyStreamLive(mockRepo, mockProvider)

	// Execute with generic map input
	input := map[string]interface{}{
		"stream_id":     "stream_456",
		"stream_title":  "Another Stream",
		"owner_user_id": "user1",
	}

	err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if mockProvider.calledCount != 1 {
		t.Fatalf("expected SendMulticast to be called, got %d", mockProvider.calledCount)
	}

	if mockProvider.lastPayload.Data["stream_id"] != "stream_456" {
		t.Fatalf("expected stream_id to be stream_456, got %s", mockProvider.lastPayload.Data["stream_id"])
	}
}

func TestNotifyStreamLive_Execute_NoProvider(t *testing.T) {
	mockRepo := NewMockNotificationRepository()

	// No provider (nil)
	usecase := NewNotifyStreamLive(mockRepo, nil)

	input := NotifyStreamLiveInput{
		StreamID:    "stream_123",
		StreamTitle: "Title",
		OwnerUserID: "user1",
	}

	// Should not error, just skip
	err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("expected no error with nil provider, got %v", err)
	}
}
