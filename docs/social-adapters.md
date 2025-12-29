# Social Media Adapters Documentation

## Overview

The social media integration system provides a unified interface for posting content to multiple social media platforms (Twitter and Bluesky) through a clean adapter pattern. The system follows hexagonal architecture principles, with adapters implementing the `SocialMediaService` interface defined in the output ports.

### Architecture

```
┌──────────────────────────────────────┐
│      SocialMediaService Interface    │
│          (Port/Output)                │
└──────────────┬───────────────────────┘
               │
               │ implements
               │
┌──────────────▼───────────────────────┐
│         Social Service               │
│      (Facade/Aggregator)             │
└──────────┬───────────┬───────────────┘
           │           │
           │           │
    ┌──────▼──────┐   │
    │  Bluesky    │   │
    │  Adapter    │   │
    └─────────────┘   │
                      │
               ┌──────▼──────┐
               │   Twitter   │
               │   Adapter   │
               └─────────────┘
```

**Location**: `/home/user/garden/internal/adapter/secondary/social/`

### Key Components

1. **service.go** - Service abstraction layer that aggregates Twitter and Bluesky adapters
2. **twitter.go** - Twitter adapter implementing OAuth 2.0 and Twitter API v2
3. **bluesky.go** - Bluesky adapter implementing AT Protocol authentication

---

## Service Abstraction Layer

### Overview

The `Service` struct acts as a facade that aggregates both Twitter and Bluesky adapters, providing a unified interface for the application layer.

**File**: `/home/user/garden/internal/adapter/secondary/social/service.go`

### Structure

```go
type Service struct {
    twitter *TwitterAdapter
    bluesky *BlueskyAdapter
}
```

### Factory Method

```go
func NewService(configRepo output.ConfigurationRepository) output.SocialMediaService
```

Creates a new social media service with both Twitter and Bluesky adapters initialized.

### Interface Methods

The service implements the `SocialMediaService` interface with the following methods:

| Method | Description | Returns |
|--------|-------------|---------|
| `PostToTwitter(ctx, content)` | Posts content to Twitter | Tweet ID, error |
| `PostToBluesky(ctx, content)` | Posts content to Bluesky | Post ID, error |
| `CheckTwitterCredentials(ctx)` | Validates Twitter credentials | TwitterProfile, error |
| `CheckBlueskyCredentials(ctx)` | Validates Bluesky credentials | BlueskyProfile, error |
| `UpdateTwitterTokens(ctx, tokens)` | Updates stored Twitter OAuth tokens | error |
| `InitiateTwitterAuth(ctx)` | Generates Twitter OAuth authorization URL | TwitterAuthURL, error |
| `HandleTwitterCallback(ctx, input)` | Processes Twitter OAuth callback | error |

### Design Benefits

- **Single Responsibility**: Each adapter handles only its platform
- **Dependency Injection**: Accepts `ConfigurationRepository` for configuration management
- **Interface Compliance**: Implements the port interface for easy testing and mocking
- **Extensibility**: New platforms can be added without modifying existing code

---

## Bluesky Integration

### Overview

The Bluesky adapter implements integration with the Bluesky social network using the AT Protocol (Authenticated Transfer Protocol). It handles session management, authentication, token refresh, and posting operations.

**File**: `/home/user/garden/internal/adapter/secondary/social/bluesky.go`

### Configuration

Bluesky requires the following configuration keys:

| Configuration Key | Description | Sensitive |
|------------------|-------------|-----------|
| `bluesky.identifier` | User identifier (handle or email) | No |
| `bluesky.password` | User password | Yes |
| `bluesky.session` | Stored session data (JWT tokens) | Yes |
| `bluesky.last_post_id` | ID of last successful post | No |

### Authentication Flow

```
┌────────────┐
│   Login    │
│  Request   │
└──────┬─────┘
       │
       ▼
┌────────────────────────────────────┐
│  POST com.atproto.server.createSession │
│  Body: {identifier, password}      │
└──────┬─────────────────────────────┘
       │
       ▼
┌────────────────────────────────────┐
│  Response: {accessJwt, refreshJwt, │
│            did, handle}            │
└──────┬─────────────────────────────┘
       │
       ▼
┌────────────────────────────────────┐
│  Store session with timestamp      │
│  Valid for 24 hours                │
└────────────────────────────────────┘
```

### Session Management

The adapter implements intelligent session management:

1. **Session Retrieval**: Checks for existing valid session (< 24 hours old)
2. **Session Refresh**: If session exists but expired, attempts refresh with `refreshJwt`
3. **New Login**: If no session or refresh fails, creates new session

```go
type BlueskySession struct {
    AccessJWT      string    `json:"accessJwt"`
    RefreshJWT     string    `json:"refreshJwt"`
    DID            string    `json:"did"`           // Decentralized Identifier
    Handle         string    `json:"handle"`        // User handle
    SessionCreated time.Time `json:"sessionCreated"`
}
```

### API Endpoints

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `https://bsky.social/xrpc/com.atproto.server.createSession` | POST | Create new session |
| `https://bsky.social/xrpc/com.atproto.server.refreshSession` | POST | Refresh existing session |
| `https://bsky.social/xrpc/com.atproto.repo.createRecord` | POST | Create a post |
| `https://bsky.social/xrpc/com.atproto.repo.describeRepo` | GET | Get repository info |

### Posting to Bluesky

#### Post Structure

```go
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
```

#### Post Flow with Error Recovery

```go
func (a *BlueskyAdapter) PostToBluesky(ctx context.Context, content string) (string, error)
```

1. Get or create session
2. Attempt post with current session
3. If auth error and refresh token available:
   - Refresh session
   - Retry post with new session
4. Extract post ID from AT URI
5. Update last post configuration

#### Post ID Extraction

Bluesky returns an AT URI like: `at://did:plc:xxx/app.bsky.feed.post/postid`

The adapter extracts just the post ID (last segment after final `/`) for easier reference.

### Credential Validation

```go
func (a *BlueskyAdapter) CheckCredentials(ctx context.Context) (*entity.BlueskyProfile, error)
```

1. Obtains valid session
2. Calls `com.atproto.repo.describeRepo` with user's DID
3. Returns profile information (DID and handle)

### Error Handling

The adapter includes sophisticated error detection:

```go
func isAuthError(err error) bool {
    // Checks for HTTP 401, 403, "Unauthorized", or "authentication" in error message
}
```

This enables automatic session refresh on authentication failures.

### Key Features

- **Automatic Session Management**: Sessions cached for 24 hours, refreshed automatically
- **Token Refresh**: Uses refresh JWT to extend session without re-authentication
- **Error Recovery**: Retries failed posts after session refresh
- **Secure Storage**: Sessions stored encrypted in configuration repository
- **Timeout Handling**: 30-second timeout on all HTTP requests

---

## Twitter Integration

### Overview

The Twitter adapter implements integration with Twitter (X) using OAuth 2.0 with PKCE (Proof Key for Code Exchange) and the Twitter API v2. It provides a complete OAuth flow implementation with token refresh capabilities.

**File**: `/home/user/garden/internal/adapter/secondary/social/twitter.go`

### Configuration

Twitter requires the following configuration:

| Configuration Key | Description | Sensitive |
|------------------|-------------|-----------|
| `twitter.client_id` | OAuth 2.0 client ID | No |
| `twitter.client_secret` | OAuth 2.0 client secret | Yes |
| `twitter.tokens` | Stored OAuth tokens (JSON) | Yes |
| `twitter.state.{state}` | OAuth state data (temporary) | Yes |

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `TWITTER_REDIRECT_URI` | OAuth callback URL | `{BASE_URL}/api/social/twitter/callback` |
| `BASE_URL` | Application base URL | `http://localhost:3000` |

### OAuth 2.0 with PKCE Flow

Twitter uses OAuth 2.0 with PKCE for enhanced security, especially important for public clients.

```
┌────────────────────────────────────────────────────────────┐
│                    1. Initiate Auth                        │
│  App generates: state, code_verifier, code_challenge       │
└────────────────┬───────────────────────────────────────────┘
                 │
                 ▼
┌────────────────────────────────────────────────────────────┐
│          2. Redirect User to Twitter                       │
│  URL: https://x.com/i/oauth2/authorize                     │
│  Params: client_id, redirect_uri, state,                   │
│          code_challenge, code_challenge_method=S256        │
└────────────────┬───────────────────────────────────────────┘
                 │
                 ▼
┌────────────────────────────────────────────────────────────┐
│           3. User Authorizes on Twitter                    │
└────────────────┬───────────────────────────────────────────┘
                 │
                 ▼
┌────────────────────────────────────────────────────────────┐
│         4. Twitter Redirects to Callback                   │
│  Params: code, state                                       │
└────────────────┬───────────────────────────────────────────┘
                 │
                 ▼
┌────────────────────────────────────────────────────────────┐
│         5. Exchange Code for Tokens                        │
│  POST https://api.x.com/2/oauth2/token                     │
│  Body: code, code_verifier, grant_type, redirect_uri      │
│  Auth: Basic {base64(client_id:client_secret)}            │
└────────────────┬───────────────────────────────────────────┘
                 │
                 ▼
┌────────────────────────────────────────────────────────────┐
│         6. Receive & Store Tokens                          │
│  {access_token, refresh_token, expires_in}                │
└────────────────────────────────────────────────────────────┘
```

### PKCE Implementation

#### Code Verifier Generation

```go
func generateRandomString(length int) (string, error) {
    bytes := make([]byte, length)
    rand.Read(bytes)
    return base64.RawURLEncoding.EncodeToString(bytes)[:length], nil
}
```

Generates a 64-character random string for the code verifier.

#### Code Challenge Generation

```go
func generateCodeChallenge(verifier string) string {
    hash := sha256.Sum256([]byte(verifier))
    return base64.RawURLEncoding.EncodeToString(hash[:])
}
```

Creates SHA-256 hash of verifier, base64-URL encoded.

### OAuth State Management

```go
type TwitterOAuthState struct {
    State        string
    CodeVerifier string
    ExpiresAt    time.Time
}
```

**Security Features**:
- State stored with 10-minute expiration
- Used once and deleted after callback (prevents replay attacks)
- Validates against CSRF attacks
- Expired states automatically cleaned up

### Token Management

```go
type TwitterTokens struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token,omitempty"`
    ExpiresIn    int    `json:"expires_in,omitempty"`
    TokenType    string `json:"token_type,omitempty"`
}
```

#### Token Refresh Flow

```go
func (a *TwitterAdapter) refreshTwitterOAuthToken(ctx context.Context,
    refreshToken, clientID, clientSecret string) (string, error)
```

1. Checks if access token is expired
2. Uses refresh token to obtain new access token
3. Stores new tokens (preserves old refresh token if new one not provided)
4. Returns new access token

### API Operations

#### Posting to Twitter

```go
func (a *TwitterAdapter) PostToTwitter(ctx context.Context, content string) (string, error)
```

**Endpoint**: `POST https://api.x.com/2/tweets`

**Request Body**:
```json
{
  "text": "Your tweet content here"
}
```

**Headers**:
- `Authorization: Bearer {access_token}`
- `Content-Type: application/json`

**Response**:
```json
{
  "data": {
    "id": "1234567890"
  }
}
```

#### Credential Validation

```go
func (a *TwitterAdapter) CheckCredentials(ctx context.Context) (*entity.TwitterProfile, error)
```

**Endpoint**: `GET https://api.x.com/2/users/me?user.fields=name,username`

Returns authenticated user's profile information.

### OAuth Methods

#### 1. Initiate Authentication

```go
func (a *TwitterAdapter) InitiateTwitterAuth(ctx context.Context) (*entity.TwitterAuthURL, error)
```

**Returns**:
```go
type TwitterAuthURL struct {
    AuthorizationURL string  // Full OAuth URL to redirect user to
    State            string  // State parameter for CSRF protection
}
```

**OAuth Scopes Requested**:
- `tweet.read` - Read tweets
- `tweet.write` - Post tweets
- `users.read` - Read user profile
- `offline.access` - Get refresh token

#### 2. Handle Callback

```go
func (a *TwitterAdapter) HandleTwitterCallback(ctx context.Context,
    input entity.TwitterCallbackInput) error
```

**Input**:
```go
type TwitterCallbackInput struct {
    Code  string  // Authorization code from Twitter
    State string  // State parameter for validation
}
```

**Process**:
1. Validates state (prevents CSRF)
2. Checks state hasn't expired
3. Exchanges code for tokens using code_verifier
4. Stores tokens securely
5. Deletes state (prevents replay)

#### 3. Update Tokens

```go
func (a *TwitterAdapter) UpdateTokens(ctx context.Context, tokens entity.TwitterTokens) error
```

Manually updates stored OAuth tokens (useful for token migration or manual intervention).

### Security Features

1. **PKCE**: Protects against authorization code interception
2. **State Parameter**: Prevents CSRF attacks
3. **State Expiration**: 10-minute lifetime for OAuth state
4. **One-Time Use**: State deleted after successful callback
5. **Secure Storage**: Tokens stored with sensitive flag in configuration
6. **Basic Auth**: Client credentials sent via HTTP Basic Auth header
7. **Token Refresh**: Automatic refresh when access token expires

### Error Handling

The adapter provides detailed error messages for:
- Missing configuration (client ID/secret)
- OAuth flow errors (invalid state, expired state)
- API errors (with HTTP status and response body)
- Token exchange failures
- Network/timeout errors

### Key Features

- **Complete OAuth 2.0 PKCE Flow**: Full implementation of modern OAuth with security best practices
- **Automatic Token Refresh**: Transparently refreshes expired access tokens
- **State Management**: Secure CSRF protection with automatic cleanup
- **Twitter API v2**: Uses latest Twitter API version
- **Flexible Configuration**: Environment variables for deployment flexibility
- **Comprehensive Error Handling**: Detailed error messages for debugging
- **Timeout Protection**: 30-second timeout on all HTTP requests

---

## Authentication and API Usage

### Configuration Repository Pattern

Both adapters depend on a `ConfigurationRepository` for storing credentials and session data:

```go
type ConfigurationRepository interface {
    GetByKey(ctx context.Context, key string) (*entity.Configuration, error)
    GetByPrefix(ctx context.Context, prefix string, sensitive bool) ([]entity.Configuration, error)
    Upsert(ctx context.Context, key, value string, sensitive bool, timestamp time.Time) (entity.Configuration, error)
    Delete(ctx context.Context, key string) error
}
```

### Sensitive Data Handling

Both adapters mark sensitive data appropriately:

| Data Type | Sensitive | Storage Key |
|-----------|-----------|-------------|
| Bluesky password | Yes | `bluesky.password` |
| Bluesky session | Yes | `bluesky.session` |
| Twitter client secret | Yes | `twitter.client_secret` |
| Twitter tokens | Yes | `twitter.tokens` |
| Twitter OAuth state | Yes | `twitter.state.{state}` |
| Post IDs | No | `*.last_post_id` |

### HTTP Client Configuration

Both adapters use consistent HTTP client settings:

```go
client := &http.Client{Timeout: 30 * time.Second}
```

- **Timeout**: 30 seconds for all requests
- **Context**: All requests use context for cancellation
- **TLS**: HTTPS enforced for all API calls

### Error Response Handling

Both adapters follow a consistent pattern:

```go
if resp.StatusCode != http.StatusOK {
    body, _ := io.ReadAll(resp.Body)
    return "", fmt.Errorf("API error: %s - %s", resp.Status, string(body))
}
```

This provides detailed error information including:
- HTTP status code
- Status message
- Full response body from API

### API Rate Limiting Considerations

While not explicitly implemented in the adapters, users should be aware:

**Twitter**:
- OAuth endpoints: Limited requests per 15-minute window
- Tweet creation: 300 tweets per 3-hour window (user context)
- Consider implementing exponential backoff for rate limit errors

**Bluesky**:
- Session refresh recommended over frequent re-authentication
- Post creation limits vary by instance
- Session tokens valid for 24 hours (reduces auth requests)

### Best Practices

1. **Credential Validation**: Always call `CheckCredentials()` before attempting posts
2. **Error Handling**: Check for specific error types (auth errors, network errors, API errors)
3. **Token Refresh**: Let adapters handle token refresh automatically
4. **Timeout Handling**: Use context with timeout for operations
5. **Secure Storage**: Ensure configuration repository encrypts sensitive data
6. **State Cleanup**: Twitter state cleanup runs automatically, but can be manually triggered
7. **Session Reuse**: Bluesky sessions cached for 24 hours, reuse when possible

### Testing Credentials

Both adapters provide credential checking methods:

```go
// Check Bluesky
profile, err := service.CheckBlueskyCredentials(ctx)
if err != nil {
    // Credentials invalid or network error
}
// profile.Handle and profile.DID available

// Check Twitter
profile, err := service.CheckTwitterCredentials(ctx)
if err != nil {
    // Credentials invalid, need OAuth, or network error
}
// profile.Username and profile.DisplayName available
```

### Example Integration

```go
// Initialize service
configRepo := // your configuration repository
socialService := social.NewService(configRepo)

// Post to both platforms
twitterID, err := socialService.PostToTwitter(ctx, "Hello from Garden!")
if err != nil {
    log.Printf("Twitter post failed: %v", err)
}

blueskyID, err := socialService.PostToBluesky(ctx, "Hello from Garden!")
if err != nil {
    log.Printf("Bluesky post failed: %v", err)
}

// Check credentials
twitterProfile, err := socialService.CheckTwitterCredentials(ctx)
blueskyProfile, err := socialService.CheckBlueskyCredentials(ctx)
```

---

## Summary

The social media adapter system provides a robust, secure, and maintainable solution for multi-platform social media integration:

- **Clean Architecture**: Follows hexagonal architecture with clear port/adapter separation
- **Platform-Specific Logic**: Each adapter handles its platform's unique requirements
- **Security First**: Implements OAuth 2.0 PKCE, session management, and secure token storage
- **Error Recovery**: Automatic session refresh and retry logic
- **Extensible**: Easy to add new platforms without modifying existing code
- **Testable**: Interface-based design enables easy mocking and testing

The system successfully abstracts the complexity of different social media APIs behind a unified interface, making it easy for application code to post content without worrying about platform-specific authentication flows or API details.
