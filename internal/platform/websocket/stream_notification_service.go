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
	Type      StreamEventType            `json:"type"`
	StreamID  string                     `json:"stream_id"`
	Title     string                     `json:"title"`
	Timestamp string                     `json:"timestamp"`
	Details   map[string]interface{}     `json:"details,omitempty"`
}

// StreamNotificationService maneja notificaciones de eventos de streams
type StreamNotificationService struct {
	subscribers map[string]map[*chan StreamEvent]bool
	mu          sync.RWMutex
}

func NewStreamNotificationService() *StreamNotificationService {
	return &StreamNotificationService{
		subscribers: make(map[string]map[*chan StreamEvent]bool),
	}
}

// Subscribe suscribe un canal a los eventos de un stream
func (s *StreamNotificationService) Subscribe(streamID string) <-chan StreamEvent {
	s.mu.Lock()
	defer s.mu.Unlock()

	eventChan := make(chan StreamEvent, 10)
	
	if _, ok := s.subscribers[streamID]; !ok {
		s.subscribers[streamID] = make(map[*chan StreamEvent]bool)
	}
	
	s.subscribers[streamID][&eventChan] = true
	return eventChan
}

// Unsubscribe desuscribe un canal de los eventos de un stream
func (s *StreamNotificationService) Unsubscribe(streamID string, eventChan <-chan StreamEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if chans, ok := s.subscribers[streamID]; ok {
		for ch := range chans {
			if ch != nil {
				delete(chans, ch)
				close(*ch)
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

	for chPtr := range chans {
		if chPtr != nil {
			select {
			case *chPtr <- event:
			default:
				// Si el canal está lleno, ignorar
			}
		}
	}
}

// BroadcastGlobalEvent envía un evento a todos los suscritos de múltiples streams
func (s *StreamNotificationService) BroadcastGlobalEvent(event StreamEvent) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, chans := range s.subscribers {
		for chPtr := range chans {
			if chPtr != nil {
				select {
				case *chPtr <- event:
				default:
				}
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
