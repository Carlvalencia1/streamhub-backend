package server

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Carlvalencia1/streamhub-backend/internal/platform/config"
	ws "github.com/Carlvalencia1/streamhub-backend/internal/platform/websocket"
	"github.com/Carlvalencia1/streamhub-backend/internal/platform/middleware"
)

type Server struct {
	cfg *config.Config
	db  *sql.DB
}

func NewServer(cfg *config.Config, db *sql.DB) *Server {
	return &Server{
		cfg: cfg,
		db:  db,
	}
}

func (s *Server) Start() error {

	router := gin.Default()

	// 🔥 Manager WebSocket (salas por stream)
	manager := ws.NewManager()

	// Rutas HTTP normales
	RegisterRoutes(router, s.cfg, s.db)

	// 🔥 Rutas WebSocket con middleware JWT
	RegisterWebSocketRoutes(
		router,
		manager,
		s.db,
		middleware.AuthMiddleware(), // ← AQUÍ ESTÁ LA CLAVE
	)

	addr := fmt.Sprintf(":%s", s.cfg.Port)

	return http.ListenAndServe(addr, router)
}