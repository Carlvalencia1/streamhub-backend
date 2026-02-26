package websocket

import "sync"

type Manager struct {
	hubs map[string]*Hub
	mu   sync.Mutex
}

func NewManager() *Manager {
	return &Manager{
		hubs: make(map[string]*Hub),
	}
}

func (m *Manager) GetHub(streamID string) *Hub {
	m.mu.Lock()
	defer m.mu.Unlock()

	hub, ok := m.hubs[streamID]
	if !ok {
		hub = NewHub(streamID)
		m.hubs[streamID] = hub
		go hub.Run()
	}

	return hub
}