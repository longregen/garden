# Command-Line Applications

This document describes the command-line applications available in the Garden project. The project provides two separate server applications with different purposes and capabilities.

## Table of Contents

- [Overview](#overview)
- [API Server (`cmd/api`)](#api-server-cmdapi)
- [Main Server (`cmd/server`)](#main-server-cmdserver)
- [Environment Variables](#environment-variables)
- [Building and Running](#building-and-running)

---

## Overview

The Garden project provides two HTTP server applications:

1. **API Server** (`cmd/api/main.go`) - A minimal API server focused on configuration management
2. **Main Server** (`cmd/server/main.go`) - A full-featured application server with comprehensive functionality

Both servers use:
- PostgreSQL as the primary database
- Chi router for HTTP routing
- Hexagonal architecture (ports and adapters pattern)
- Graceful shutdown capabilities

---

## API Server (`cmd/api`)

### Purpose

The API server is a lightweight HTTP server that provides a minimal API focused on configuration management. It's designed for scenarios where only configuration-related endpoints are needed.

### Features

- Configuration management endpoints
- PostgreSQL database connectivity
- HTTP server with Chi router
- Port: 8080 (hardcoded)

### Architecture

The API server initializes the following components:

1. **Database Connection**: Connects to PostgreSQL using `postgres.NewDB()`
2. **Repository Layer**:
   - `ConfigurationRepository` - Handles configuration data persistence
3. **Service Layer**:
   - `ConfigurationService` - Business logic for configuration management
4. **HTTP Layer**:
   - `ConfigurationHandler` - HTTP handlers for configuration endpoints
   - Chi router for request routing

### Source Code Location

```
/home/user/garden/cmd/api/main.go
```

---

## Main Server (`cmd/server`)

### Purpose

The main server is a comprehensive application server that provides a full suite of functionality including knowledge management, bookmarking, social media integration, AI-powered search, and more.

### Features

- **Knowledge Management**: Notes, items, entities, and categories
- **Communication**: Contacts, rooms, messages, and sessions
- **Bookmarks**: Web page bookmarking with content extraction and AI summaries
- **Social Media**: Social post tracking and management
- **AI Integration**:
  - Semantic search using vector embeddings
  - LLM-powered search and summarization
  - Content processing and analysis
- **Browser Integration**: Browser history tracking
- **Logseq Sync**: Integration with Logseq knowledge base
- **Dashboard**: Analytics and insights
- **Observations**: Data observation and tracking
- **Tags**: Cross-cutting tagging system
- **Health Monitoring**: `/health` endpoint for service health checks
- **Graceful Shutdown**: Proper signal handling for clean server shutdown

### Architecture

The main server follows a clean hexagonal architecture with clear separation of concerns:

#### Primary Adapters (Inbound)
- HTTP handlers for all domain entities
- REST API endpoints
- Chi router with middleware (request ID, logging, CORS, recovery, timeouts)

#### Secondary Adapters (Outbound)
- **PostgreSQL**: Data persistence for all repositories
- **Ollama**: LLM and embedding services
- **AI Service**: Content summarization and analysis
- **HTTP Fetcher**: Web content retrieval
- **Social Media Services**: Social platform integrations
- **Content Processor**: Web page content extraction and processing

#### Domain Services

The server initializes 15+ domain services:
- Configuration, Contact, Room, Message, Session
- Note, Item, Bookmark, Entity, Category
- Social Post, Observation, Dashboard
- Browser History, Search, Utility
- Logseq Sync, Tag

#### Repositories

Corresponding repositories for each domain entity, all backed by PostgreSQL.

### Endpoints

The server registers routes for all domain entities:
- Configuration management
- Contact and communication (rooms, messages, sessions)
- Knowledge management (notes, items, entities, categories)
- Bookmarks and browser history
- Social posts
- Search functionality
- Dashboard and observations
- Logseq synchronization
- Tag management
- Health check at `/health`

### Graceful Shutdown

The server implements graceful shutdown:
- Listens for `SIGINT` and `SIGTERM` signals
- Allows up to 30 seconds for in-flight requests to complete
- Cleanly closes database connections
- Logs shutdown progress

### Source Code Location

```
/home/user/garden/cmd/server/main.go
```

---

## Environment Variables

### Database Configuration

Both servers require database configuration:

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://gardener@localhost:5432/garden?sslmode=disable` | No |

**Example:**
```bash
DATABASE_URL="postgres://user:password@localhost:5432/garden?sslmode=disable"
```

### Server Configuration

| Variable | Description | Default | Used By |
|----------|-------------|---------|---------|
| `PORT` | HTTP server port | `8080` | Main Server only |

**Note:** The API server uses hardcoded port `8080`.

### AI and LLM Services (Main Server Only)

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `OLLAMA_API_URL` | Ollama API endpoint | None | Yes (for LLM features) |
| `OLLAMA_MODEL` | Ollama model name | None | Yes (for LLM features) |
| `OLLAMA_EMBED_API_URL` | Ollama embeddings API endpoint | Falls back to `OLLAMA_API_URL` | No |
| `OLLAMA_EMBED_MODEL` | Ollama embeddings model | `nomic-embed-text:latest` | No |
| `AI_SERVICE_URL` | AI service endpoint for summaries | Falls back to `OLLAMA_API_URL` | No |
| `AI_SERVICE_KEY` | AI service API key | None | No |

**Example:**
```bash
OLLAMA_API_URL="http://localhost:11434"
OLLAMA_MODEL="llama2"
OLLAMA_EMBED_MODEL="nomic-embed-text:latest"
AI_SERVICE_URL="http://localhost:11434"
AI_SERVICE_KEY=""
```

---

## Building and Running

### Prerequisites

- Go 1.24 or later
- PostgreSQL database
- (Optional) Ollama for AI features (main server)

### Building

Build both applications from the project root:

```bash
# Build API server
go build -o bin/api ./cmd/api

# Build main server
go build -o bin/server ./cmd/server

# Build both
go build -o bin/api ./cmd/api && go build -o bin/server ./cmd/server
```

### Running

#### API Server

**Using `go run`:**
```bash
cd /home/user/garden
go run ./cmd/api
```

**Using compiled binary:**
```bash
./bin/api
```

**With custom database:**
```bash
DATABASE_URL="postgres://user:pass@localhost:5432/mydb?sslmode=disable" go run ./cmd/api
```

The API server will start on `http://localhost:8080`.

#### Main Server

**Using `go run`:**
```bash
cd /home/user/garden
go run ./cmd/server
```

**Using compiled binary:**
```bash
./bin/server
```

**With environment variables:**
```bash
export DATABASE_URL="postgres://gardener@localhost:5432/garden?sslmode=disable"
export PORT="8080"
export OLLAMA_API_URL="http://localhost:11434"
export OLLAMA_MODEL="llama2"
export OLLAMA_EMBED_MODEL="nomic-embed-text:latest"

go run ./cmd/server
```

**Using `.env` file:**
```bash
# Create a .env file
cat > .env << EOF
DATABASE_URL=postgres://gardener@localhost:5432/garden?sslmode=disable
PORT=8080
OLLAMA_API_URL=http://localhost:11434
OLLAMA_MODEL=llama2
OLLAMA_EMBED_MODEL=nomic-embed-text:latest
AI_SERVICE_URL=http://localhost:11434
EOF

# Load and run
set -a; source .env; set +a
go run ./cmd/server
```

The main server will start on `http://localhost:8080` (or the port specified in `PORT`).

### Verifying the Server

Check if the main server is running:

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{"status":"ok"}
```

### Database Setup

Both servers expect a PostgreSQL database to be available. Ensure you have:

1. PostgreSQL installed and running
2. A database created (default: `garden`)
3. A user with access (default: `gardener`)
4. Schema migrations applied (see `schema.sql` in project root)

**Example database setup:**
```bash
# Create database and user
psql -U postgres << EOF
CREATE DATABASE garden;
CREATE USER gardener;
GRANT ALL PRIVILEGES ON DATABASE garden TO gardener;
EOF

# Apply schema
psql -U gardener -d garden -f /home/user/garden/schema.sql
```

### Stopping the Server

Both servers support graceful shutdown:

- Press `Ctrl+C` (sends `SIGINT`)
- Send `SIGTERM` signal: `kill <pid>`

The servers will:
1. Stop accepting new connections
2. Wait for existing requests to complete (up to 30 seconds for main server)
3. Close database connections
4. Exit cleanly

---

## Command-Line Flags

**Note:** Neither the API server nor the main server currently accepts command-line flags. All configuration is done through environment variables.

---

## Development

### Project Structure

```
/home/user/garden/
├── cmd/
│   ├── api/          # Minimal API server
│   │   └── main.go
│   └── server/       # Full-featured server
│       └── main.go
├── internal/
│   ├── adapter/      # Hexagonal architecture adapters
│   │   ├── primary/  # Inbound adapters (HTTP)
│   │   └── secondary/# Outbound adapters (DB, AI, etc.)
│   └── domain/       # Business logic and entities
│       └── service/  # Domain services
├── docs/             # Documentation
├── go.mod            # Go module definition
└── schema.sql        # Database schema
```

### Module Name

```
garden3
```

### Key Dependencies

- **github.com/go-chi/chi/v5** - HTTP router
- **github.com/jackc/pgx/v5** - PostgreSQL driver
- **github.com/pgvector/pgvector-go** - PostgreSQL vector extension
- **github.com/go-shiori/go-readability** - Content extraction
- **github.com/google/uuid** - UUID generation

### Adding New Features

To add new features to the servers:

1. **Domain Layer**: Create domain entities and services in `internal/domain/`
2. **Repository**: Implement repository in `internal/adapter/secondary/postgres/repository/`
3. **Handler**: Create HTTP handlers in `internal/adapter/primary/http/handler/`
4. **Wire Up**: Update the appropriate `main.go` to initialize and register the new components

### Testing

The servers can be tested locally by:

1. Starting a PostgreSQL instance
2. Running database migrations
3. Starting the server with appropriate environment variables
4. Making HTTP requests to the endpoints

Example:
```bash
# Start the main server
go run ./cmd/server

# Test health endpoint
curl http://localhost:8080/health

# Test configuration endpoints
curl http://localhost:8080/configurations
```

---

## Troubleshooting

### Database Connection Issues

**Error:** `Failed to connect to database`

**Solutions:**
- Verify PostgreSQL is running: `pg_isready`
- Check `DATABASE_URL` is correct
- Ensure database exists and user has permissions
- Check network connectivity to database host

### Port Already in Use

**Error:** `bind: address already in use`

**Solutions:**
- Change `PORT` environment variable (main server only)
- Stop other services using port 8080
- Find and kill the process: `lsof -ti:8080 | xargs kill`

### Ollama Connection Issues (Main Server)

**Error:** Failed to connect to Ollama or AI service

**Solutions:**
- Verify Ollama is running: `curl http://localhost:11434/api/tags`
- Check `OLLAMA_API_URL` is correct
- Ensure required models are pulled: `ollama pull llama2` and `ollama pull nomic-embed-text`
- Verify network connectivity to Ollama host

### Missing Environment Variables

**Issue:** Server starts but features don't work

**Solution:**
- Ensure all required environment variables are set
- Check logs for warnings about missing configuration
- Use `.env` file for easier environment management

---

## Additional Resources

- **Database Schema**: `/home/user/garden/schema.sql`
- **Go Module**: `/home/user/garden/go.mod`
- **Internal Documentation**: See code comments in respective packages

For more information about specific domains and handlers, refer to the source code in `/home/user/garden/internal/`.
