package server

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Carlvalencia1/streamhub-backend/internal/platform/config"
	ws "github.com/Carlvalencia1/streamhub-backend/internal/platform/websocket"
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

	hub := ws.Manager
	go hub.Run()

	RegisterRoutes(router)
	RegisterWebSocketRoutes(router, hub)

	addr := fmt.Sprintf(":%s", s.cfg.Port)

	return http.ListenAndServe(addr, router)
}