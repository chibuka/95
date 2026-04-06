package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/chibuka/95-cli/internal/config"
	"github.com/google/uuid"
	"github.com/pkg/browser"
)

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserId       string `json:"id"`
	Login        string `json:"login"`
}

var ErrTimeout = errors.New("authentication timed out — please try again")

func Login() error {
	apiURL := getAPIURL()
	sessionID := uuid.New().String()

	fmt.Println("Opening browser for GitHub login...")
	err := browser.OpenURL(fmt.Sprintf("%s/oauth2/cli-login?session=%s", apiURL, sessionID))
	if err != nil {
		return fmt.Errorf("failed to open browser: %w", err)
	}

	fmt.Println("Waiting for authentication...")
	auth, err := pollForTokens(sessionID, apiURL)
	if err != nil {
		return err
	}

	cfg := config.Config{
		APIUrl:       apiURL,
		AccessToken:  auth.AccessToken,
		RefreshToken: auth.RefreshToken,
		UserId:       auth.UserId,
		Username:     auth.Login,
	}
	return cfg.Save()
}

func pollForTokens(sessionID, apiURL string) (*AuthResponse, error) {
	pollURL := fmt.Sprintf("%s/api/auth/cli/poll", apiURL)
	reqBody, err := json.Marshal(map[string]string{"session_id": sessionID})
	if err != nil {
		return nil, err
	}

	deadline := time.Now().Add(5 * time.Minute)
	for time.Now().Before(deadline) {
		res, err := http.Post(pollURL, "application/json", bytes.NewReader(reqBody))
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		if res.StatusCode == http.StatusAccepted {
			res.Body.Close()
			time.Sleep(2 * time.Second)
			continue
		}

		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(res.Body)
			return nil, fmt.Errorf("poll failed: %d - %s", res.StatusCode, body)
		}

		var auth AuthResponse
		if err := json.NewDecoder(res.Body).Decode(&auth); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		return &auth, nil
	}

	return nil, ErrTimeout
}

func getAPIURL() string {
	if os.Getenv("DEV_MODE") == "true" {
		return config.LocalAPIURL
	}
	return config.DefaultAPIURL
}

func RefreshToken(refreshToken string, apiURL string) (*AuthResponse, error) {
	reqBody := map[string]string{"refresh_token": refreshToken}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal refresh request: %w", err)
	}

	refreshURL := fmt.Sprintf("%s/api/auth/refresh", apiURL)
	res, err := http.Post(refreshURL, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to call refresh endpoint: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed: %d - %s", res.StatusCode, string(body))
	}

	var authResponse AuthResponse
	if err := json.Unmarshal(body, &authResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal refresh response: %w", err)
	}

	return &authResponse, nil
}

func Logout(refreshToken string, apiURL string) error {
	body, err := json.Marshal(map[string]string{"refresh_token": refreshToken})
	if err != nil {
		return fmt.Errorf("failed to marshal logout request: %w", err)
	}

	logoutURL := fmt.Sprintf("%s/api/auth/logout", apiURL)
	res, err := http.Post(logoutURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to call logout endpoint: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("logout failed: %d - %s", res.StatusCode, string(respBody))
	}

	return nil
}
