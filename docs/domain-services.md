# Domain Services Documentation

## Overview of Service Layer

The service layer in the Garden application implements the business logic and use cases of the domain. It sits between the HTTP handlers (input adapters) and repositories (output adapters), following a hexagonal/clean architecture pattern.

### Architecture Pattern

Services in this layer:
- **Implement use case interfaces** defined in `internal/port/input`
- **Depend on repository interfaces** defined in `internal/port/output`
- **Contain business logic** that orchestrates domain operations
- **Handle error wrapping** with contextual information
- **Maintain separation of concerns** between infrastructure and domain logic

### Common Patterns

1. **Constructor Pattern**: All services use a `New<ServiceName>Service()` constructor that accepts dependencies via dependency injection
2. **Error Wrapping**: Errors are wrapped with `fmt.Errorf("context: %w", err)` to provide call stack context
3. **Pagination**: Services returning lists typically return `PaginatedResponse[T]` with metadata
4. **Context-aware**: All service methods accept `context.Context` as the first parameter
5. **Repository delegation**: Services delegate persistence operations to repository interfaces

---

## Service Documentation

### 1. Bookmark Service

**Location**: `/home/user/garden/internal/domain/service/bookmark.go`

#### Responsibilities

The Bookmark Service manages web bookmarks, including:
- CRUD operations for bookmarks
- Content fetching and processing from URLs
- Embedding generation for semantic search
- Question/answer generation and management
- Similarity search across bookmarks

#### Dependencies

- `output.BookmarkRepository`: Persistence layer for bookmark data
- `output.HTTPFetcher`: HTTP client for fetching web content
- `output.EmbeddingsService`: AI service for generating embeddings
- `output.AIService`: AI service for content summarization
- `output.ContentProcessor`: Content extraction (Lynx, Reader)

#### Key Business Logic

**Content Processing Pipeline**:
1. **Fetch** → Retrieve raw HTML from URL with timeout (25 seconds)
2. **Process** → Extract readable content using Lynx or Reader strategies
3. **Embed** → Generate vector embeddings for semantic search
4. **Summarize** → Create AI-generated summaries with word count optimization

**Smart Title Extraction**:
- First attempts to use Reader-extracted title
- Falls back to HTML `<title>` tag parsing
- Caches extracted titles to avoid reprocessing

**Embedding Strategy**:
- Limits content to 15,000 characters to prevent token overflow
- Caps at 20 embedding chunks per bookmark
- Returns warnings when content truncation occurs

**Summary Generation Loop**:
```go
// Iteratively reduces word count if embedding splits
for wordCount > 200 {
    summary = generateSummary(content, wordCount)
    embeddings = getEmbedding(summary)
    if len(embeddings) == 1 { break }
    wordCount -= 10
}
```

#### Error Handling

- **URL Sanitization**: Basic URL cleaning with `sanitizeURL()`
- **Timeout Management**: 25-second timeout for HTTP fetches
- **Graceful Degradation**: Stores fetch errors in database for debugging
- **Content Type Validation**: Checks for HTML/text before processing
- **Observation Logging**: Creates audit trail for Q&A edits/deletes

#### Search Capabilities

- **Pagination**: Supports offset-based pagination with page/limit
- **Filtering**: By category, date range, search query
- **Semantic Search**: Vector similarity using embeddings with configurable strategy
- **Random Selection**: `GetRandomBookmark()` for discovery features

---

### 2. Browser History Service

**Location**: `/home/user/garden/internal/domain/service/browser_history.go`

#### Responsibilities

Manages browser history data imported from external sources:
- Paginated listing with filters
- Domain-based analytics
- Recent history queries

#### Dependencies

- `output.BrowserHistoryRepository`: Data access layer

#### Key Business Logic

**Pagination Calculation**:
- Converts page number to SQL offset: `offset = (page - 1) * pageSize`
- Calculates total pages using ceiling division
- Default page size: 10 items

**Domain Analytics**:
- Aggregates visits by domain
- Returns top N domains by visit count
- Default limit: 10 domains

#### Error Handling

- **Input Validation**: Normalizes page/pageSize to valid ranges (≥1)
- **Direct Error Propagation**: Returns repository errors without additional wrapping

---

### 3. Category Service

**Location**: `/home/user/garden/internal/domain/service/category.go`

#### Responsibilities

Manages hierarchical bookmark categories:
- CRUD operations for categories
- Category merging for consolidation
- Source URI management per category

#### Dependencies

- `output.CategoryRepository`: Persistence layer

#### Key Business Logic

**Category Merging**:
- Transfers all bookmarks from source to target category
- Deletes source category after successful transfer
- Atomic operation to prevent data loss

**Source Management**:
- Categories can have multiple source URIs
- Sources track raw source data for import traceability
- Updates preserve existing values for partial updates

#### Error Handling

- Wraps all repository errors with operation context
- Preserves existing field values during updates when input is nil
- Returns detailed error messages for debugging

---

### 4. Configuration Service

**Location**: `/home/user/garden/internal/domain/service/configuration.go`

#### Responsibilities

Application-wide configuration management:
- Key-value storage with type conversions
- Secret/non-secret value distinction
- Prefix-based querying
- Typed getters (bool, number, JSON)

#### Dependencies

- `output.ConfigurationRepository`: Configuration persistence

#### Key Business Logic

**Type Conversion Helpers**:
```go
GetBoolValue(key, default) → bool    // "true" → true
GetNumberValue(key, default) → float64
GetJSONValue(key) → []byte           // Validates JSON structure
```

**Secret Management**:
- `isSecret` flag controls visibility
- Secrets excluded from prefix queries with `includeSecrets=false`
- No automatic encryption (handled at infrastructure layer)

**Upsert Pattern**:
- `SetConfiguration()` creates or updates atomically
- Updates `updated_at` timestamp automatically

#### Error Handling

- **Required Validation**: Returns error if key or value is empty
- **JSON Validation**: Unmarshals to validate structure before storing
- **Graceful Defaults**: Type getters return default value on parse failure
- **Missing Key Handling**: Returns nil pointer for missing values

---

### 5. Contact Service

**Location**: `/home/user/garden/internal/domain/service/contact.go`

#### Responsibilities

Comprehensive contact management system:
- Contact CRUD with rich metadata
- Relationship evaluations (importance, closeness, fondness)
- Tag management for categorization
- Batch update operations
- Contact merging for deduplication
- Statistics refresh for message counts

#### Dependencies

- `output.ContactRepository`: Data access and relationships

#### Key Business Logic

**Full Contact Assembly**:
```go
GetContact(id) returns:
  - Base contact info
  - Evaluation metrics (nullable)
  - Tags array
  - Known names/aliases
  - Associated rooms
  - Source references
```

**Batch Update Processing**:
- Iterates through updates sequentially
- Continues on error (returns per-item results)
- Updates basic info, evaluations, and tags in separate operations
- Performs tag diffing to add/remove only changed tags

**Evaluation Management**:
- Auto-creates evaluation if it doesn't exist
- Updates existing evaluation if present
- Three metrics: importance, closeness, fondness

**Statistics Refresh**:
- Ensures all contacts have stats records
- Recalculates message counts from message table
- Two-phase operation (ensure, then update)

#### Error Handling

- Graceful evaluation handling (returns nil if not found)
- Batch operations collect errors but don't fail transaction
- Tag operations validate existence before modification
- Search falls back to simple list if query is empty

#### Search and Filtering

- **Text Search**: LIKE pattern matching on name/notes
- **Pagination**: Standard offset-based pagination
- **Batch Loading**: Loads evaluations and tags for all contacts in list queries

---

### 6. Dashboard Service

**Location**: `/home/user/garden/internal/domain/service/dashboard.go`

#### Responsibilities

Aggregates statistics across the application:
- Contact statistics
- Session/conversation statistics
- Bookmark statistics
- Browser history statistics
- Recent activity items

#### Dependencies

- `output.DashboardRepository`: Aggregation queries

#### Key Business Logic

**Composite Statistics**:
- Fetches multiple stat categories in parallel potential
- Assembles into single `DashboardStats` response
- Delegates all calculations to repository layer

#### Error Handling

- Returns error on first failure (no partial results)
- Simple delegation pattern with minimal logic

---

### 7. Entity Service

**Location**: `/home/user/garden/internal/domain/service/entity.go`

#### Responsibilities

Generic entity management for knowledge graph:
- Entity CRUD (people, organizations, concepts)
- Relationship management between entities
- Reference tracking from content
- Soft deletion support
- Wiki-style reference parsing `[[entity]]`

#### Dependencies

- `output.EntityRepository`: Entity and relationship persistence

#### Key Business Logic

**Reference Parsing**:
```regex
Pattern: \[\[(.*?)(?:\]\[([^\]]+))?\]\]
Matches:
  - [[EntityName]]           → simple reference
  - [[Display][EntityName]]  → reference with display text
```

**Cascading Name Updates**:
- When person entity name changes
- Automatically updates related contact names
- Uses relationship type "identity" to find contacts

**Relationship Filtering**:
- Can filter by `relatedType` (contact, bookmark, etc.)
- Can filter by `relationshipType` (identity, mention, etc.)
- Uses optimized query if both filters provided

#### Error Handling

- Wraps all errors with operation context
- Logs warnings for contact update failures (non-fatal)
- Validates filters before applying

---

### 8. Item Service

**Location**: `/home/user/garden/internal/domain/service/item.go`

#### Responsibilities

Manages generic items (notes/documents):
- Item CRUD operations
- Tag management
- Semantic search via embeddings

#### Dependencies

- `output.ItemRepository`: Item persistence and search

#### Key Business Logic

**Tag Handling During Creation**:
1. Upsert each tag name (creates if new)
2. Create item-tag relationships
3. Uses item creation timestamp for tag metadata

**Update Operations**:
- Preserves existing values for nil inputs
- Handles title and contents independently
- Returns full item with tags after update

**Deletion Cleanup**:
1. Delete all item-tag relationships
2. Delete semantic index entries (ignores errors)
3. Delete item record

#### Error Handling

- Returns item with tags in response (requires second query)
- Ignores semantic index deletion errors (may not exist)
- Wraps all errors with context

---

### 9. Logseq Service

**Location**: `/home/user/garden/internal/domain/service/logseq.go`

#### Responsibilities

Bi-directional synchronization with Logseq markdown files:
- Git repository management (clone/pull/push)
- Parse Logseq markdown with frontmatter
- Sync entities to/from markdown pages
- Conflict detection and resolution
- Force sync operations (DB→File, File→DB)

#### Dependencies

- `input.ConfigurationUseCase`: Git settings (URL, path, SSH key)
- `output.EntityRepository`: Entity persistence

#### Key Business Logic

**Sync Strategy**:
```
1. Pull latest from Git
2. Sync Pages → Entities (creates/updates)
3. Sync Entities → Pages (creates/updates)
4. Commit and push changes
5. Update last sync timestamp
```

**Frontmatter Structure**:
```yaml
---
id: <uuid>
title: <entity name>
last_sync: <timestamp>
---
```

**Conflict Resolution**:
- Compares `last_sync` timestamps
- If both changed since last sync, prefers newer modification
- Tracks timestamps in both DB (entity properties) and Git (frontmatter)

**Hard Sync Check**:
- Scans all entities with `logseq_path` property
- Scans all markdown files in pages directory
- Returns three lists:
  - Missing in DB (files without entities)
  - Missing in Git (entities without files)
  - Out of sync (timestamp mismatch)

**SSH Key Management**:
- Creates temporary SSH key file for Git operations
- Sets restrictive permissions (0600)
- Cleans up temp file after use

#### Error Handling

- Collects errors in stats structure (doesn't fail on first error)
- Validates file content (rejects Hugo templates)
- Handles missing frontmatter gracefully
- Stores Git command output on failure for debugging

---

### 10. Message Service

**Location**: `/home/user/garden/internal/domain/service/message.go`

#### Responsibilities

Manages chat/messaging data:
- Message retrieval by ID and room
- Full-text search across messages
- Text representation management
- Bulk content export

#### Dependencies

- `output.MessageRepository`: Message data access

#### Key Business Logic

**Pagination**:
- Limits page size to max 100 items
- Supports `beforeDatetime` cursor for time-based pagination
- Returns -1 for total/totalPages (not calculated for efficiency)

**Search Implementation**:
- Uses full-text search (FTS) on message bodies
- Falls back to total=-1 (expensive to calculate on FTS)
- Default page size: 50 for search

#### Error Handling

- Validates and normalizes page/pageSize inputs
- Wraps repository errors with operation context
- Simple delegation pattern

---

### 11. Note Service

**Location**: `/home/user/garden/internal/domain/service/note.go`

#### Responsibilities

Rich note-taking with entity references:
- Note CRUD operations
- Entity reference parsing and linking
- Tag management
- Semantic search capabilities
- Bidirectional entity relationship management

#### Dependencies

- `output.NoteRepository`: Note persistence
- `output.EmbeddingsService`: Semantic search embeddings

#### Key Business Logic

**Entity Reference Processing**:

**Storage Format** (internal):
```
[[entity-id]]                  → Simple reference
[[Display Text][entity-id]]    → Reference with display text
```

**Display Format** (user-facing):
```
[Entity Name](/entities/entity-id)      → Markdown link
[Display Text](/entities/entity-id)     → Custom display
```

**Auto-entity Creation**:
- Parses `[[EntityName]]` references
- Creates entity if not found
- Creates bidirectional reference (note → entity, entity → note)
- Updates content with entity UUIDs

**Content Processing Pipeline**:

_On Create_:
1. Create note with original content
2. Parse entity references
3. Create/lookup entities
4. Update note with processed content (UUIDs)
5. Create entity relationship for note itself

_On Read_:
1. Fetch note with UUID references
2. Convert UUIDs to display links
3. Return processed content to user

_On Update_:
1. Delete old entity references
2. Parse new references
3. Create/update entities
4. Update note content

#### Error Handling

- Creates entities automatically (no failure on missing entity)
- Gracefully handles missing entity relationships
- Wraps all errors with context
- Continues processing on individual reference failures

---

### 12. Observation Service

**Location**: `/home/user/garden/internal/domain/service/observation.go`

#### Responsibilities

User feedback and observation tracking:
- Store user feedback on Q&A pairs
- Track feedback statistics per bookmark
- Manage content reference deletion

#### Dependencies

- `output.ObservationRepository`: Observation persistence

#### Key Business Logic

**Feedback Data Structure**:
```json
{
  "question": "...",
  "answer": "...",
  "bookmarkId": "uuid",
  "userQuestion": "...",
  "similarity": 0.95,
  "feedbackType": "helpful|not-helpful",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

**Observation Types**:
- `qa-feedback`: User feedback on answers
- `qa-edit`: Question/answer modifications
- `qa-delete`: Q&A deletions

**Content Reference Cleanup**:
- Optionally deletes bookmark content reference when storing feedback
- Logs warning but doesn't fail if deletion fails

#### Error Handling

- Marshals feedback to JSON (fails on marshal error)
- Wraps all errors with context
- Warns on deletion failures but continues

---

### 13. Room Service

**Location**: `/home/user/garden/internal/domain/service/room.go`

#### Responsibilities

Chat room/conversation management:
- Room listing with search
- Message retrieval with context
- Room participant tracking
- Custom room naming
- Search within rooms with context messages

#### Dependencies

- `output.RoomRepository`: Room and message data access

#### Key Business Logic

**Room Details Assembly**:
```go
Returns:
  - Room metadata
  - Participant list
  - Recent messages (50)
  - Contact info map (for message senders)
```

**Context Message Fetching**:
- Search returns matching messages
- Fetches 3 messages before and after each match
- Provides conversation context for search results

**Message Pagination**:
- Supports cursor-based pagination with `beforeMessageID`
- Converts message ID to datetime for filtering
- Limits to 100 messages per page max

#### Error Handling

- Validates and caps page size
- Wraps all errors with operation context
- Returns nil for non-existent rooms (not an error)

---

### 14. Search Service

**Location**: `/home/user/garden/internal/domain/service/search.go`

#### Responsibilities

Advanced search across application:
- Unified search (bookmarks, notes, contacts, etc.)
- LLM-powered question answering
- Semantic similarity search
- Template-based prompt generation

#### Dependencies

- `output.SearchRepository`: Search queries
- `output.EmbeddingService`: Vector embeddings
- `output.LLMService`: Language model for answers
- `input.ConfigurationUseCase`: Prompt template storage

#### Key Business Logic

**Unified Search Weighting**:
```go
Score = (exactMatchWeight × exactMatch) +
        (similarityWeight × cosineSimilarity) +
        (recencyWeight × recencyScore)
```

**Advanced Search Pipeline**:
1. **Embed Query** → Generate vector embedding
2. **Find Similar** → Retrieve top 10 similar Q&As from bookmarks
3. **Build Context** → Load question, answer, summary, URL
4. **Apply Template** → Render prompt with context
5. **Call LLM** → Get response with optional `<think>` tags
6. **Parse Response** → Extract thinking process and final answer

**Default Prompt Template**:
- Instructs LLM to cite sources with markdown links
- Provides context as question-answer pairs
- Requests reliance on provided context
- Admits when context insufficient

**Response Parsing**:
```go
<think>reasoning here</think>
Final answer here
```
- Extracts thinking process for transparency
- Separates from user-facing answer

#### Error Handling

- Falls back to default template if config missing
- Returns full response if parsing fails
- Wraps all errors with context

---

### 15. Session Service

**Location**: `/home/user/garden/internal/domain/service/session.go`

#### Responsibilities

Conversation session management:
- Session search (semantic and text)
- Room session summaries
- Contact session history
- Timeline visualization
- Session message retrieval

#### Dependencies

- `output.SessionRepository`: Session data access
- `output.EmbeddingService`: Semantic search

#### Key Business Logic

**Dual Search Strategy**:
1. Try semantic search with embeddings first
2. Fall back to text search if embedding fails or no results
3. Pattern: `%query%` for LIKE matching

**Timeline Generation**:
```go
For each session:
  - Group by month (YYYY-MM)
  - Calculate duration (first → last message)
  - Aggregate participant activity
  - Compute totals and averages per month
```

**Session Messages**:
- Retrieves all messages in chronological order
- Batch loads contact info for all senders
- Returns contact map for efficient lookup

#### Error Handling

- Gracefully handles missing embeddings (falls back)
- Wraps all errors with context
- Returns sorted timeline data

---

### 16. Social Post Service

**Location**: `/home/user/garden/internal/domain/service/social_post.go`

#### Responsibilities

Cross-platform social media posting:
- Post to Twitter and Bluesky simultaneously
- OAuth credential management
- Retry logic with exponential backoff
- Post status tracking

#### Dependencies

- `output.SocialPostRepository`: Post persistence
- `output.SocialMediaService`: Platform APIs

#### Key Business Logic

**Multi-platform Posting**:
```go
Status Logic:
  - Both succeed → "completed"
  - One succeeds → "partial" (with error details)
  - Both fail → "failed" (with both errors)
```

**Retry Strategy**:
- Default: 3 attempts with 2-second initial delay
- Exponential backoff: `delay × 1.5^attempt`
- Jitter: Random factor (0.85-1.15) to prevent thundering herd
- Context-aware: Respects cancellation

**OAuth Flow**:
1. `InitiateTwitterAuth()` → Returns authorization URL
2. User authorizes in browser
3. `HandleTwitterCallback()` → Exchanges code for tokens
4. `UpdateTwitterTokens()` → Stores for future use

**Credential Checking**:
- Tests both platforms independently
- Returns profile info if working
- Marks as configured but not working if error

#### Error Handling

- Retries on transient failures
- Stores partial success (one platform fails)
- Collects error messages for both platforms
- Returns combined error information

---

### 17. Tag Service

**Location**: `/home/user/garden/internal/domain/service/tag.go`

#### Responsibilities

Tag management for items/notes:
- Tag CRUD operations
- Tag-item relationship management
- Tag usage statistics
- Item filtering by tag

#### Dependencies

- `output.TagRepository`: Tag persistence

#### Key Business Logic

**Tag Auto-creation**:
- `AddTag()` upserts tag (creates if doesn't exist)
- Uses current timestamp for creation tracking
- Creates item-tag relationship atomically

**Tag Deletion**:
1. Delete all relationships first
2. Then delete tag record
3. Prevents orphaned relationships

**Usage Statistics**:
- Optional parameter for including usage counts
- Counts number of items per tag
- Returns zero usage if not requested

#### Error Handling

- Returns error if tag not found on remove
- Wraps all repository errors
- Validates tag existence before operations

---

### 18. Utility Service

**Location**: `/home/user/garden/internal/domain/service/utility.go`

#### Responsibilities

System utilities and maintenance:
- Debug information aggregation
- Database health checks
- Stale conversation cleanup
- Uptime tracking

#### Dependencies

- `output.SessionRepository`: Session cleanup
- `output.MessageRepository`: Message access
- `output.ConfigurationRepository`: Config inspection
- `*pgxpool.Pool`: Direct database connection

#### Key Business Logic

**Debug Information**:
- Tests database connectivity
- Retrieves non-sensitive configuration
- Filters out keys containing: secret, password, token, key, credential, auth
- Calculates uptime since service start
- Returns version information

**Stale Conversation Cleanup**:
- Deletes sessions older than 30 days with no messages
- Configurable threshold (currently hardcoded)
- Returns count of deleted sessions

#### Error Handling

- Captures database ping errors in status string
- Gracefully handles configuration fetch failures
- Returns partial results if some operations fail

---

## Common Business Logic Patterns

### 1. Pagination Pattern

Most list operations follow this pattern:

```go
// Normalize inputs
if page < 1 { page = 1 }
if pageSize < 1 { pageSize = 10 }

// Calculate offset
offset := (page - 1) * pageSize

// Query data
items, err := repo.List(ctx, pageSize, offset)
total, err := repo.Count(ctx)

// Calculate pages
totalPages := int32((total + int64(pageSize) - 1) / int64(pageSize))

// Return response
return &PaginatedResponse[T]{
    Data:       items,
    Total:      total,
    Page:       page,
    PageSize:   pageSize,
    TotalPages: totalPages,
}
```

### 2. Retry Pattern with Exponential Backoff

Used in SocialPostService for external API calls:

```go
func retry(operation func() (string, error)) (string, error) {
    for attempt := 0; attempt < retryCount; attempt++ {
        result, err := operation()
        if err == nil { return result, nil }

        // Exponential backoff: delay × 1.5^attempt
        // Jitter: ±15% randomization
        jitter := 0.85 + rand.Float64() * 0.3
        backoff := delay * math.Pow(1.5, attempt) * jitter

        time.Sleep(backoff)
    }
    return "", lastError
}
```

### 3. Batch Loading Pattern

Reduces N+1 queries by batch loading related data:

```go
// Extract IDs
contactIDs := extractUniqueIDs(items)

// Batch load
contacts, err := repo.GetContactsByIDs(ctx, contactIDs)

// Build lookup map
contactMap := make(map[uuid.UUID]Contact)
for _, c := range contacts {
    contactMap[c.ID] = c
}

// Attach to results
for i := range items {
    items[i].Contact = contactMap[items[i].ContactID]
}
```

### 4. Entity Reference Parsing

Pattern for wiki-style links in content:

```go
// Regex: \[\[(.*?)(?:\]\[([^\]]+))?\]\]
matches := entityRefRegex.FindAllStringSubmatch(content, -1)

for _, match := range matches {
    entityName := match[1]
    displayText := match[2] // optional

    // Lookup or create entity
    entity := lookupOrCreate(entityName)

    // Replace with ID
    if displayText != "" {
        content = replace(match[0], "[[" + displayText + "][" + entity.ID + "]]")
    } else {
        content = replace(match[0], "[[" + entity.ID + "]]")
    }
}
```

### 5. Composite Data Assembly

Building rich responses with multiple related queries:

```go
// Main entity
entity, err := repo.GetEntity(ctx, id)

// Related data (can be parallel)
tags, err := repo.GetTags(ctx, id)
relationships, err := repo.GetRelationships(ctx, id)
references, err := repo.GetReferences(ctx, id)

// Assemble
return &FullEntity{
    Entity:        entity,
    Tags:          tags,
    Relationships: relationships,
    References:    references,
}
```

---

## Dependencies and Interfaces

### Repository Interfaces (Output Ports)

Services depend on repository interfaces defined in `internal/port/output`:

- **BookmarkRepository**: CRUD, search, embedding storage
- **BrowserHistoryRepository**: History data access, domain analytics
- **CategoryRepository**: Category CRUD, merging
- **ConfigurationRepository**: Key-value storage, type conversion
- **ContactRepository**: Contact CRUD, evaluations, tags, stats
- **DashboardRepository**: Aggregation queries
- **EntityRepository**: Entity graph management
- **ItemRepository**: Item CRUD, tag relationships
- **MessageRepository**: Message access, full-text search
- **NoteRepository**: Note CRUD, entity references
- **ObservationRepository**: Feedback storage
- **RoomRepository**: Room/conversation management
- **SearchRepository**: Unified search, similarity queries
- **SessionRepository**: Session management, timeline
- **SocialPostRepository**: Social media post tracking
- **TagRepository**: Tag management, usage stats

### External Service Interfaces

- **HTTPFetcher**: Web content retrieval
- **EmbeddingsService/EmbeddingService**: Vector embedding generation
- **AIService**: Text summarization
- **ContentProcessor**: Content extraction (Lynx, Reader)
- **LLMService**: Language model inference
- **SocialMediaService**: Twitter/Bluesky APIs

### Use Case Interfaces (Input Ports)

Services implement use case interfaces defined in `internal/port/input`:

- **BookmarkUseCase**
- **BrowserHistoryUseCase**
- **ConfigurationUseCase** (implements and depends on)
- **ContactUseCase**
- **EntityUseCase**
- **ItemUseCase**
- **LogseqSyncUseCase**
- **MessageUseCase**
- **NoteUseCase**
- **RoomUseCase**
- **SessionUseCase**
- **SocialPostUseCase**

---

## Error Handling Patterns

### 1. Error Wrapping with Context

All services wrap errors with contextual information:

```go
if err != nil {
    return nil, fmt.Errorf("failed to <operation>: %w", err)
}
```

This creates an error chain that can be unwrapped for inspection and provides a clear call stack.

### 2. Input Validation

Services validate and normalize inputs before processing:

```go
// Required field validation
if input.Key == "" {
    return nil, fmt.Errorf("key is required")
}

// Range normalization
if page < 1 { page = 1 }
if limit < 1 { limit = 10 }
if limit > 100 { limit = 100 }
```

### 3. Graceful Degradation

Some services continue operation despite partial failures:

```go
// Logseq sync collects errors instead of failing
stats.Errors = append(stats.Errors, fmt.Sprintf("Error: %v", err))

// Social post returns partial success
if oneSucceeded { status = "partial" } else { status = "failed" }
```

### 4. Optional Related Data

Related data fetches handle missing data gracefully:

```go
evaluation, err := repo.GetEvaluation(ctx, id)
if err != nil {
    // Not an error - evaluation might not exist
    evaluation = nil
}
```

### 5. Cleanup on Error

Some operations clean up partial state on failure:

```go
created, err := createItem(ctx, input)
if err != nil {
    // Cleanup if needed
    return nil, err
}

// Continue with subsequent operations
if err := addTags(ctx, created.ID, tags); err != nil {
    deleteItem(ctx, created.ID) // Rollback
    return nil, err
}
```

### 6. Transaction Boundaries

While not explicitly shown in services (handled at repository layer), services are designed to support transactions:

- Single-entity operations are atomic
- Multi-step operations should be wrapped in repository transactions
- Services don't directly manage transactions (repository responsibility)

### 7. Timeout Handling

Operations that call external services use context timeouts:

```go
fetchCtx, cancel := context.WithTimeout(ctx, 25*time.Second)
defer cancel()

response, err := httpFetcher.Fetch(fetchCtx, url)
if err != nil {
    // Store error for debugging
    storeError(ctx, err)
    return nil, err
}
```

### 8. Error Classification

Some services classify errors for better handling:

```go
// Retriable errors
if isRetriable(err) {
    return retry(operation)
}

// User errors vs system errors
if isValidationError(err) {
    return &ValidationError{...}
}
```

---

## Testing Considerations

### Service Testing Strategy

Services are designed to be testable:

1. **Constructor Injection**: All dependencies injected via constructor
2. **Interface Dependencies**: Can mock repositories and external services
3. **Context Awareness**: Supports timeout and cancellation in tests
4. **No Global State**: Each service instance is independent

### Mock Requirements

To test services, provide mocks for:

```go
type MockBookmarkRepository struct {
    // Implement output.BookmarkRepository interface
}

func TestBookmarkService_CreateEmbedding(t *testing.T) {
    mockRepo := &MockBookmarkRepository{
        GetProcessedContentFunc: func(ctx, id) (string, error) {
            return "test content", nil
        },
    }

    mockEmbeddings := &MockEmbeddingsService{
        GetEmbeddingFunc: func(ctx, text) ([]Embedding, error) {
            return []Embedding{{Embedding: []float32{0.1, 0.2}}}, nil
        },
    }

    service := NewBookmarkService(mockRepo, nil, mockEmbeddings, nil, nil)

    result, err := service.CreateEmbeddingChunks(ctx, testID)
    // assertions...
}
```

---

## Performance Considerations

### 1. Database Query Optimization

- **Batch Loading**: Load related data in bulk instead of N+1 queries
- **Pagination**: Always use LIMIT/OFFSET for large result sets
- **Indexed Searches**: Rely on database indexes for filtering
- **Projection**: Select only needed columns (handled at repository layer)

### 2. Caching Opportunities

Services that could benefit from caching:

- **Configuration Service**: Frequently accessed config values
- **Category Service**: Category hierarchy rarely changes
- **Contact Tags**: Tag list relatively stable
- **Dashboard Stats**: Expensive aggregations

### 3. Async Processing

Some operations should be async (not currently implemented):

- Bookmark content fetching (slow HTTP)
- Embedding generation (slow AI inference)
- Social media posting (external API)
- Logseq synchronization (Git operations)

### 4. Resource Limits

Services implement limits to prevent abuse:

- **Bookmark Embedding**: Max 20 chunks per bookmark
- **Message Pagination**: Max 100 messages per page
- **Content Processing**: Truncates at 15,000 characters
- **Search Results**: Default limit of 50 items

---

## Security Considerations

### 1. Input Validation

- All user inputs validated before processing
- SQL injection prevented by parameterized queries (at repository layer)
- Path traversal prevented (Logseq file operations)

### 2. Secret Management

- Configuration service distinguishes secret vs non-secret values
- Debug endpoint filters sensitive configuration keys
- OAuth tokens stored in configuration (encrypted at rest recommended)

### 3. Rate Limiting

Not implemented at service layer (should be handled at API gateway/middleware):

- External API calls (Twitter, Bluesky)
- Embedding generation
- LLM inference

### 4. Authorization

Not implemented at service layer (should be handled at handler layer):

- Services assume caller has been authorized
- Multi-tenant isolation would require user context

### 5. Content Sanitization

- URL sanitization in bookmark service
- Filename sanitization in Logseq service
- No HTML/script sanitization (assume trusted users or sanitize at presentation layer)

---

## Future Enhancements

### Potential Improvements

1. **Event Sourcing**: Emit domain events for audit trails
2. **Command/Query Separation (CQRS)**: Split read/write models
3. **Circuit Breaker**: Protect external service calls
4. **Metrics/Observability**: Add structured logging and metrics
5. **Async Workers**: Background job processing for slow operations
6. **Caching Layer**: Add Redis for frequently accessed data
7. **Webhook Support**: Notify external systems of changes
8. **Bulk Operations**: More efficient batch processing
9. **Soft Delete**: Consistent soft delete across all entities
10. **Versioning**: Track entity version history

### Code Quality

1. **Unit Tests**: Comprehensive test coverage for business logic
2. **Integration Tests**: Test service + repository together
3. **Documentation**: Godoc comments for all exported functions
4. **Validation Library**: Use structured validation instead of manual checks
5. **Error Types**: Define custom error types for better handling

---

## Conclusion

The service layer provides a clean separation between business logic and infrastructure concerns. Services are designed to be:

- **Testable**: Dependencies injected, interfaces used throughout
- **Maintainable**: Clear responsibilities, consistent patterns
- **Scalable**: Stateless, can be horizontally scaled
- **Robust**: Comprehensive error handling, graceful degradation
- **Flexible**: Easy to swap implementations (e.g., different repositories)

The use of interfaces (ports) allows the application to remain independent of specific infrastructure choices while maintaining a clear architecture that's easy to understand and extend.
