package infrastructure

import (
	"context"
	"fmt"

	"firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"

	"github.com/Carlvalencia1/streamhub-backend/internal/notifications/domain"
	"github.com/Carlvalencia1/streamhub-backend/internal/platform/logger"
)

type FirebasePushProvider struct {
	client          *messaging.Client
	tokenRepository domain.NotificationRepository
}

// NewFirebasePushProvider inicializa el cliente de Firebase con la ruta a las credenciales
func NewFirebasePushProvider(credentialsPath string) (*FirebasePushProvider, error) {
	ctx := context.Background()

	// Validar que la ruta a las credenciales no esté vacía
	if credentialsPath == "" {
		logger.Error("CRITICAL: credentialsPath is empty. Firebase cannot be initialized.")
		return nil, fmt.Errorf("firebase credentials path is required")
	}

	logger.Info(fmt.Sprintf("Attempting to initialize Firebase with credentials from: %s", credentialsPath))

	// 🔥 CONFIGURACIÓN EXPLÍCITA
	// 1. Crear la opción con la ruta ABSOLUTA al archivo JSON
	opt := option.WithCredentialsFile(credentialsPath)
	
	// 2. Forzar el uso de la API v1 con el endpoint correcto
	opt = option.WithEndpoint("https://fcm.googleapis.com/v1")
	
	// 3. Configurar el proyecto manualmente
	conf := &firebase.Config{
		ProjectID: "streamhub-64704",
	}

	// 4. Inicializar la app de Firebase
	app, err := firebase.NewApp(ctx, conf, opt)
	if err != nil {
		logger.Error(fmt.Sprintf("error initializing Firebase app: %v", err))
		return nil, err
	}

	// 5. Crear el cliente de mensajería
	client, err := app.Messaging(ctx)
	if err != nil {
		logger.Error(fmt.Sprintf("error creating Firebase Messaging client: %v", err))
		return nil, err
	}

	logger.Info("Firebase Messaging client initialized successfully")
	return &FirebasePushProvider{client: client}, nil
}

// ... (el resto de tus funciones SetTokenRepository, SendMulticast, etc. se quedan IGUAL) ...