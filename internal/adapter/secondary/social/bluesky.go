package social

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"garden3/internal/domain/entity"
	"garden3/internal/port/output"
)

const (
	blueskyAPIBase = "https://bsky.social/xrpc"
)

// BlueskyAdapter handles Bluesky API operations
type BlueskyAdapter struct {
	configRepo output.ConfigurationRepository
}

// BlueskySession represents a Bluesky session
type BlueskySession struct {
	AccessJWT      string    `json:"accessJwt"`
	RefreshJWT     string    `json:"refreshJwt"`
	DID            string    `json:"did"`
	Handle         string    `json:"handle"`
	SessionCreated time.Time `json:"sessionCreated"`
}

// NewBlueskyAdapter creates a new Bluesky adapter
func NewBlueskyAdapter(configRepo output.ConfigurationRepository) *BlueskyAdapter {
	return &BlueskyAdapter{
		configRepo: configRepo,
	}
}

// PostToBluesky posts content to Bluesky
func (a *BlueskyAdapter) PostToBluesky(ctx context.Context, content string) (string, error) {
	session, err := a.getBlueskySession(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get session: %w", err)
	}

	postID, err := a.postWithSession(ctx, content, session)
	if err != nil {
		// If posting failed with auth error and we have a refresh token, try refreshing
		if isAuthError(err) && session.RefreshJWT != "" {
			refreshed, refreshErr := a.refreshBlueskySession(ctx, session)
			if refreshErr == nil {
				// Retry with refreshed session
				return a.postWithSession(ctx, content, refreshed)
			}
		}
		return "", err
	}

	return postID, nil
}

// postWithSession performs the actual post operation with a given session
func (a *BlueskyAdapter) postWithSession(ctx context.Context, content string, session *BlueskySession) (string, error) {
	postRecord := map[string]interface{}{
		"$type":     "app.bsky.feed.post",
		"text":      content,
		"createdAt": time.Now().UTC().Format(time.RFC3339),
	}

	requestBody := map[string]interface{}{
		"repo":       session.DID,
		"collection": "app.bsky.feed.post",
		"record":     postRecord,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", blueskyAPIBase+"/com.atproto.repo.createRecord", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+session.AccessJWT)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("post failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		URI string `json:"uri"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if result.URI == "" {
		return "", errors.New("no URI in response")
	}

	// Extract post ID from URI (at://did:plc:xxx/app.bsky.feed.post/postid)
	postID := result.URI
	if len(result.URI) > 0 {
		// Try to extract just the last part
		parts := []rune(result.URI)
		for i := len(parts) - 1; i >= 0; i-- {
			if parts[i] == '/' {
				postID = string(parts[i+1:])
				break
			}
		}
	}

	// Update config with successful post
	if err := a.updateLastPost(ctx, postID); err != nil {
		// Log but don't fail
		fmt.Printf("Warning: failed to update last post: %v\n", err)
	}

	return postID, nil
}

// CheckCredentials checks if Bluesky credentials are valid
func (a *BlueskyAdapter) CheckCredentials(ctx context.Context) (*entity.BlueskyProfile, error) {
	session, err := a.getBlueskySession(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Use the session to get profile information
	req, err := http.NewRequestWithContext(ctx, "GET", blueskyAPIBase+"/com.atproto.repo.describeRepo?repo="+session.DID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+session.AccessJWT)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("credential check failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		DID    string `json:"did"`
		Handle string `json:"handle"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	did := result.DID
	handle := result.Handle

	return &entity.BlueskyProfile{
		DID:    &did,
		Handle: &handle,
	}, nil
}

// getBlueskySession gets or creates a Bluesky session
func (a *BlueskyAdapter) getBlueskySession(ctx context.Context) (*BlueskySession, error) {
	// Try to get existing session
	sessionConfig, err := a.configRepo.GetByKey(ctx, "bluesky.session")
	if err == nil && sessionConfig != nil {
		var session BlueskySession
		if err := json.Unmarshal([]byte(sessionConfig.Value), &session); err == nil {
			// Check if session is still valid (less than 24 hours old)
			if time.Since(session.SessionCreated) < 24*time.Hour {
				return &session, nil
			}

			// Try to refresh the session if we have a refresh token
			if session.RefreshJWT != "" {
				refreshed, err := a.refreshBlueskySession(ctx, &session)
				if err == nil {
					return refreshed, nil
				}
				// If refresh fails, fall through to create new session
				fmt.Printf("Warning: failed to refresh session: %v\n", err)
			}
		}
	}

	// Need to create new session
	return a.loginToBluesky(ctx)
}

// loginToBluesky creates a new Bluesky session
func (a *BlueskyAdapter) loginToBluesky(ctx context.Context) (*BlueskySession, error) {
	identifier, err := a.getConfigValue(ctx, "bluesky.identifier")
	if err != nil || identifier == "" {
		return nil, errors.New("Bluesky identifier not configured")
	}

	password, err := a.getConfigValue(ctx, "bluesky.password")
	if err != nil || password == "" {
		return nil, errors.New("Bluesky password not configured")
	}

	requestBody := map[string]string{
		"identifier": identifier,
		"password":   password,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", blueskyAPIBase+"/com.atproto.server.createSession", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		AccessJWT  string `json:"accessJwt"`
		RefreshJWT string `json:"refreshJwt"`
		DID        string `json:"did"`
		Handle     string `json:"handle"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	session := &BlueskySession{
		AccessJWT:      result.AccessJWT,
		RefreshJWT:     result.RefreshJWT,
		DID:            result.DID,
		Handle:         result.Handle,
		SessionCreated: time.Now(),
	}

	// Store the session
	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal session: %w", err)
	}

	_, err = a.configRepo.Upsert(ctx, "bluesky.session", string(sessionJSON), true, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	return session, nil
}

// refreshBlueskySession refreshes a Bluesky session using the refresh token
func (a *BlueskyAdapter) refreshBlueskySession(ctx context.Context, session *BlueskySession) (*BlueskySession, error) {
	if session == nil || session.RefreshJWT == "" {
		return nil, errors.New("no refresh token available")
	}

	req, err := http.NewRequestWithContext(ctx, "POST", blueskyAPIBase+"/com.atproto.server.refreshSession", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+session.RefreshJWT)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		AccessJWT  string `json:"accessJwt"`
		RefreshJWT string `json:"refreshJwt"`
		DID        string `json:"did"`
		Handle     string `json:"handle"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	newSession := &BlueskySession{
		AccessJWT:      result.AccessJWT,
		RefreshJWT:     result.RefreshJWT,
		DID:            result.DID,
		Handle:         result.Handle,
		SessionCreated: time.Now(),
	}

	// Store the refreshed session
	sessionJSON, err := json.Marshal(newSession)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal session: %w", err)
	}

	_, err = a.configRepo.Upsert(ctx, "bluesky.session", string(sessionJSON), true, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to store refreshed session: %w", err)
	}

	return newSession, nil
}

// Helper methods

func (a *BlueskyAdapter) getConfigValue(ctx context.Context, key string) (string, error) {
	config, err := a.configRepo.GetByKey(ctx, key)
	if err != nil {
		return "", err
	}
	if config == nil {
		return "", nil
	}
	return config.Value, nil
}

func (a *BlueskyAdapter) updateLastPost(ctx context.Context, postID string) error {
	_, err := a.configRepo.Upsert(ctx, "bluesky.last_post_id", postID, false, time.Now())
	return err
}

// isAuthError checks if an error is an authentication error
func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return bytes.Contains([]byte(errStr), []byte("401")) ||
		bytes.Contains([]byte(errStr), []byte("403")) ||
		bytes.Contains([]byte(errStr), []byte("Unauthorized")) ||
		bytes.Contains([]byte(errStr), []byte("authentication"))
}
