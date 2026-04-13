package webrtc

import (
	"encoding/json"
	"sync"

	"github.com/Carlvalencia1/streamhub-backend/internal/platform/logger"
)

// SignalingServer maneja la coordinación de conexiones WebRTC
type SignalingServer struct {
	// Sessions: streamID -> BroadcastSession
	Sessions map[string]*BroadcastSession
	mu       sync.RWMutex

	// PendingOffers: streamID+fromUserID -> SignalingMessage
	PendingOffers map[string]*SignalingMessage
	offersMu      sync.RWMutex

	// Channels para comunicación
	MessageChan chan *SignalingMessage
	StopChan    chan struct{}
}

// NewSignalingServer crea un nuevo servidor de signaling
func NewSignalingServer() *SignalingServer {
	return &SignalingServer{
		Sessions:     make(map[string]*BroadcastSession),
		PendingOffers: make(map[string]*SignalingMessage),
		MessageChan:  make(chan *SignalingMessage, 100),
		StopChan:     make(chan struct{}),
	}
}

// StartBroadcast inicia una nueva sesión de transmisión
func (s *SignalingServer) StartBroadcast(streamID, broadcasterID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.Sessions[streamID]; exists {
		logger.Warn("Stream already broadcasting: " + streamID)
		return nil // Ya existe, no es error
	}

	session := &BroadcastSession{
		StreamID:      streamID,
		BroadcasterID: broadcasterID,
		Viewers:       make(map[string]bool),
		IsActive:      true,
	}

	s.Sessions[streamID] = session
	logger.StreamEvent("BROADCAST_STARTED", streamID, "Broadcaster: "+broadcasterID)

	return nil
}

// StopBroadcast detiene una sesión de transmisión
func (s *SignalingServer) StopBroadcast(streamID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.Sessions[streamID]
	if !exists {
		logger.Warn("Stream not found for stop: " + streamID)
		return nil
	}

	session.IsActive = false
	delete(s.Sessions, streamID)

	logger.StreamEvent("BROADCAST_STOPPED", streamID, "Viewers: "+itoa(len(session.Viewers)))

	return nil
}

// RegisterViewer agrega un espectador a la sesión
func (s *SignalingServer) RegisterViewer(streamID, viewerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.Sessions[streamID]
	if !exists {
		logger.Warn("Stream not found for viewer registration: " + streamID)
		return nil
	}

	session.Viewers[viewerID] = true
	logger.Debug("Viewer registered: " + viewerID + " to stream: " + streamID)

	return nil
}

// UnregisterViewer elimina un espectador de la sesión
func (s *SignalingServer) UnregisterViewer(streamID, viewerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.Sessions[streamID]
	if !exists {
		return nil
	}

	delete(session.Viewers, viewerID)
	logger.Debug("Viewer unregistered: " + viewerID + " from stream: " + streamID)

	return nil
}

// ProcessSignalingMessage procesa mensajes WebRTC
func (s *SignalingServer) ProcessSignalingMessage(msg *SignalingMessage) error {
	switch msg.Type {
	case TypeOfferSDP:
		return s.handleOfferSDP(msg)

	case TypeAnswerSDP:
		return s.handleAnswerSDP(msg)

	case TypeICECandidate:
		return s.handleICECandidate(msg)

	default:
		logger.Warn("Unknown signaling message type: " + string(msg.Type))
	}

	return nil
}

// handleOfferSDP guarda la oferta SDP del transmisor
func (s *SignalingServer) handleOfferSDP(msg *SignalingMessage) error {
	key := msg.StreamID + ":" + msg.FromUserID

	s.offersMu.Lock()
	s.PendingOffers[key] = msg
	s.offersMu.Unlock()

	logger.Debug("Offer SDP received for stream: " + msg.StreamID)
	s.MessageChan <- msg

	return nil
}

// handleAnswerSDP procesa la respuesta SDP del espectador
func (s *SignalingServer) handleAnswerSDP(msg *SignalingMessage) error {
	logger.Debug("Answer SDP received for stream: " + msg.StreamID)
	s.MessageChan <- msg
	return nil
}

// handleICECandidate procesa candidatos ICE
func (s *SignalingServer) handleICECandidate(msg *SignalingMessage) error {
	logger.Debug("ICE Candidate received for stream: " + msg.StreamID)
	s.MessageChan <- msg
	return nil
}

// GetBroadcaster obtiene el ID del transmisor de un stream
func (s *SignalingServer) GetBroadcaster(streamID string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.Sessions[streamID]
	if !exists {
		return "", false
	}

	return session.BroadcasterID, true
}

// IsBroadcasting verifica si un stream está en vivo
func (s *SignalingServer) IsBroadcasting(streamID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.Sessions[streamID]
	if !exists {
		return false
	}

	return session.IsActive
}

// GetViewerCount retorna el número de espectadores
func (s *SignalingServer) GetViewerCount(streamID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.Sessions[streamID]
	if !exists {
		return 0
	}

	return len(session.Viewers)
}

// BroadcastSignalingMessage envía un mensaje a todos los espectadores
func (s *SignalingServer) BroadcastSignalingMessage(streamID string, msg *SignalingMessage) error {
	s.mu.RLock()
	session, exists := s.Sessions[streamID]
	s.mu.RUnlock()

	if !exists {
		return nil
	}

	// Serializar mensaje
	data, err := json.Marshal(msg)
	if err != nil {
		logger.ErrorWithContext("SignalingServer", "failed to marshal message", err)
		return err
	}

	logger.Debug("Broadcasting signaling message for stream: " + streamID)
	s.MessageChan <- msg

	return nil
}

// Helper
func itoa(i int) string {
	return json.Number(string(rune(i))).String()
}
