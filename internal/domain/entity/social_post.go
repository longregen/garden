package entity

import (
	"time"

	"github.com/google/uuid"
)

// SocialPost represents a post to social media platforms
type SocialPost struct {
	PostID         uuid.UUID  `json:"post_id"`
	Content        string     `json:"content"`
	TwitterPostID  *string    `json:"twitter_post_id,omitempty"`
	BlueskyPostID  *string    `json:"bluesky_post_id,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at,omitempty"`
	Status         string     `json:"status"`
	ErrorMessage   *string    `json:"error_message,omitempty"`
}

// SocialPostFilters represents filter criteria for listing social posts
type SocialPostFilters struct {
	Status *string
	Page   int32
	Limit  int32
}

// CreateSocialPostInput represents input for creating a social post
type CreateSocialPostInput struct {
	Content string `json:"content"`
}

// UpdateStatusInput represents input for updating post status
type UpdateStatusInput struct {
	Status       string
	ErrorMessage *string
}

// PostResult represents the result of posting to platforms
type PostResult struct {
	PostID        uuid.UUID `json:"post_id"`
	TwitterPostID *string   `json:"twitter_post_id,omitempty"`
	BlueskyPostID *string   `json:"bluesky_post_id,omitempty"`
	Status        string    `json:"status"`
	ErrorMessage  *string   `json:"error_message,omitempty"`
}

// CredentialsStatus represents the status of social media credentials
type CredentialsStatus struct {
	Twitter TwitterCredentials `json:"twitter"`
	Bluesky BlueskyCredentials `json:"bluesky"`
}

// TwitterCredentials represents Twitter credential status
type TwitterCredentials struct {
	Configured bool                 `json:"configured"`
	Working    bool                 `json:"working"`
	Error      *string              `json:"error,omitempty"`
	Profile    *TwitterProfile      `json:"profile,omitempty"`
}

// TwitterProfile represents Twitter user profile
type TwitterProfile struct {
	Username    *string `json:"username,omitempty"`
	DisplayName *string `json:"display_name,omitempty"`
}

// BlueskyCredentials represents Bluesky credential status
type BlueskyCredentials struct {
	Configured bool            `json:"configured"`
	Working    bool            `json:"working"`
	Error      *string         `json:"error,omitempty"`
	Profile    *BlueskyProfile `json:"profile,omitempty"`
}

// BlueskyProfile represents Bluesky user profile
type BlueskyProfile struct {
	DID    *string `json:"did,omitempty"`
	Handle *string `json:"handle,omitempty"`
}

// TwitterTokens represents Twitter OAuth tokens
type TwitterTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
}

// TwitterAuthURL represents the authorization URL response
type TwitterAuthURL struct {
	AuthorizationURL string `json:"authorization_url"`
	State            string `json:"state"`
}

// TwitterCallbackInput represents the callback parameters
type TwitterCallbackInput struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

// TwitterOAuthState represents stored OAuth state data
type TwitterOAuthState struct {
	State        string
	CodeVerifier string
	ExpiresAt    time.Time
}
