# HTTP Server Documentation

This document provides comprehensive documentation for the HTTP server implementation in the Garden project. The server is built using the [Chi router](https://github.com/go-chi/chi) and follows RESTful API design patterns.

## Table of Contents

1. [Server Configuration and Setup](#server-configuration-and-setup)
2. [Router Configuration](#router-configuration)
3. [Middleware Chain](#middleware-chain)
4. [Response Helpers and Patterns](#response-helpers-and-patterns)
5. [Cross-Cutting Concerns](#cross-cutting-concerns)

---

## Server Configuration and Setup

### Server Structure

The HTTP server is defined in `/internal/adapter/primary/http/server.go` and consists of a simple struct that wraps the Chi router and HTTP server:

```go
type Server struct {
    router *chi.Mux
    server *http.Server
}
```

### Creating a New Server

The `NewServer()` function initializes a new server instance with a pre-configured middleware chain:

```go
func NewServer() *Server {
    r := chi.NewRouter()

    // Middleware
    r.Use(middleware.RequestID)
    r.Use(middleware.RealIP)
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(middleware.Timeout(60 * time.Second))
    r.Use(corsMiddleware)

    return &Server{
        router: r,
    }
}
```

### Starting the Server

The `Start(port string)` method configures and starts the HTTP server with the following features:

#### Port Configuration

The server port is determined in the following order of precedence:
1. Port passed as a parameter to `Start()`
2. `PORT` environment variable
3. Default port `8080`

#### HTTP Server Timeouts

The server is configured with conservative timeout values to prevent resource exhaustion:

```go
s.server = &http.Server{
    Addr:         ":" + port,
    Handler:      s.router,
    ReadTimeout:  15 * time.Second,  // Maximum time to read request
    WriteTimeout: 15 * time.Second,  // Maximum time to write response
    IdleTimeout:  60 * time.Second,  // Maximum idle time for keep-alive
}
```

| Timeout | Duration | Purpose |
|---------|----------|---------|
| `ReadTimeout` | 15 seconds | Prevents slow client attacks by limiting request read time |
| `WriteTimeout` | 15 seconds | Ensures responses are written in a timely manner |
| `IdleTimeout` | 60 seconds | Closes idle keep-alive connections |

#### Graceful Shutdown

The server implements graceful shutdown to ensure in-flight requests complete before termination:

```go
// Listen for interrupt signals
stop := make(chan os.Signal, 1)
signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

// Start server in goroutine
go func() {
    log.Printf("Server starting on port %s", port)
    if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatalf("Server error: %v", err)
    }
}()

// Wait for interrupt signal
<-stop
log.Println("Shutting down server...")

// Graceful shutdown with 10-second timeout
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

if err := s.server.Shutdown(ctx); err != nil {
    return fmt.Errorf("server shutdown error: %w", err)
}
```

**Shutdown Process:**
1. Server listens for `SIGINT` (Ctrl+C) or `SIGTERM` signals
2. Upon receiving signal, initiates graceful shutdown
3. Stops accepting new connections
4. Waits up to 10 seconds for active requests to complete
5. Returns error if shutdown exceeds timeout

---

## Router Configuration

### Chi Router

The server uses [Chi v5](https://github.com/go-chi/chi), a lightweight, idiomatic router for building Go HTTP services.

#### Accessing the Router

The router is accessible via the `Router()` method, allowing external code to register routes:

```go
func (s *Server) Router() *chi.Mux {
    return s.router
}
```

#### Usage Example

```go
server := NewServer()
router := server.Router()

// Register routes
router.Get("/health", healthHandler)
router.Post("/api/plants", createPlantHandler)
router.Get("/api/plants/{id}", getPlantHandler)
```

#### Router Features

Chi provides several features that make it ideal for API development:

- **HTTP method routing**: `GET`, `POST`, `PUT`, `DELETE`, etc.
- **URL parameters**: Path variables like `{id}` or `{slug}`
- **Subrouters**: Group related routes with shared middleware
- **Route groups**: Organize routes hierarchically
- **Standard library compatibility**: Works with `http.Handler` interface

---

## Middleware Chain

Middleware functions execute in order for every request. The server configures the following middleware chain:

### 1. RequestID Middleware

```go
r.Use(middleware.RequestID)
```

**Purpose**: Generates a unique ID for each request and adds it to the request context.

**Benefits**:
- Request tracing across services
- Log correlation
- Debugging distributed systems

**Header**: Adds `X-Request-ID` header to responses.

### 2. RealIP Middleware

```go
r.Use(middleware.RealIP)
```

**Purpose**: Extracts the client's real IP address from headers set by reverse proxies.

**Checks** (in order):
1. `X-Real-IP` header
2. `X-Forwarded-For` header
3. Direct remote address

**Use Case**: Essential when running behind load balancers or reverse proxies (nginx, HAProxy, CloudFlare, etc.).

### 3. Logger Middleware

```go
r.Use(middleware.Logger)
```

**Purpose**: Logs HTTP requests in a structured format.

**Logged Information**:
- Request ID
- HTTP method
- Request path
- Remote IP
- Response status code
- Response time
- Response size

**Example Output**:
```
INFO: "GET /api/plants/123 HTTP/1.1" from 192.168.1.100 - 200 1.234ms
```

### 4. Recoverer Middleware

```go
r.Use(middleware.Recoverer)
```

**Purpose**: Recovers from panics and prevents the server from crashing.

**Behavior**:
- Catches panics in handlers
- Logs the panic and stack trace
- Returns `500 Internal Server Error` to client
- Prevents entire server shutdown

**Critical for Production**: Ensures one bad request doesn't crash the entire server.

### 5. Timeout Middleware

```go
r.Use(middleware.Timeout(60 * time.Second))
```

**Purpose**: Enforces a maximum execution time for request handlers.

**Configuration**: 60-second timeout for all requests.

**Behavior**:
- Cancels request context after timeout
- Returns `504 Gateway Timeout` to client
- Prevents long-running requests from blocking resources

**Note**: Handlers should check `context.Done()` to respond to cancellation.

### 6. CORS Middleware

```go
r.Use(corsMiddleware)
```

**Purpose**: Handles Cross-Origin Resource Sharing (CORS) for browser-based clients.

See [CORS Configuration](#cors-configuration) section for details.

---

## Response Helpers and Patterns

The `/internal/adapter/primary/http/response.go` file provides helper functions for consistent HTTP responses.

### Error Response Structure

All error responses follow a consistent JSON structure:

```go
type ErrorResponse struct {
    Error   string `json:"error"`           // HTTP status text
    Message string `json:"message,omitempty"` // Detailed error message
}
```

**Example Error Response**:
```json
{
    "error": "Bad Request",
    "message": "Invalid plant ID format"
}
```

### JSON Response Helper

The primary response helper for sending JSON data:

```go
func JSON(w http.ResponseWriter, status int, data any)
```

**Usage**:
```go
plant := Plant{ID: "123", Name: "Monstera"}
http.JSON(w, http.StatusOK, plant)
```

**Behavior**:
1. Sets `Content-Type: application/json` header
2. Writes HTTP status code
3. Encodes data as JSON
4. Handles `nil` data gracefully

### Error Response Helpers

#### Generic Error Response

```go
func Error(w http.ResponseWriter, status int, err error)
```

**Usage**:
```go
http.Error(w, http.StatusUnauthorized, errors.New("Invalid API key"))
```

**Response**:
```json
{
    "error": "Unauthorized",
    "message": "Invalid API key"
}
```

#### NotFound Helper

```go
func NotFound(w http.ResponseWriter)
```

**Usage**:
```go
http.NotFound(w)
```

**Response**: `404 Not Found`
```json
{
    "error": "Not Found"
}
```

**Use Case**: Resource doesn't exist (e.g., plant ID not found in database).

#### BadRequest Helper

```go
func BadRequest(w http.ResponseWriter, err error)
```

**Usage**:
```go
http.BadRequest(w, errors.New("Missing required field: name"))
```

**Response**: `400 Bad Request`
```json
{
    "error": "Bad Request",
    "message": "Missing required field: name"
}
```

**Use Case**: Invalid input, validation errors, malformed JSON.

#### InternalError Helper

```go
func InternalError(w http.ResponseWriter, err error)
```

**Usage**:
```go
http.InternalError(w, errors.New("Database connection failed"))
```

**Response**: `500 Internal Server Error`
```json
{
    "error": "Internal Server Error",
    "message": "Database connection failed"
}
```

**Use Case**: Unexpected server errors, database failures, third-party service errors.

**Security Note**: Be cautious about exposing internal error details in production. Consider logging detailed errors server-side and returning generic messages to clients.

### Response Pattern Best Practices

#### Success Responses

```go
// Single resource
plant, err := service.GetPlant(id)
if err != nil {
    http.NotFound(w)
    return
}
http.JSON(w, http.StatusOK, plant)

// Collection
plants, err := service.ListPlants()
if err != nil {
    http.InternalError(w, err)
    return
}
http.JSON(w, http.StatusOK, plants)

// Created resource
newPlant, err := service.CreatePlant(data)
if err != nil {
    http.BadRequest(w, err)
    return
}
http.JSON(w, http.StatusCreated, newPlant)

// No content
err := service.DeletePlant(id)
if err != nil {
    http.InternalError(w, err)
    return
}
w.WriteHeader(http.StatusNoContent)
```

#### Error Handling Pattern

```go
func handler(w http.ResponseWriter, r *http.Request) {
    // Parse request
    var input Input
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.BadRequest(w, errors.New("Invalid JSON"))
        return
    }

    // Validate input
    if err := input.Validate(); err != nil {
        http.BadRequest(w, err)
        return
    }

    // Business logic
    result, err := service.DoSomething(input)
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            http.NotFound(w)
            return
        }
        http.InternalError(w, err)
        return
    }

    // Success response
    http.JSON(w, http.StatusOK, result)
}
```

---

## Cross-Cutting Concerns

### CORS Configuration

The `corsMiddleware` function handles Cross-Origin Resource Sharing for browser-based clients:

```go
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-Request-ID")

        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

#### CORS Headers

| Header | Value | Purpose |
|--------|-------|---------|
| `Access-Control-Allow-Origin` | `*` | Allows requests from any origin |
| `Access-Control-Allow-Methods` | `GET, POST, PUT, DELETE, OPTIONS` | Permitted HTTP methods |
| `Access-Control-Allow-Headers` | `Accept, Authorization, Content-Type, X-Request-ID` | Permitted request headers |

#### Preflight Requests

Browsers send `OPTIONS` requests (preflight) before actual requests to check CORS permissions. The middleware handles these by:
1. Setting CORS headers
2. Returning `200 OK` immediately
3. Not executing the handler chain

#### Security Considerations

**Current Configuration**: `Access-Control-Allow-Origin: *` allows **all origins**.

**For Production**, consider:
- Restricting to specific origins: `https://yourdomain.com`
- Using environment variables for allowed origins
- Implementing origin validation logic
- Adding `Access-Control-Allow-Credentials` if using cookies

**Recommended Update**:
```go
allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
if allowedOrigin == "" {
    allowedOrigin = "*"
}
w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
```

### Logging

#### Request Logging

The Chi `Logger` middleware automatically logs:
- Every incoming request
- Request method and path
- Client IP address
- Response status code
- Request duration
- Response size

#### Log Format

```
[timestamp] "METHOD /path HTTP/version" from IP - STATUS duration
```

**Example**:
```
2025-12-29T10:15:30Z "GET /api/plants/123 HTTP/1.1" from 192.168.1.100 - 200 12.5ms
```

#### Custom Logging

For application-level logging, use Go's standard `log` package or structured logging libraries like:
- [zerolog](https://github.com/rs/zerolog)
- [zap](https://github.com/uber-go/zap)
- [logrus](https://github.com/sirupsen/logrus)

### Error Recovery

The `Recoverer` middleware prevents panics from crashing the server:

**Without Recoverer**:
- Handler panic → Server crash → All requests fail

**With Recoverer**:
- Handler panic → Recovery → 500 error to client → Server continues

**Logging**: Panics are logged with stack traces for debugging.

### Request Timeout

The 60-second timeout prevents resource exhaustion from:
- Slow database queries
- Unresponsive external APIs
- Client connection issues
- Infinite loops (if context-aware)

**Context-Aware Handlers**:
```go
func handler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    select {
    case result := <-doWork(ctx):
        http.JSON(w, http.StatusOK, result)
    case <-ctx.Done():
        http.Error(w, http.StatusGatewayTimeout, errors.New("Request timeout"))
    }
}
```

### Security Considerations

#### Current Implementation

The server includes basic security measures:
- Request timeouts prevent slowloris attacks
- Panic recovery prevents crash exploits
- Real IP detection for accurate rate limiting
- CORS headers for browser security

#### Additional Recommendations

For production deployments, consider adding:

1. **Rate Limiting**: Prevent abuse and DoS attacks
   ```go
   import "github.com/go-chi/httprate"
   r.Use(httprate.LimitByIP(100, 1*time.Minute))
   ```

2. **Security Headers**: Add security headers middleware
   ```go
   w.Header().Set("X-Content-Type-Options", "nosniff")
   w.Header().Set("X-Frame-Options", "DENY")
   w.Header().Set("X-XSS-Protection", "1; mode=block")
   ```

3. **Request Size Limits**: Prevent large payload attacks
   ```go
   r.Body = http.MaxBytesReader(w, r.Body, 1048576) // 1MB limit
   ```

4. **TLS/HTTPS**: Always use HTTPS in production
   ```go
   s.server.ListenAndServeTLS("cert.pem", "key.pem")
   ```

5. **Authentication/Authorization**: Add auth middleware
   ```go
   r.Use(authMiddleware)
   ```

---

## Summary

The HTTP server implementation provides a robust foundation for building RESTful APIs in Go:

- **Chi Router**: Lightweight, idiomatic, standard library compatible
- **Comprehensive Middleware**: Request ID, logging, recovery, timeout, CORS
- **Graceful Shutdown**: Ensures clean server termination
- **Response Helpers**: Consistent JSON responses and error handling
- **Production-Ready Timeouts**: Prevents resource exhaustion
- **Extensible Design**: Easy to add routes and middleware

### Quick Start

```go
// Create and configure server
server := http.NewServer()
router := server.Router()

// Register routes
router.Get("/health", healthHandler)
router.Post("/api/plants", createPlantHandler)

// Start server (blocks until shutdown)
if err := server.Start("8080"); err != nil {
    log.Fatal(err)
}
```

### File Locations

- **Server Implementation**: `/internal/adapter/primary/http/server.go`
- **Response Helpers**: `/internal/adapter/primary/http/response.go`

---

*Last Updated: 2025-12-29*
