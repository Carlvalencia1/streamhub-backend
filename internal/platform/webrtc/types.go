package webrtc

// WebRTC Signaling Message Types
type MessageType string

const (
	// Signaling messages
	TypeOfferSDP    MessageType = "offer_sdp"
	TypeAnswerSDP   MessageType = "answer_sdp"
	TypeICECandidate MessageType = "ice_candidate"
	TypeStartBroadcast MessageType = "start_broadcast"
	TypeStopBroadcast  MessageType = "stop_broadcast"
	TypeBroadcasterReady MessageType = "broadcaster_ready"
	TypeError        MessageType = "error"
)

// SignalingMessage es el mensaje que se intercambia entre transmisor y espectadores
type SignalingMessage struct {
	Type        MessageType `json:"type"`
	StreamID    string      `json:"stream_id"`
	FromUserID  string      `json:"from_user_id"`
	ToUserID    string      `json:"to_user_id,omitempty"`
	SDP         string      `json:"sdp,omitempty"`
	Candidate   *ICECandidate `json:"candidate,omitempty"`
	Error       string      `json:"error,omitempty"`
}

// ICECandidate para intercambio de candidatos
type ICECandidate struct {
	Candidate        string `json:"candidate"`
	SDPMLineIndex    int    `json:"sdp_m_line_index,omitempty"`
	SDPMid           string `json:"sdp_mid,omitempty"`
}

// BroadcastSession información de una transmisión en vivo
type BroadcastSession struct {
	StreamID      string            // ID del stream
	BroadcasterID string            // ID del usuario que transmite
	Viewers       map[string]bool    // Espectadores conectados
	IsActive      bool              // ¿Está transmitiendo?
}
