package social

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"garden3/internal/domain/entity"
	"garden3/internal/port/output"
)

// TwitterAdapter handles Twitter API operations
type TwitterAdapter struct {
	configRepo output.ConfigurationRepository
}

// NewTwitterAdapter creates a new Twitter adapter
func NewTwitterAdapter(configRepo output.ConfigurationRepository) *TwitterAdapter {
	return &TwitterAdapter{
		configRepo: configRepo,
	}
}

// PostToTwitter posts content to Twitter using v2 API with OAuth 2.0
func (a *TwitterAdapter) PostToTwitter(ctx context.Context, content string) (string, error) {
	// Get client credentials
	clientID, err := a.getConfigValue(ctx, "twitter.client_id")
	if err != nil || clientID == "" {
		return "", errors.New("Twitter client ID not configured")
	}

	clientSecret, err := a.getConfigValue(ctx, "twitter.client_secret")
	if err != nil || clientSecret == "" {
		return "", errors.New("Twitter client secret not configured")
	}

	// Get or refresh OAuth token
	accessToken, err := a.getTwitterOAuthToken(ctx, clientID, clientSecret)
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}

	// Use Twitter's v2 API to create a tweet
	apiURL := "https://api.x.com/2/tweets"

	requestBody := map[string]string{
		"text": content,
	}

	bodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(string(bodyJSON)))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("Twitter API error: %s - %s", resp.Status, string(body))
	}

	var response struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Data.ID == "" {
		return "", errors.New("failed to get tweet ID from response")
	}

	return response.Data.ID, nil
}

// CheckCredentials checks if Twitter credentials are valid
func (a *TwitterAdapter) CheckCredentials(ctx context.Context) (*entity.TwitterProfile, error) {
	// Get client credentials
	clientID, err := a.getConfigValue(ctx, "twitter.client_id")
	if err != nil || clientID == "" {
		return nil, errors.New("Twitter client ID not configured")
	}

	clientSecret, err := a.getConfigValue(ctx, "twitter.client_secret")
	if err != nil || clientSecret == "" {
		return nil, errors.New("Twitter client secret not configured")
	}

	// Get or refresh OAuth token
	accessToken, err := a.getTwitterOAuthToken(ctx, clientID, clientSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Use Twitter's v2 API to get user information
	apiURL := "https://api.x.com/2/users/me?user.fields=name,username"

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Twitter credentials check failed: %s - %s", resp.Status, string(body))
	}

	var response struct {
		Data struct {
			Username string `json:"username"`
			Name     string `json:"name"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	username := response.Data.Username
	displayName := response.Data.Name

	return &entity.TwitterProfile{
		Username:    &username,
		DisplayName: &displayName,
	}, nil
}

// getTwitterOAuthToken gets a valid Twitter OAuth token, refreshing if needed
func (a *TwitterAdapter) getTwitterOAuthToken(ctx context.Context, clientID, clientSecret string) (string, error) {
	// Get stored tokens
	tokensJSON, err := a.getConfigValue(ctx, "twitter.tokens")
	if err != nil {
		return "", fmt.Errorf("failed to get tokens: %w", err)
	}

	if tokensJSON == "" {
		return "", errors.New("no valid Twitter access token. User needs to authorize via OAuth flow")
	}

	var tokens entity.TwitterTokens
	if err := json.Unmarshal([]byte(tokensJSON), &tokens); err != nil {
		return "", fmt.Errorf("failed to unmarshal tokens: %w", err)
	}

	// Check if we have a valid token that hasn't expired
	if tokens.AccessToken != "" && tokens.ExpiresIn > 0 {
		// Note: In production, we'd store the actual expiration timestamp
		// For now, we'll try to use the token and refresh on 401
		// Check if token is still valid (with 5 minute buffer)
		// This is a simplified version - a full implementation would store the timestamp
		return tokens.AccessToken, nil
	}

	// If we have a refresh token, try to refresh
	if tokens.RefreshToken != "" {
		newAccessToken, err := a.refreshTwitterOAuthToken(ctx, tokens.RefreshToken, clientID, clientSecret)
		if err != nil {
			return "", fmt.Errorf("failed to refresh token: %w", err)
		}
		return newAccessToken, nil
	}

	return "", errors.New("no valid Twitter access token. User needs to authorize via OAuth flow")
}

// refreshTwitterOAuthToken refreshes the Twitter OAuth 2.0 token
func (a *TwitterAdapter) refreshTwitterOAuthToken(ctx context.Context, refreshToken, clientID, clientSecret string) (string, error) {
	tokenURL := "https://api.x.com/2/oauth2/token"

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	authHeader := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+authHeader)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token refresh failed: %s - %s", resp.Status, string(body))
	}

	var newTokens entity.TwitterTokens
	if err := json.Unmarshal(body, &newTokens); err != nil {
		return "", fmt.Errorf("failed to unmarshal tokens: %w", err)
	}

	// If no new refresh token was provided, keep the old one
	if newTokens.RefreshToken == "" {
		newTokens.RefreshToken = refreshToken
	}

	// Store the refreshed tokens
	if err := a.UpdateTokens(ctx, newTokens); err != nil {
		return "", fmt.Errorf("failed to store refreshed tokens: %w", err)
	}

	return newTokens.AccessToken, nil
}

// UpdateTokens updates Twitter OAuth tokens
func (a *TwitterAdapter) UpdateTokens(ctx context.Context, tokens entity.TwitterTokens) error {
	tokensJSON, err := json.Marshal(tokens)
	if err != nil {
		return fmt.Errorf("failed to marshal tokens: %w", err)
	}

	_, err = a.configRepo.Upsert(ctx, "twitter.tokens", string(tokensJSON), true, time.Now())
	if err != nil {
		return fmt.Errorf("failed to store tokens: %w", err)
	}

	return nil
}

// InitiateTwitterAuth generates the Twitter OAuth authorization URL
func (a *TwitterAdapter) InitiateTwitterAuth(ctx context.Context) (*entity.TwitterAuthURL, error) {
	// Clean up expired states
	if err := a.cleanupExpiredStates(ctx); err != nil {
		return nil, fmt.Errorf("failed to cleanup expired states: %w", err)
	}

	// Get Twitter OAuth configuration
	clientID, err := a.getConfigValue(ctx, "twitter.client_id")
	if err != nil || clientID == "" {
		return nil, errors.New("Twitter client ID not configured")
	}

	// Generate state, code verifier, and code challenge
	state, err := generateRandomString(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}

	codeVerifier, err := generateRandomString(64)
	if err != nil {
		return nil, fmt.Errorf("failed to generate code verifier: %w", err)
	}

	codeChallenge := generateCodeChallenge(codeVerifier)

	// Store state and code verifier
	expiresAt := time.Now().Add(10 * time.Minute)
	oauthState := entity.TwitterOAuthState{
		State:        state,
		CodeVerifier: codeVerifier,
		ExpiresAt:    expiresAt,
	}

	stateJSON, err := json.Marshal(oauthState)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal state: %w", err)
	}

	_, err = a.configRepo.Upsert(ctx, fmt.Sprintf("twitter.state.%s", state), string(stateJSON), true, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to store state: %w", err)
	}

	// Get redirect URI from environment or use default
	redirectURI := os.Getenv("TWITTER_REDIRECT_URI")
	if redirectURI == "" {
		baseURL := os.Getenv("BASE_URL")
		if baseURL == "" {
			baseURL = "http://localhost:3000"
		}
		redirectURI = baseURL + "/api/social/twitter/callback"
	}

	// Construct the authorization URL
	authURL := url.URL{
		Scheme: "https",
		Host:   "x.com",
		Path:   "/i/oauth2/authorize",
	}

	query := url.Values{}
	query.Set("response_type", "code")
	query.Set("client_id", clientID)
	query.Set("redirect_uri", redirectURI)
	query.Set("scope", "tweet.read tweet.write users.read offline.access")
	query.Set("state", state)
	query.Set("code_challenge", codeChallenge)
	query.Set("code_challenge_method", "S256")
	authURL.RawQuery = query.Encode()

	return &entity.TwitterAuthURL{
		AuthorizationURL: authURL.String(),
		State:            state,
	}, nil
}

// HandleTwitterCallback processes the OAuth callback and exchanges code for tokens
func (a *TwitterAdapter) HandleTwitterCallback(ctx context.Context, input entity.TwitterCallbackInput) error {
	// Verify state to prevent CSRF attacks
	stateData, err := a.getOAuthState(ctx, input.State)
	if err != nil {
		return fmt.Errorf("invalid state: %w", err)
	}

	// Delete the state from the store to prevent replay attacks
	if err := a.configRepo.Delete(ctx, fmt.Sprintf("twitter.state.%s", input.State)); err != nil {
		// Log but don't fail on cleanup error
		fmt.Printf("Warning: failed to delete state: %v\n", err)
	}

	// Check if state has expired
	if time.Now().After(stateData.ExpiresAt) {
		return errors.New("state has expired")
	}

	// Get Twitter OAuth configuration
	clientID, err := a.getConfigValue(ctx, "twitter.client_id")
	if err != nil || clientID == "" {
		return errors.New("Twitter client ID not configured")
	}

	clientSecret, err := a.getConfigValue(ctx, "twitter.client_secret")
	if err != nil || clientSecret == "" {
		return errors.New("Twitter client secret not configured")
	}

	// Exchange authorization code for access token
	tokens, err := a.exchangeCodeForToken(ctx, input.Code, stateData.CodeVerifier, clientID, clientSecret)
	if err != nil {
		return fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Store the tokens
	if err := a.UpdateTokens(ctx, tokens); err != nil {
		return fmt.Errorf("failed to store tokens: %w", err)
	}

	return nil
}

// Helper methods

func (a *TwitterAdapter) getConfigValue(ctx context.Context, key string) (string, error) {
	config, err := a.configRepo.GetByKey(ctx, key)
	if err != nil {
		return "", err
	}
	if config == nil {
		return "", nil
	}
	return config.Value, nil
}

func (a *TwitterAdapter) getOAuthState(ctx context.Context, state string) (*entity.TwitterOAuthState, error) {
	config, err := a.configRepo.GetByKey(ctx, fmt.Sprintf("twitter.state.%s", state))
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, errors.New("state not found")
	}

	var oauthState entity.TwitterOAuthState
	if err := json.Unmarshal([]byte(config.Value), &oauthState); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return &oauthState, nil
}

func (a *TwitterAdapter) cleanupExpiredStates(ctx context.Context) error {
	prefix := "twitter.state."
	configs, err := a.configRepo.GetByPrefix(ctx, prefix, true)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, config := range configs {
		var oauthState entity.TwitterOAuthState
		if err := json.Unmarshal([]byte(config.Value), &oauthState); err != nil {
			continue
		}

		if oauthState.ExpiresAt.Before(now) {
			a.configRepo.Delete(ctx, config.Key)
		}
	}

	return nil
}

func (a *TwitterAdapter) exchangeCodeForToken(ctx context.Context, code, codeVerifier, clientID, clientSecret string) (entity.TwitterTokens, error) {
	tokenURL := "https://api.x.com/2/oauth2/token"

	redirectURI := os.Getenv("TWITTER_REDIRECT_URI")
	if redirectURI == "" {
		baseURL := os.Getenv("BASE_URL")
		if baseURL == "" {
			baseURL = "http://localhost:3000"
		}
		redirectURI = baseURL + "/api/social/twitter/callback"
	}

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("code_verifier", codeVerifier)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return entity.TwitterTokens{}, fmt.Errorf("failed to create request: %w", err)
	}

	authHeader := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+authHeader)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return entity.TwitterTokens{}, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return entity.TwitterTokens{}, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return entity.TwitterTokens{}, fmt.Errorf("token exchange failed: %s - %s", resp.Status, string(body))
	}

	var tokens entity.TwitterTokens
	if err := json.Unmarshal(body, &tokens); err != nil {
		return entity.TwitterTokens{}, fmt.Errorf("failed to unmarshal tokens: %w", err)
	}

	return tokens, nil
}

func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes)[:length], nil
}

func generateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
