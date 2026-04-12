package application

import (
	"context"
	"errors"
	"testing"

	"github.com/Carlvalencia1/streamhub-backend/internal/streams/domain"
)

// Mock Repository para tests
type mockRepository struct {
	createFunc       func(ctx context.Context, stream *domain.Stream) error
	getByIDFunc      func(ctx context.Context, streamID string) (*domain.Stream, error)
	startStreamFunc  func(ctx context.Context, streamID string) error
	stopStreamFunc   func(ctx context.Context, streamID string) error
	joinStreamFunc   func(ctx context.Context, streamID string) error
}

func (m *mockRepository) Create(ctx context.Context, stream *domain.Stream) error {
	return m.createFunc(ctx, stream)
}

func (m *mockRepository) GetAll(ctx context.Context) ([]*domain.Stream, error) {
	return nil, nil
}

func (m *mockRepository) GetByID(ctx context.Context, streamID string) (*domain.Stream, error) {
	return m.getByIDFunc(ctx, streamID)
}

func (m *mockRepository) StartStream(ctx context.Context, streamID string) error {
	return m.startStreamFunc(ctx, streamID)
}

func (m *mockRepository) StopStream(ctx context.Context, streamID string) error {
	return m.stopStreamFunc(ctx, streamID)
}

func (m *mockRepository) JoinStream(ctx context.Context, streamID string) error {
	return m.joinStreamFunc(ctx, streamID)
}

// Tests para CreateStream
func TestCreateStreamSuccess(t *testing.T) {
	mockRepo := &mockRepository{
		createFunc: func(ctx context.Context, stream *domain.Stream) error {
			return nil
		},
	}

	uc := NewCreateStream(mockRepo)
	ctx := context.Background()

	stream, err := uc.Execute(
		ctx,
		"Test Stream",
		"Test Description",
		"https://example.com/thumb.jpg",
		"gaming",
		"user123",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if stream == nil {
		t.Fatal("expected stream, got nil")
	}

	if stream.Title != "Test Stream" {
		t.Errorf("expected title 'Test Stream', got '%s'", stream.Title)
	}

	if stream.OwnerID != "user123" {
		t.Errorf("expected owner_id 'user123', got '%s'", stream.OwnerID)
	}

	if stream.StreamKey == "" {
		t.Error("expected non-empty stream_key")
	}

	if stream.PlaybackURL == "" {
		t.Error("expected non-empty playback_url")
	}

	if stream.IsLive {
		t.Error("expected IsLive to be false on creation")
	}
}

func TestCreateStreamDBError(t *testing.T) {
	mockRepo := &mockRepository{
		createFunc: func(ctx context.Context, stream *domain.Stream) error {
			return errors.New("database error")
		},
	}

	uc := NewCreateStream(mockRepo)
	ctx := context.Background()

	stream, err := uc.Execute(ctx, "Test", "Desc", "https://example.com/thumb.jpg", "gaming", "user123")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if stream != nil {
		t.Fatal("expected nil stream on error")
	}
}

// Tests para GetStreamByID
func TestGetStreamByIDSuccess(t *testing.T) {
	testStream := &domain.Stream{
		ID:    "stream123",
		Title: "Test Stream",
	}

	mockRepo := &mockRepository{
		getByIDFunc: func(ctx context.Context, streamID string) (*domain.Stream, error) {
			if streamID == "stream123" {
				return testStream, nil
			}
			return nil, errors.New("not found")
		},
	}

	uc := NewGetStreamByID(mockRepo)
	ctx := context.Background()

	stream, err := uc.Execute(ctx, "stream123")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if stream == nil {
		t.Fatal("expected stream, got nil")
	}

	if stream.ID != "stream123" {
		t.Errorf("expected id 'stream123', got '%s'", stream.ID)
	}
}

func TestGetStreamByIDNotFound(t *testing.T) {
	mockRepo := &mockRepository{
		getByIDFunc: func(ctx context.Context, streamID string) (*domain.Stream, error) {
			return nil, errors.New("stream not found")
		},
	}

	uc := NewGetStreamByID(mockRepo)
	ctx := context.Background()

	stream, err := uc.Execute(ctx, "nonexistent")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if stream != nil {
		t.Fatal("expected nil stream on error")
	}
}

// Tests para StartStream
func TestStartStreamSuccess(t *testing.T) {
	mockRepo := &mockRepository{
		startStreamFunc: func(ctx context.Context, streamID string) error {
			return nil
		},
	}

	uc := NewStartStream(mockRepo)
	ctx := context.Background()

	err := uc.Execute(ctx, "stream123")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestStartStreamError(t *testing.T) {
	mockRepo := &mockRepository{
		startStreamFunc: func(ctx context.Context, streamID string) error {
			return errors.New("failed to update")
		},
	}

	uc := NewStartStream(mockRepo)
	ctx := context.Background()

	err := uc.Execute(ctx, "stream123")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// Tests para StopStream
func TestStopStreamSuccess(t *testing.T) {
	mockRepo := &mockRepository{
		stopStreamFunc: func(ctx context.Context, streamID string) error {
			return nil
		},
	}

	uc := NewStopStream(mockRepo)
	ctx := context.Background()

	err := uc.Execute(ctx, "stream123")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestStopStreamError(t *testing.T) {
	mockRepo := &mockRepository{
		stopStreamFunc: func(ctx context.Context, streamID string) error {
			return errors.New("failed to update")
		},
	}

	uc := NewStopStream(mockRepo)
	ctx := context.Background()

	err := uc.Execute(ctx, "stream123")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
