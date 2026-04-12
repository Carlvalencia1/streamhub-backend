package websocket

import (
	"encoding/json"
	"sync"
)

// StreamEventType define los tipos de eventos de stream
type StreamEventType string

const (
	StreamStarted StreamEventType = "stream_started"
	StreamStopped StreamEventType = "stream_stopped"
	ViewerJoined  StreamEventType = "viewer_joined"
	StreamCreated StreamEventType = "stream_created"
)

// StreamEvent representa un evento de stream
type StreamEvent struct {
	Type      StreamEventType `json:"type"`
	StreamID  string          `json:"stream_id"`
	Title     string          `json:"title"`
	Timestamp string          `json:"timestamp"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// StreamNotificationService maneja notificaciones de eventos de streams
type StreamNotificationService struct {
	subscribers map[string][]chan StreamEvent
	mu          sync.RWMutex
}

func NewStreamNotificationService() *StreamNotificationService {
	return &StreamNotificationService{
		subscribers: make(map[string][]chan StreamEvent),
	}
}

// Subscribe suscribe un canal a los eventos de un stream
func (s *StreamNotificationService) Subscribe(streamID string) <-chan StreamEvent {
	s.mu.Lock()
	defer s.mu.Unlock()

	eventChan := make(chan StreamEvent, 10)
	s.subscribers[streamID] = append(s.subscribers[streamID], eventChan)

	return eventChan
}

// Unsubscribe desuscribe un canal de los eventos de un stream
func (s *StreamNotificationService) Unsubscribe(streamID string, eventChan <-chan StreamEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if chans, ok := s.subscribers[streamID]; ok {
		for i, ch := range chans {
			if ch == eventChan {
				// Remover el canal de la lista
				s.subscribers[streamID] = append(chans[:i], chans[i+1:]...)
				close(ch.(chan StreamEvent))
				break
			}
		}
	}
}

// BroadcastStreamEvent envía un evento a todos los suscriptores de un stream
func (s *StreamNotificationService) BroadcastStreamEvent(event StreamEvent) {
	s.mu.RLock()
	chans := s.subscribers[event.StreamID]
	s.mu.RUnlock()

	for _, ch := range chans {
		select {
		case ch <- event:
		default:
			// Si el canal está lleno, ignorar
		}
	}
}

// BroadcastGlobalEvent envía un evento a todos los suscritos de múltiples streams
func (s *StreamNotificationService) BroadcastGlobalEvent(event StreamEvent) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, chans := range s.subscribers {
		for _, ch := range chans {
			select {
			case ch <- event:
			default:
			}
		}
	}
}

// GetSubscriberCount retorna el número de suscriptores para un stream
func (s *StreamNotificationService) GetSubscriberCount(streamID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.subscribers[streamID])
}

// MarshalJSON implementa json.Marshaler para Stream Event
func (e StreamEvent) MarshalJSON() ([]byte, error) {
	type Alias StreamEvent
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(&e),
	})
}
