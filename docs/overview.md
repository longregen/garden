# Garden3 - Personal Knowledge Management System

## Table of Contents

- [Introduction](#introduction)
- [Architecture Overview](#architecture-overview)
- [Technology Stack](#technology-stack)
- [Directory Structure](#directory-structure)
- [Core Features](#core-features)
- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [Development](#development)

## Introduction

Garden3 is a comprehensive personal knowledge management system designed to help users organize, search, and interact with various forms of information. Built with modern Go practices and clean architecture principles, Garden3 provides a robust platform for managing:

- **Notes and Documentation**: Create and organize markdown-based notes with tagging and entity references
- **Bookmarks**: Save and analyze web content with automatic summarization and semantic search
- **Browser History**: Track and search through browsing history
- **Knowledge Graph**: Build relationships between entities and concepts
- **Chat Sessions**: Interactive AI-powered conversations with context persistence
- **Social Media**: Track and analyze social media posts
- **Observations**: Record personal observations and insights
- **Logseq Integration**: Sync and integrate with Logseq knowledge base

The system leverages modern AI capabilities including vector embeddings for semantic search, LLM integration for summarization and chat, and content processing for extracting meaningful information from web pages.

## Architecture Overview

Garden3 follows **Hexagonal Architecture** (also known as Ports and Adapters), which provides a clean separation of concerns and makes the codebase highly maintainable and testable.

### Architectural Layers

```
┌─────────────────────────────────────────────────────────────┐
│                    Primary Adapters                          │
│                  (HTTP Handlers/API)                         │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│                    Input Ports                               │
│              (Use Case Interfaces)                           │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│                  Domain Layer                                │
│                                                              │
│  ┌────────────┐  ┌────────────┐  ┌──────────────┐          │
│  │  Entities  │  │  Services  │  │ Value Objects│          │
│  └────────────┘  └────────────┘  └──────────────┘          │
│                                                              │
│  Business Logic & Domain Rules                              │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│                   Output Ports                               │
│         (Repository & Service Interfaces)                    │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│                 Secondary Adapters                           │
│                                                              │
│  ┌──────────┐ ┌──────────┐ ┌────────┐ ┌──────────┐        │
│  │PostgreSQL│ │  Ollama  │ │ HTTP   │ │  Social  │        │
│  │          │ │Embeddings│ │Fetcher │ │  Media   │        │
│  └──────────┘ └──────────┘ └────────┘ └──────────┘        │
└─────────────────────────────────────────────────────────────┘
```

### Key Architectural Benefits

1. **Dependency Inversion**: The domain layer has no dependencies on external frameworks or libraries
2. **Testability**: Each layer can be tested independently with mocks/stubs
3. **Flexibility**: Easy to swap implementations (e.g., change database or AI service)
4. **Clear Boundaries**: Well-defined interfaces between layers prevent coupling

### Layer Responsibilities

#### Domain Layer (`internal/domain/`)
- **Entities**: Core business objects (Note, Bookmark, Entity, etc.)
- **Services**: Business logic and use case implementations
- **Value Objects**: Immutable objects representing domain concepts

#### Ports (`internal/port/`)
- **Input Ports**: Define use cases that the application provides
- **Output Ports**: Define interfaces for external dependencies (repositories, services)

#### Adapters (`internal/adapter/`)
- **Primary Adapters**: Entry points to the application (HTTP handlers)
- **Secondary Adapters**: Implementations of output ports (PostgreSQL repos, AI services)

## Technology Stack

### Core Technologies

| Technology | Version | Purpose |
|------------|---------|---------|
| **Go** | 1.25.5 | Primary programming language |
| **PostgreSQL** | 17.7+ | Primary data store with advanced extensions |
| **Chi Router** | v5.2.3 | HTTP routing and middleware |
| **pgx** | v5.8.0 | PostgreSQL driver and toolkit |

### PostgreSQL Extensions

The system relies on several PostgreSQL extensions for advanced functionality:

- **pgvector**: Vector similarity search for semantic embeddings
- **pg_trgm**: Trigram-based text similarity and fuzzy search
- **fuzzystrmatch**: Fuzzy string matching (Levenshtein distance, etc.)
- **unaccent**: Remove accents from text for better search
- **uuid-ossp**: Generate UUIDs for primary keys

### External Services

| Service | Library | Purpose |
|---------|---------|---------|
| **Ollama** | Custom HTTP client | LLM inference and text embeddings |
| **Embeddings** | pgvector-go v0.3.0 | Vector embeddings for semantic search |
| **Content Processing** | go-readability | Extract readable content from web pages |
| **HTTP Fetcher** | net/http | Fetch web content for bookmarks |

### Key Dependencies

```go
// Core HTTP & Routing
github.com/go-chi/chi/v5 v5.2.3

// Database
github.com/jackc/pgx/v5 v5.8.0
github.com/pgvector/pgvector-go v0.3.0

// Content Processing
github.com/go-shiori/go-readability v0.0.0-20251205110129-5db1dc9836f0
github.com/JohannesKaufmann/html-to-markdown/v2 v2.5.0

// Utilities
github.com/google/uuid v1.6.0
gopkg.in/yaml.v3 v3.0.1
```

## Directory Structure

```
/home/user/garden/
├── cmd/                          # Application entry points
│   ├── api/                      # API-specific entry point
│   └── server/                   # Main server application
│       └── main.go              # Server initialization and wiring
│
├── internal/                     # Private application code
│   ├── adapter/                  # Adapters (implementations)
│   │   ├── primary/             # Incoming adapters
│   │   │   └── http/            # HTTP handlers and routing
│   │   │       └── handler/     # Domain-specific HTTP handlers
│   │   └── secondary/           # Outgoing adapters
│   │       ├── ai/              # AI service integration
│   │       ├── contentprocessor/ # Content processing utilities
│   │       ├── embedding/       # Embedding service (Ollama)
│   │       ├── httpfetch/       # HTTP fetching service
│   │       ├── llm/             # LLM service integration
│   │       ├── postgres/        # PostgreSQL connection
│   │       │   └── repository/  # Repository implementations
│   │       └── social/          # Social media integrations
│   │
│   ├── domain/                   # Core business logic
│   │   ├── entity/              # Domain entities (19 entities)
│   │   ├── service/             # Business logic services
│   │   └── valueobject/         # Value objects
│   │
│   └── port/                     # Interfaces (contracts)
│       ├── input/               # Use case interfaces (18 use cases)
│       └── output/              # Repository & service interfaces
│
├── docs/                         # Documentation
│   └── overview.md              # This file
│
├── go.mod                        # Go module definition
├── go.sum                        # Go dependency checksums
├── schema.sql                    # Database schema (156KB)
└── sqlc.yaml                     # SQL code generation config
```

### Entity Overview

The system manages 19 core domain entities:

1. **bookmark.go** - Web bookmarks with content and metadata
2. **browser_history.go** - Browser browsing history
3. **category.go** - Categorization system
4. **configuration.go** - System and user configuration
5. **contact.go** - Contact management
6. **dashboard.go** - Dashboard data and metrics
7. **entity.go** - Knowledge graph entities
8. **entity_reference.go** - References between entities
9. **item.go** - Generic items in the system
10. **logseq.go** - Logseq integration entities
11. **message.go** - Chat messages
12. **note.go** - Notes and documentation
13. **observation.go** - Personal observations
14. **room.go** - Chat rooms/conversations
15. **search.go** - Search functionality entities
16. **session.go** - Chat sessions
17. **social_post.go** - Social media posts
18. **tag.go** - Tagging system
19. **utility.go** - Utility entities

## Core Features

### 1. Semantic Search

Garden3 uses vector embeddings to enable semantic search across all content types:

- **Vector Storage**: PostgreSQL with pgvector extension
- **Embedding Model**: Configurable Ollama model (default: nomic-embed-text:latest)
- **Search Types**:
  - Similarity search across bookmarks
  - Note content search
  - Session context retrieval

### 2. Bookmark Management

Comprehensive bookmark system with:

- **Content Fetching**: Automatic HTTP fetching of bookmark content
- **Content Processing**: Extract readable content using go-readability
- **Summarization**: AI-generated summaries of bookmark content
- **Q&A**: Store questions and answers about bookmarks
- **Categorization**: Organize bookmarks by category
- **Full-Text Search**: Trigram-based fuzzy search

### 3. Knowledge Graph

Entity-based knowledge management:

- **Entity Types**: Flexible entity typing system
- **Relationships**: Track connections between entities
- **Entity References**: Link notes and other content to entities
- **Logseq Integration**: Sync entities with Logseq

### 4. AI-Powered Features

- **Chat Sessions**: Conversational AI with context persistence
- **Summarization**: Automatic content summarization
- **Question Generation**: Generate Q&A from content
- **Semantic Search**: Find related content using embeddings

### 5. Content Aggregation

- **Browser History**: Import and search browser history
- **Social Media**: Track social media posts
- **Notes**: Markdown-based note-taking
- **Observations**: Record daily observations

## Getting Started

### Prerequisites

1. **Go 1.24 or higher**
   ```bash
   go version
   ```

2. **PostgreSQL 17.7+ with extensions**
   ```bash
   psql --version
   ```

3. **Ollama** (for AI features)
   ```bash
   ollama --version
   ```

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd /home/user/garden
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up the database**

   Create a PostgreSQL database and apply the schema:
   ```bash
   createdb garden3
   psql garden3 < schema.sql
   ```

4. **Configure environment variables**

   Create a `.env` file or export the following variables:
   ```bash
   # Database
   export DATABASE_URL="postgres://user:password@localhost:5432/garden3"

   # Ollama Configuration
   export OLLAMA_API_URL="http://localhost:11434"
   export OLLAMA_MODEL="llama2:latest"
   export OLLAMA_EMBED_MODEL="nomic-embed-text:latest"

   # Optional: Separate embedding service
   export OLLAMA_EMBED_API_URL="http://localhost:11434"

   # Optional: AI service for summarization
   export AI_SERVICE_URL="http://localhost:11434"
   export AI_SERVICE_KEY=""
   ```

5. **Install Ollama models**
   ```bash
   ollama pull llama2:latest
   ollama pull nomic-embed-text:latest
   ```

### Running the Application

1. **Start the server**
   ```bash
   cd /home/user/garden
   go run cmd/server/main.go
   ```

2. **Verify the server is running**
   ```bash
   curl http://localhost:8080/health
   ```

   Expected response:
   ```json
   {"status": "ok"}
   ```

### API Endpoints

The server exposes RESTful endpoints for all entities. Examples:

```bash
# Bookmarks
GET    /api/bookmarks
POST   /api/bookmarks
GET    /api/bookmarks/:id
PUT    /api/bookmarks/:id
DELETE /api/bookmarks/:id

# Notes
GET    /api/notes
POST   /api/notes
GET    /api/notes/:id
PUT    /api/notes/:id
DELETE /api/notes/:id

# Search
POST   /api/search
GET    /api/search/similar

# Chat Sessions
GET    /api/sessions
POST   /api/sessions
GET    /api/sessions/:id
POST   /api/sessions/:id/messages

# And many more...
```

## Configuration

### Database Connection

The application uses the `DATABASE_URL` environment variable or defaults to PostgreSQL connection parameters. The connection is managed through pgx connection pooling.

### Ollama Services

Garden3 integrates with Ollama for multiple AI capabilities:

1. **LLM Service**: For chat and text generation
   - `OLLAMA_API_URL`: Ollama server URL
   - `OLLAMA_MODEL`: Model for text generation

2. **Embedding Service**: For vector embeddings
   - `OLLAMA_EMBED_API_URL`: Embedding service URL (defaults to OLLAMA_API_URL)
   - `OLLAMA_EMBED_MODEL`: Embedding model (default: nomic-embed-text:latest)

3. **AI Service**: For summarization
   - `AI_SERVICE_URL`: AI service URL (defaults to OLLAMA_API_URL)
   - `AI_SERVICE_KEY`: Optional API key

### Server Configuration

- **Port**: Default 8080 (configured in `cmd/server/main.go`)
- **Graceful Shutdown**: 30-second timeout for in-flight requests
- **CORS**: Configured in HTTP adapter

## Development

### Project Structure Philosophy

Garden3 follows clean architecture principles:

1. **Domain-Driven Design**: Business logic is in the domain layer
2. **Dependency Rule**: Dependencies point inward toward the domain
3. **Interface Segregation**: Small, focused interfaces
4. **Testability**: All layers can be tested independently

### Adding a New Feature

To add a new feature (e.g., a new entity type):

1. **Define the entity** in `internal/domain/entity/`
2. **Create output port** interface in `internal/port/output/`
3. **Create input port** interface in `internal/port/input/`
4. **Implement domain service** in `internal/domain/service/`
5. **Implement repository** in `internal/adapter/secondary/postgres/repository/`
6. **Create HTTP handler** in `internal/adapter/primary/http/handler/`
7. **Register routes** in `cmd/server/main.go`
8. **Update database schema** in `schema.sql`

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for a specific package
go test ./internal/domain/service/...
```

### Code Generation

The project uses `sqlc` for type-safe SQL:

```bash
# Generate code from SQL
sqlc generate
```

Configuration is in `sqlc.yaml`.

### Building

```bash
# Build the server
go build -o bin/garden3-server cmd/server/main.go

# Build with optimizations
go build -ldflags="-s -w" -o bin/garden3-server cmd/server/main.go
```

### Database Migrations

Database schema is managed in `schema.sql`. To apply changes:

```bash
psql garden3 < schema.sql
```

For production, consider using a migration tool like `golang-migrate` or `goose`.

---

## Next Steps

- Explore individual entity documentation in `/home/user/garden/docs/entities/`
- Review API documentation for detailed endpoint information
- Check configuration options for customizing behavior
- Set up development environment for contributing

For questions or issues, please refer to the project repository or contact the maintainer.
