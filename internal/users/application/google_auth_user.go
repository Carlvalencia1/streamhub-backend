package application

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/Carlvalencia1/streamhub-backend/internal/platform/security"
	"github.com/Carlvalencia1/streamhub-backend/internal/users/domain"
)

type googleClaims struct {
	Sub     string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

type GoogleAuthUser struct {
	repo domain.Repository
}

func NewGoogleAuthUser(repo domain.Repository) *GoogleAuthUser {
	return &GoogleAuthUser{repo: repo}
}

func (uc *GoogleAuthUser) Execute(ctx context.Context, idToken string) (string, error) {
	claims, err := verifyGoogleToken(idToken)
	if err != nil {
		return "", err
	}

	if claims.Email == "" {
		return "", fmt.Errorf("email not present in google token")
	}

	user, err := uc.repo.GetByEmail(ctx, claims.Email)
	if err == nil {
		return security.GenerateToken(user.ID, user.Username)
	}

	username := claims.Name
	if username == "" {
		at := strings.Index(claims.Email, "@")
		if at > 0 {
			username = claims.Email[:at]
		} else {
			username = claims.Email
		}
	}

	hash, err := security.HashPassword(uuid.NewString())
	if err != nil {
		return "", fmt.Errorf("failed to generate password: %w", err)
	}

	newUser := &domain.User{
		ID:       uuid.NewString(),
		Username: username,
		Email:    claims.Email,
		Password: hash,
	}

	if claims.Picture != "" {
		newUser.AvatarURL = &claims.Picture
	}

	if err := uc.repo.Create(ctx, newUser); err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	return security.GenerateToken(newUser.ID, newUser.Username)
}

func verifyGoogleToken(idToken string) (*googleClaims, error) {
	resp, err := http.Get("https://oauth2.googleapis.com/tokeninfo?id_token=" + idToken)
	if err != nil {
		return nil, fmt.Errorf("failed to contact google: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read google response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid google token")
	}

	var claims googleClaims
	if err := json.Unmarshal(body, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse google claims: %w", err)
	}

	return &claims, nil
}
