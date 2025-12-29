# PostgreSQL Adapter Documentation

## Overview

The PostgreSQL adapter provides the data persistence layer for the Garden application using PostgreSQL with pgx/v5 driver and pgvector extension for vector embeddings. The adapter follows the Repository pattern and is located at `/home/user/garden/internal/adapter/secondary/postgres/`.

## Table of Contents

1. [Database Connection Setup](#database-connection-setup)
2. [Repository Pattern Implementation](#repository-pattern-implementation)
3. [Repository Methods and Queries](#repository-methods-and-queries)
4. [Data Converters](#data-converters)
5. [Transaction Handling](#transaction-handling)
6. [pgvector Usage for Embeddings](#pgvector-usage-for-embeddings)

---

## Database Connection Setup

### DB Structure

The main database connection is defined in `/home/user/garden/internal/adapter/secondary/postgres/db.go`:

```go
type DB struct {
    Pool *pgxpool.Pool
}
```

### Connection Initialization

The `NewDB` function creates a new database connection pool:

```go
func NewDB(ctx context.Context) (*DB, error)
```

**Features:**
- Reads connection string from `DATABASE_URL` environment variable
- Falls back to default connection: `postgres://gardener@localhost:5432/garden?sslmode=disable`
- Uses `pgxpool` for connection pooling
- Validates connection with a ping before returning
- Returns error with context if any step fails

**Example Usage:**
```go
db, err := postgres.NewDB(context.Background())
if err != nil {
    log.Fatal(err)
}
defer db.Close()
```

### Connection Pool Configuration

The connection pool is configured using `pgxpool.ParseConfig()` and `pgxpool.NewWithConfig()`, providing:
- Automatic connection pooling
- Connection health monitoring
- Prepared statement caching
- Connection lifecycle management

---

## Repository Pattern Implementation

### Architecture

Each domain aggregate has a corresponding repository implementation:

```
internal/adapter/secondary/postgres/repository/
├── bookmark.go           - Bookmark and web content management
├── browser_history.go    - Browser history tracking
├── contact.go            - Contact information and relationships
├── configuration.go      - Application configuration
├── category.go           - Category management
├── dashboard.go          - Dashboard data aggregation
├── entity.go             - Entity and relationship management
├── message.go            - Message storage and retrieval
├── item.go               - Item/note management
├── note.go               - Note operations with transactions
├── observation.go        - Observation and feedback tracking
├── room.go               - Chat room management
├── search.go             - Unified search functionality
├── session.go            - Conversation session management
├── social_post.go        - Social media post tracking
├── tag.go                - Tag management
└── converters.go         - Type conversion utilities
```

### Repository Structure Pattern

Each repository follows this structure:

```go
type XxxRepository struct {
    pool *pgxpool.Pool
}

func NewXxxRepository(pool *pgxpool.Pool) *XxxRepository {
    return &XxxRepository{
        pool: pool,
    }
}
```

### Generated Queries

Repositories use sqlc-generated code from `internal/adapter/secondary/postgres/generated/db/`:
- Type-safe query execution
- Automatic parameter binding
- Struct mapping
- Reduced boilerplate code

---

## Repository Methods and Queries

### BookmarkRepository

**Location:** `/home/user/garden/internal/adapter/secondary/postgres/repository/bookmark.go`

#### Core Methods

**GetBookmark(ctx, bookmarkID) -> *entity.Bookmark**
- Retrieves a single bookmark by ID
- Returns bookmark with URL and creation date

**ListBookmarks(ctx, categoryID, searchQuery, startDate, endDate, limit, offset) -> []entity.BookmarkWithTitle**
- Paginated bookmark listing with filtering
- Supports category filtering, text search, and date range
- Returns bookmarks with titles

**CountBookmarks(ctx, categoryID, searchQuery, startDate, endDate) -> int64**
- Counts bookmarks matching filter criteria
- Used for pagination

**GetRandomBookmark(ctx) -> uuid.UUID**
- Returns a random bookmark ID

**GetBookmarkDetails(ctx, bookmarkID) -> *entity.BookmarkDetails**
- Retrieves comprehensive bookmark information including:
  - URL and metadata
  - Category name
  - Raw source HTML
  - Lynx-extracted content
  - Reader-mode content
  - AI-generated summary
  - HTTP response details

#### Question Management

**GetBookmarkQuestions(ctx, bookmarkID) -> []entity.BookmarkQuestion**
- Retrieves Q&A pairs associated with a bookmark

**UpdateBookmarkQuestion(ctx, content, embedding, referenceID, bookmarkID)**
- Updates question content and embedding vector
- Uses pgvector for semantic search

**DeleteBookmarkQuestion(ctx, referenceID, bookmarkID)**
- Removes a Q&A pair

**GetBookmarkForQuestion(ctx, bookmarkID) -> *entity.BookmarkDetails**
- Retrieves bookmark context for Q&A generation

#### HTTP Response Management

**InsertHttpResponse(ctx, bookmarkID, statusCode, headers, content, fetchDate)**
- Stores raw HTTP response data
- Captures status code, headers, and body

**GetLatestHttpResponse(ctx, bookmarkID) -> *output.HTTPResponse**
- Retrieves most recent HTTP response for a bookmark

**GetMissingHttpResponses(ctx) -> []entity.Bookmark**
- Finds bookmarks without HTTP responses (needs fetching)

#### Content Processing

**InsertProcessedContent(ctx, bookmarkID, strategyUsed, processedContent)**
- Stores content processed by various extraction strategies
- Strategy examples: "reader", "lynx", "llm-summary"

**GetProcessedContentByStrategy(ctx, bookmarkID, strategy) -> *string**
- Retrieves processed content for a specific strategy

**GetMissingReaderContent(ctx) -> []entity.Bookmark**
- Finds bookmarks without reader-mode content

#### Embedding and Semantic Search

**CreateEmbeddingChunk(ctx, bookmarkID, content, strategy, embedding) -> uuid.UUID**
- Stores content chunk with embedding vector
- Returns chunk ID
- Used for semantic search

**SearchSimilarBookmarks(ctx, embedding, strategy, limit) -> []entity.BookmarkWithTitle**
- Finds bookmarks similar to query embedding
- Uses pgvector cosine similarity
- Strategy-aware search

**InsertBookmarkTitle(ctx, bookmarkID, title, source)**
- Stores bookmark title from various sources
- Source examples: "html-meta", "llm-generated"

**GetBookmarkTitle(ctx, bookmarkID) -> *output.TitleData**
- Retrieves title data for processing

#### Observations

**CreateObservation(ctx, data, observationType, source, tags, ref)**
- Creates an observation record (feedback, notes, etc.)
- Linked to bookmark via ref UUID

---

### BrowserHistoryRepository

**Location:** `/home/user/garden/internal/adapter/secondary/postgres/repository/browser_history.go`

**ListHistory(ctx, filters) -> ([]entity.BrowserHistory, int64)**
- Paginated history listing with filters
- Filters: search query, domain, date range
- Returns items and total count

**TopDomains(ctx, limit) -> []entity.DomainVisitCount**
- Returns most visited domains with visit counts

**RecentHistory(ctx, limit) -> []entity.BrowserHistory**
- Returns N most recent history entries

---

### ContactRepository

**Location:** `/home/user/garden/internal/adapter/secondary/postgres/repository/contact.go**

#### CRUD Operations

**GetContact(ctx, contactID) -> *entity.Contact**
- Retrieves contact with all fields
- Includes JSON extras field
- Includes computed stats (last week messages, groups in common)

**ListContacts(ctx, limit, offset) -> []entity.Contact**
- Paginated contact listing

**SearchContacts(ctx, searchPattern, limit, offset) -> []entity.Contact**
- Text search across contact names

**CountContacts(ctx) -> int64**
- Total contact count

**CreateContact(ctx, input) -> *entity.Contact**
- Creates new contact
- Handles JSON extras marshaling

**UpdateContact(ctx, contactID, input)**
- Updates contact fields
- Uses COALESCE for partial updates

**DeleteContact(ctx, contactID)**
- Removes contact

#### Relationships

**GetContactEvaluation(ctx, contactID) -> *entity.ContactEvaluation**
- Retrieves importance, closeness, fondness ratings

**EvaluationExists(ctx, contactID) -> bool**
- Checks if evaluation record exists

**CreateEvaluation(ctx, contactID, input)**
- Creates evaluation record

**UpdateEvaluation(ctx, contactID, input)**
- Updates evaluation ratings

**GetContactTags(ctx, contactID) -> []entity.ContactTag**
- Retrieves tags assigned to contact

**GetContactKnownNames(ctx, contactID) -> []entity.ContactKnownName**
- Retrieves alternate names for contact

**GetContactRooms(ctx, contactID) -> []entity.ContactRoom**
- Retrieves chat rooms where contact participates

**GetContactSources(ctx, contactID) -> []entity.ContactSource**
- Retrieves external source mappings (Matrix IDs, etc.)

#### Tag Management

**GetAllTagNames(ctx) -> []entity.ContactTagName**
- Lists all available contact tags

**GetContactTagByName(ctx, tagName) -> *entity.ContactTagName**
- Finds tag by name

**CreateTagName(ctx, tagName) -> *entity.ContactTagName**
- Creates new tag

**ContactTagExists(ctx, contactID, tagID) -> bool**
- Checks if contact has specific tag

**AddContactTag(ctx, contactID, tagID)**
- Assigns tag to contact

**RemoveContactTag(ctx, contactID, tagID)**
- Removes tag from contact

#### Batch Operations

**GetEvaluationsByContactIDs(ctx, contactIDs) -> map[uuid.UUID]*entity.ContactEvaluation**
- Batch retrieves evaluations
- Optimizes N+1 queries

**GetTagsByContactIDs(ctx, contactIDs) -> map[uuid.UUID][]entity.ContactTag**
- Batch retrieves tags
- Optimizes N+1 queries

#### Statistics

**EnsureContactStatsExist(ctx)**
- Initializes contact statistics table

**UpdateContactMessageStats(ctx)**
- Updates message counts for all contacts

#### Source Management

**ListAllContactSources(ctx) -> []entity.ContactSource**
- Lists all contact source mappings

**CreateContactSource(ctx, input) -> *entity.ContactSource**
- Creates external source mapping

**UpdateContactSource(ctx, id, input) -> *entity.ContactSource**
- Updates source mapping

**DeleteContactSource(ctx, id)**
- Removes source mapping

#### Advanced Operations

**MergeContacts(ctx, sourceID, targetID)**
- Merges two contact records
- Moves all relationships to target

---

### NoteRepository

**Location:** `/home/user/garden/internal/adapter/secondary/postgres/repository/note.go`

**Note:** This repository implements transaction support.

#### Transaction Support

**BeginTx(ctx) -> interface{}**
- Starts new transaction
- Returns pgx.Tx

**CommitTx(ctx, tx)**
- Commits transaction

**RollbackTx(ctx, tx)**
- Rolls back transaction

**WithTx(ctx, tx) -> output.NoteRepository**
- Returns repository instance using transaction
- Enables transactional operations

**getQuerier() -> db.DBTX**
- Internal method returning active transaction or pool
- Ensures queries use correct context

#### CRUD Operations

**GetNote(ctx, noteID) -> *entity.Note**
- Retrieves note by ID

**GetNoteWithTags(ctx, noteID) -> (*entity.Note, []string)**
- Retrieves note with associated tags

**ListNotes(ctx, limit, offset) -> []entity.NoteListItem**
- Paginated note listing with tags

**SearchNotes(ctx, searchPattern, limit, offset) -> []entity.NoteListItem**
- Text search in note titles

**CountNotes(ctx) -> int64**
- Total note count

**CountSearchNotes(ctx, searchPattern) -> int64**
- Count of matching search results

**CreateNote(ctx, title, slug, contents) -> *entity.Note**
- Creates new note

**UpdateNote(ctx, noteID, title, contents)**
- Updates note content

**DeleteNote(ctx, noteID)**
- Removes note

#### Entity Management

**GetNoteEntityRelationship(ctx, noteID) -> *uuid.UUID**
- Gets entity ID linked to note

**CreateEntity(ctx, name, entityType, description, properties) -> *uuid.UUID**
- Creates entity record
- Properties stored as JSON

**CreateEntityRelationship(ctx, entityID, relatedType, relatedID, relationshipType, metadata)**
- Creates relationship between entities
- Metadata stored as JSON

**DeleteEntityRelationships(ctx, entityID)**
- Removes all relationships for entity

**SoftDeleteEntity(ctx, entityID)**
- Marks entity as deleted

**GetEntityByID(ctx, entityID) -> *entity.Entity**
- Retrieves entity by ID

**GetEntityByName(ctx, name) -> *entity.Entity**
- Retrieves entity by name

**CreateEntityReference(ctx, ref)**
- Creates reference from note to entity
- Stores position in text

**DeleteNoteEntityReferences(ctx, noteID)**
- Removes all entity references in note

#### Tag Management

**GetNoteTagByName(ctx, tagName) -> *entity.NoteTag**
- Finds tag by name

**UpsertNoteTag(ctx, tagName, lastActivity) -> *entity.NoteTag**
- Creates or updates tag
- Updates last activity timestamp

**CreateItemTag(ctx, itemID, tagID)**
- Associates tag with note

**DeleteNoteItemTags(ctx, noteID)**
- Removes all tags from note

**GetAllNoteTags(ctx) -> []entity.NoteTag**
- Lists all available tags

#### Semantic Search

**DeleteNoteSemanticIndex(ctx, noteID)**
- Removes semantic index for note

**SearchSimilarNotes(ctx, embedding, strategy, limit) -> []entity.NoteListItem**
- Finds similar notes using vector embeddings
- Raw SQL implementation with pgvector
- Uses cosine distance operator (`<=>`)

---

### ItemRepository

**Location:** `/home/user/garden/internal/adapter/secondary/postgres/repository/item.go`

**Note:** Items are a variant of notes in the system.

#### CRUD Operations

**GetItem(ctx, itemID) -> *entity.Item**
- Retrieves item by ID

**GetItemWithTags(ctx, itemID) -> (*entity.Item, []string)**
- Retrieves item with tags

**ListItems(ctx, limit, offset) -> []entity.ItemListItem**
- Paginated item listing

**CountItems(ctx) -> int64**
- Total item count

**CreateItem(ctx, title, slug, contents) -> *entity.Item**
- Creates new item

**UpdateItem(ctx, itemID, title, contents)**
- Updates item content

**DeleteItem(ctx, itemID)**
- Removes item

#### Tag Management

**GetItemTagByName(ctx, tagName) -> *entity.ItemTag**
- Finds item tag by name

**UpsertItemTag(ctx, tagName, lastActivity) -> *entity.ItemTag**
- Creates or updates item tag

**CreateItemTagRelation(ctx, itemID, tagID)**
- Associates tag with item

**DeleteItemTags(ctx, itemID)**
- Removes all tags from item

**GetItemTags(ctx, itemID) -> []string**
- Retrieves tag names for item

#### Semantic Search

**DeleteItemSemanticIndex(ctx, itemID)**
- Removes semantic index for item

**SearchSimilarItems(ctx, embedding, limit) -> []entity.ItemListItem**
- Finds similar items using vector embeddings
- Uses pgvector cosine distance

---

### MessageRepository

**Location:** `/home/user/garden/internal/adapter/secondary/postgres/repository/message.go`

**GetMessage(ctx, messageID) -> *entity.Message**
- Retrieves message with sender details

**SearchMessages(ctx, query, limit, offset) -> []entity.RoomMessage**
- Full-text search in messages
- Uses PostgreSQL tsquery

**GetMessageContacts(ctx, contactIDs) -> []entity.Contact**
- Batch retrieves contacts for messages

**GetMessagesByIDs(ctx, messageIDs) -> []entity.Message**
- Batch retrieves messages
- Raw SQL with ANY() operator

**GetMessagesByRoomID(ctx, roomID, limit, offset, beforeDatetime) -> []entity.RoomMessage**
- Retrieves room messages with pagination
- Optional time filtering
- Includes transcription data and bookmark links

**GetAllMessageContents(ctx) -> []string**
- Retrieves all message text content
- Used for analysis

**GetMessageTextRepresentations(ctx, messageID) -> []entity.MessageTextRepresentation**
- Retrieves alternative text representations
- Examples: transcriptions, translations

---

### RoomRepository

**Location:** `/home/user/garden/internal/adapter/secondary/postgres/repository/room.go`

**ListRooms(ctx, limit, offset, searchText) -> []entity.Room**
- Paginated room listing
- Optional text search
- Includes participant count and last activity

**CountRooms(ctx, searchText) -> int64**
- Count rooms matching search

**GetRoom(ctx, roomID) -> *entity.Room**
- Retrieves room details

**GetRoomParticipants(ctx, roomID) -> []entity.RoomParticipant**
- Lists room participants with presence info

**GetRoomMessages(ctx, roomID, limit, offset, beforeDatetime) -> []entity.RoomMessage**
- Retrieves room messages
- Includes transcription data and bookmarks

**SearchRoomMessages(ctx, roomID, searchText, limit, offset) -> []entity.RoomMessage**
- Searches messages within specific room

**UpdateRoomName(ctx, roomID, name)**
- Updates user-defined room name

**GetSessionsCount(ctx, roomID) -> int32**
- Counts conversation sessions in room

**GetMessageContacts(ctx, contactIDs) -> []entity.Contact**
- Batch retrieves message senders

**GetBeforeMessageDatetime(ctx, messageID) -> *time.Time**
- Gets timestamp for pagination

**GetContextMessages(ctx, roomID, messageIDs, beforeCount, afterCount) -> []entity.RoomMessage**
- Retrieves context around specific messages
- Used for displaying search results with context

---

### SessionRepository

**Location:** `/home/user/garden/internal/adapter/secondary/postgres/repository/session.go`

#### Search Methods

**SearchSessionsWithEmbeddings(ctx, embedding, limit) -> []entity.SessionSearchResult**
- Semantic search using vector embeddings
- Returns sessions with similarity scores
- Uses pgvector cosine distance

**SearchSessionsWithText(ctx, searchPattern, limit) -> []entity.SessionSearchResult**
- Text-based session search
- Pattern matching on summaries

**SearchContactSessionSummaries(ctx, contactID, searchPattern) -> []entity.SessionSearchResult**
- Finds sessions where contact participated
- With text filtering

#### Session Queries

**GetSessionSummaries(ctx, roomID) -> []entity.SessionSummary**
- Lists all sessions in room with summaries

**GetSessionsForRoom(ctx, roomID) -> []entity.SessionWithMessages**
- Lists sessions with message counts

**GetSessionParticipantActivity(ctx, roomID) -> []entity.ParticipantActivity**
- Participant message counts per session

**GetSessionMessages(ctx, sessionID) -> []entity.SessionMessage**
- Retrieves all messages in session
- Includes transcription data

**GetContactsByIDs(ctx, contactIDs) -> []entity.SessionMessageContact**
- Batch retrieves contacts for session display

#### Maintenance

**DeleteStaleConversations(ctx, olderThan) -> int64**
- Removes old sessions with no messages
- Returns count of deleted sessions
- Raw SQL implementation

---

### SearchRepository

**Location:** `/home/user/garden/internal/adapter/secondary/postgres/repository/search.go`

**SearchAll(ctx, query, exactMatchWeight, similarityWeight, recencyWeight, limit) -> []entity.UnifiedSearchResult**
- Unified search across multiple entity types
- Weighted scoring combining:
  - Exact match score
  - Similarity score
  - Recency score
- Returns heterogeneous results (notes, bookmarks, contacts, etc.)

**GetSimilarQuestions(ctx, embedding, limit) -> []entity.RetrievedItem**
- Semantic search for Q&A pairs
- Uses pgvector for similarity
- Returns question-answer pairs with bookmark context
- Splits question/answer from newline-separated format

---

### ObservationRepository

**Location:** `/home/user/garden/internal/adapter/secondary/postgres/repository/observation.go`

**Create(ctx, data, obsType, source, tags, ref) -> *entity.Observation**
- Creates observation record
- Data stored as bytes (flexible format)
- Linked to parent entity via ref UUID

**GetFeedbackStats(ctx, bookmarkID) -> *entity.FeedbackStats**
- Aggregates feedback counts
- Returns upvotes, downvotes, trash counts

**DeleteBookmarkContentReference(ctx, referenceID, bookmarkID)**
- Removes specific content reference

---

### EntityRepository

**Location:** `/home/user/garden/internal/adapter/secondary/postgres/repository/entity.go`

#### CRUD Operations

**GetEntity(ctx, entityID) -> *entity.Entity**
- Retrieves entity by ID
- Properties stored as JSON

**CreateEntity(ctx, input) -> *entity.Entity**
- Creates new entity
- Type, name, description, properties (JSON)

**UpdateEntity(ctx, entityID, input) -> *entity.Entity**
- Partial update using COALESCE
- Updates timestamp automatically

**DeleteEntity(ctx, entityID)**
- Soft deletes entity (sets deleted_at)

**SearchEntities(ctx, query, entityType) -> []entity.Entity**
- Text search in entity names
- Optional type filtering

**ListEntities(ctx, entityType, updatedSince) -> []entity.Entity**
- Lists entities by type
- Optional timestamp filtering

**ListDeletedEntities(ctx) -> []entity.Entity**
- Lists soft-deleted entities

**ListEntitiesUpdatedSince(ctx, since) -> []entity.Entity**
- Raw SQL query for sync operations

**ListEntitiesByProperty(ctx, propertyKey, propertyValue) -> []entity.Entity**
- JSON property filtering
- Raw SQL with JSONB operators

#### Relationships

**GetEntityRelationships(ctx, entityID) -> []entity.EntityRelationship**
- Lists all relationships for entity

**GetEntityRelationshipsByTypeAndRelationship(ctx, entityID, relatedType, relationshipType) -> []entity.EntityRelationship**
- Filtered relationship query

**CreateEntityRelationship(ctx, entityID, relatedType, relatedID, relationshipType, metadata) -> *entity.EntityRelationship**
- Creates relationship with JSON metadata

**DeleteEntityRelationship(ctx, relationshipID)**
- Removes relationship

**UpdateContactNameByRelationship(ctx, entityID, relationshipType, name)**
- Updates contact via entity relationship

#### References

**GetEntityReferences(ctx, entityID, sourceType) -> []entity.EntityReference**
- Gets references to entity
- Optional source type filter (e.g., "note", "message")
- Includes position in source text

---

### ConfigurationRepository

**Location:** `/home/user/garden/internal/adapter/secondary/postgres/repository/configuration.go`

**ListConfigurations(ctx, filter) -> []entity.Configuration**
- Lists configurations with filtering
- Can exclude secrets
- Supports key prefix filtering

**GetByKey(ctx, key) -> *entity.Configuration**
- Retrieves single configuration

**GetByPrefix(ctx, prefix, includeSecrets) -> []entity.Configuration**
- Retrieves configurations by key prefix

**Create(ctx, config) -> *entity.Configuration**
- Creates new configuration

**Update(ctx, key, update) -> *entity.Configuration**
- Updates existing configuration

**Delete(ctx, key)**
- Removes configuration

**Upsert(ctx, key, value, isSecret, updatedAt) -> *entity.Configuration**
- Creates or updates configuration
- Atomic operation

---

## Data Converters

**Location:** `/home/user/garden/internal/adapter/secondary/postgres/repository/converters.go`

The converters file provides type conversion utilities between pgtype and Go native types.

### Time Conversions

**convertTimeToPgTimestamp(t time.Time) -> pgtype.Timestamp**
- Converts Go time.Time to pgx timestamp
- Sets Valid flag to true

**convertPgTimestampToTime(ts pgtype.Timestamp) -> time.Time**
- Converts pgx timestamp to Go time
- Returns zero time if invalid

**convertPgTimestampToTimePtr(ts pgtype.Timestamp) -> *time.Time**
- Converts to pointer type
- Returns nil if invalid

**convertInterfaceToTimePtr(i interface{}) -> *time.Time**
- Converts interface{} to time pointer
- Handles type assertions

### UUID Conversions

**convertUUIDToPgUUID(id uuid.UUID) -> pgtype.UUID**
- Converts google/uuid to pgtype.UUID
- Sets Valid flag to true

**convertPgUUIDToUUIDPtr(id pgtype.UUID) -> *uuid.UUID**
- Converts pgtype.UUID to pointer
- Returns nil if invalid

### String Conversions

**convertStringToPtr(s string) -> *string**
- Converts empty string to nil
- Otherwise returns pointer

**convertInterfaceToStringSlice(i interface{}) -> []string**
- Converts PostgreSQL array to []string
- Used for array_agg results
- Filters out empty strings

### JSON Conversions

**convertMapToJSON(m map[string]interface{}) -> json.RawMessage**
- Marshals map to JSON
- Returns empty object "{}" on error

### Additional Converter Functions

Found in individual repository files:

**nullableString(s *string) -> *string**
- Pass-through for nullable strings

**nullableTimestamp(t *pgtype.Timestamp) -> *time.Time**
- Converts nullable timestamp

**nullableTimestamptz(t *pgtype.Timestamptz) -> *time.Time**
- Converts nullable timestamp with timezone

---

## Transaction Handling

### Transaction Support in NoteRepository

The `NoteRepository` demonstrates full transaction support:

```go
type NoteRepository struct {
    pool *pgxpool.Pool
    tx   pgx.Tx
}
```

#### Transaction Lifecycle

1. **Begin Transaction:**
```go
tx, err := repo.BeginTx(ctx)
if err != nil {
    return err
}
```

2. **Use Transaction:**
```go
txRepo := repo.WithTx(ctx, tx)
```

3. **Execute Operations:**
```go
note, err := txRepo.CreateNote(ctx, title, slug, contents)
if err != nil {
    repo.RollbackTx(ctx, tx)
    return err
}

entityID, err := txRepo.CreateEntity(ctx, name, entityType, description, properties)
if err != nil {
    repo.RollbackTx(ctx, tx)
    return err
}
```

4. **Commit or Rollback:**
```go
if err := repo.CommitTx(ctx, tx); err != nil {
    return err
}
```

#### Transaction Query Routing

The `getQuerier()` method ensures operations use the correct context:

```go
func (r *NoteRepository) getQuerier() db.DBTX {
    if r.tx != nil {
        return r.tx
    }
    return r.pool
}
```

All query methods use `db.New(r.getQuerier())` to automatically route to transaction or pool.

### Transaction Best Practices

1. **Atomic Operations:** Group related changes in single transaction
2. **Error Handling:** Always rollback on error
3. **Context Propagation:** Pass context through all operations
4. **Resource Cleanup:** Use defer for rollback
5. **Short Transactions:** Keep transaction scope minimal

### Example: Complex Transactional Operation

```go
// Create note with entity and relationships atomically
tx, err := noteRepo.BeginTx(ctx)
if err != nil {
    return err
}

txRepo := noteRepo.WithTx(ctx, tx)

// Create note
note, err := txRepo.CreateNote(ctx, title, slug, contents)
if err != nil {
    noteRepo.RollbackTx(ctx, tx)
    return err
}

// Create entity
entityID, err := txRepo.CreateEntity(ctx, name, "person", nil, properties)
if err != nil {
    noteRepo.RollbackTx(ctx, tx)
    return err
}

// Create relationship
err = txRepo.CreateEntityRelationship(ctx, *entityID, "note", note.ID, "mentions", nil)
if err != nil {
    noteRepo.RollbackTx(ctx, tx)
    return err
}

// Create reference
err = txRepo.CreateEntityReference(ctx, entity.EntityReference{
    SourceType: "note",
    SourceID: note.ID,
    EntityID: *entityID,
    ReferenceText: name,
})
if err != nil {
    noteRepo.RollbackTx(ctx, tx)
    return err
}

// Commit all changes
return noteRepo.CommitTx(ctx, tx)
```

---

## pgvector Usage for Embeddings

The application uses pgvector extension for semantic search via vector embeddings.

### Installation

pgvector must be enabled in PostgreSQL:

```sql
CREATE EXTENSION vector;
```

### Vector Type

Vectors are represented as `pgvector.Vector` from `github.com/pgvector/pgvector-go`:

```go
import "github.com/pgvector/pgvector-go"
```

### Creating Vectors

Convert float32 slices to pgvector format:

```go
embedding := []float32{0.1, 0.2, 0.3, ...}
vec := pgvector.NewVector(embedding)
```

### Storing Embeddings

#### Bookmark Embeddings

```go
func (r *BookmarkRepository) CreateEmbeddingChunk(
    ctx context.Context,
    bookmarkID uuid.UUID,
    content, strategy string,
    embedding []float32,
) (uuid.UUID, error) {
    queries := db.New(r.pool)

    // Convert embedding to pgvector.Vector
    embeddingVec := pgvector.NewVector(embedding)
    bookmarkIDPg := pgtype.UUID{Bytes: bookmarkID, Valid: true}

    id, err := queries.CreateEmbeddingChunk(ctx, db.CreateEmbeddingChunkParams{
        BookmarkID: bookmarkIDPg,
        Content:    &content,
        Strategy:   &strategy,
        Column4:    &embeddingVec,
    })
    if err != nil {
        return uuid.Nil, err
    }

    return id, nil
}
```

#### Question Embeddings

```go
func (r *BookmarkRepository) UpdateBookmarkQuestion(
    ctx context.Context,
    content string,
    embedding []float32,
    referenceID, bookmarkID uuid.UUID,
) error {
    queries := db.New(r.pool)

    // Convert embedding to pgvector.Vector
    embeddingVec := pgvector.NewVector(embedding)
    bookmarkIDPg := pgtype.UUID{Bytes: bookmarkID, Valid: true}

    return queries.UpdateBookmarkQuestion(ctx, db.UpdateBookmarkQuestionParams{
        Content:    &content,
        Column2:    &embeddingVec,
        ID:         referenceID,
        BookmarkID: bookmarkIDPg,
    })
}
```

### Searching with Vectors

#### Similarity Search Types

pgvector supports multiple distance operators:

- **Cosine Distance:** `<=>` (used throughout codebase)
- **Euclidean Distance:** `<->`
- **Inner Product:** `<#>`

#### Bookmark Similarity Search

```go
func (r *BookmarkRepository) SearchSimilarBookmarks(
    ctx context.Context,
    embedding []float32,
    strategy string,
    limit int32,
) ([]entity.BookmarkWithTitle, error) {
    queries := db.New(r.pool)

    // Convert embedding to pgvector.Vector
    embeddingVec := pgvector.NewVector(embedding)

    dbBookmarks, err := queries.SearchSimilarBookmarks(ctx, db.SearchSimilarBookmarksParams{
        Strategy: &strategy,
        Column2:  &embeddingVec,
        Limit:    limit,
    })
    if err != nil {
        return nil, err
    }

    // Convert results...
    bookmarks := make([]entity.BookmarkWithTitle, len(dbBookmarks))
    for i, dbBookmark := range dbBookmarks {
        bookmarks[i] = entity.BookmarkWithTitle{
            BookmarkID:   dbBookmark.BookmarkID,
            URL:          dbBookmark.Url,
            CreationDate: dbBookmark.CreationDate.Time,
            Title:        dbBookmark.Title,
            Summary:      dbBookmark.Summary,
        }
    }

    return bookmarks, nil
}
```

#### Raw SQL Vector Search

Example from NoteRepository showing manual query:

```go
func (r *NoteRepository) SearchSimilarNotes(
    ctx context.Context,
    embedding []float32,
    strategy string,
    limit int32,
) ([]entity.NoteListItem, error) {
    query := `
        SELECT
            i.id,
            i.title,
            i.created,
            i.modified,
            array_agg(DISTINCT t.name) FILTER (WHERE t.name IS NOT NULL) as tags
        FROM items i
        INNER JOIN item_semantic_index isi ON i.id = isi.item_id
        LEFT JOIN item_tags it ON i.id = it.item_id
        LEFT JOIN tags t ON it.tag_id = t.id
        WHERE isi.strategy = $1
        GROUP BY i.id, i.title, i.created, i.modified, isi.embedding
        ORDER BY isi.embedding <=> $2::vector
        LIMIT $3
    `

    // Convert []float32 to pgvector.Vector
    vec := pgvector.NewVector(embedding)

    rows, err := r.pool.Query(ctx, query, strategy, vec, limit)
    if err != nil {
        return nil, fmt.Errorf("failed to search similar notes: %w", err)
    }
    defer rows.Close()

    var notes []entity.NoteListItem
    for rows.Next() {
        var note entity.NoteListItem
        var created, modified *int64

        err := rows.Scan(
            &note.ID,
            &note.Title,
            &created,
            &modified,
            &note.Tags,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan note row: %w", err)
        }

        if created != nil {
            note.Created = *created
        }
        if modified != nil {
            note.Modified = *modified
        }
        if note.Tags == nil {
            note.Tags = []string{}
        }

        notes = append(notes, note)
    }

    return notes, rows.Err()
}
```

Key points:
- Use `<=>` operator for cosine distance
- Cast parameter to `::vector` type
- Order by distance (ascending = most similar)
- Group by when joining with other tables

#### Question Similarity Search

```go
func (r *searchRepository) GetSimilarQuestions(
    ctx context.Context,
    embedding []float32,
    limit int32,
) ([]entity.RetrievedItem, error) {
    // Convert embedding to pgvector format
    vec := pgvector.NewVector(embedding)

    rows, err := r.queries.GetSimilarQuestions(ctx, db.GetSimilarQuestionsParams{
        Embedding:   &vec,
        SearchLimit: limit,
    })
    if err != nil {
        return nil, err
    }

    results := make([]entity.RetrievedItem, 0, len(rows))
    for i, row := range rows {
        // Split question into question and answer parts
        question := ""
        answer := ""
        if row.Question != nil {
            parts := strings.SplitN(*row.Question, "\n", 2)
            question = parts[0]
            if len(parts) > 1 {
                answer = parts[1]
            }
        }

        results = append(results, entity.RetrievedItem{
            ID:            i + 1,
            Question:      question,
            Answer:        answer,
            BookmarkID:    row.BookmarkID.String(),
            BookmarkTitle: row.Title,
            BookmarkURL:   row.Url,
            Summary:       row.Summary,
            Similarity:    float64(row.Similarity),
            Strategy:      *row.Strategy,
        })
    }

    return results, nil
}
```

#### Session Embeddings

```go
func (r *SessionRepository) SearchSessionsWithEmbeddings(
    ctx context.Context,
    embedding []float32,
    limit int32,
) ([]entity.SessionSearchResult, error) {
    queries := db.New(r.pool)

    // Convert []float32 to pgvector.Vector
    vec := pgvector.NewVector(embedding)

    results, err := queries.SearchSessionsWithEmbeddings(ctx, db.SearchSessionsWithEmbeddingsParams{
        Column1: &vec,
        Limit:   limit,
    })
    if err != nil {
        return nil, err
    }

    searchResults := make([]entity.SessionSearchResult, len(results))
    for i, r := range results {
        similarity := float64(r.Similarity)
        searchResults[i] = entity.SessionSearchResult{
            SessionID:       r.SessionID,
            RoomID:          r.RoomID,
            FirstDateTime:   r.FirstDateTime.Time,
            LastDateTime:    pgTimestampToTimePtr(r.LastDateTime),
            Summary:         r.Summary,
            DisplayName:     r.DisplayName,
            UserDefinedName: r.UserDefinedName,
            Similarity:      &similarity,
        }
    }
    return searchResults, nil
}
```

#### Item Similarity Search

```go
func (r *ItemRepository) SearchSimilarItems(
    ctx context.Context,
    embedding []float32,
    limit int32,
) ([]entity.ItemListItem, error) {
    queries := db.New(r.pool)

    // Convert []float32 to pgvector.Vector
    vec := pgvector.NewVector(embedding)

    dbItems, err := queries.SearchSimilarItemsByEmbedding(ctx, db.SearchSimilarItemsByEmbeddingParams{
        Column1: &vec,
        Limit:   limit,
    })
    if err != nil {
        return nil, err
    }

    // Convert results...
    items := make([]entity.ItemListItem, 0, len(dbItems))
    for _, dbItem := range dbItems {
        // Convert and append...
    }

    return items, nil
}
```

### Vector Indexing

For optimal performance, create indexes on vector columns:

```sql
-- IVFFlat index (faster build, approximate search)
CREATE INDEX ON bookmark_embeddings
USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);

-- HNSW index (slower build, better recall)
CREATE INDEX ON bookmark_embeddings
USING hnsw (embedding vector_cosine_ops);
```

### Embedding Strategies

The codebase uses multiple embedding strategies:

- **"summary"** - Embeddings of bookmark summaries
- **"full-content"** - Embeddings of full text content
- **"question"** - Embeddings of Q&A pairs
- **"chunk"** - Embeddings of content chunks

Different strategies allow for strategy-specific searches:

```go
// Search using summary embeddings
results, err := repo.SearchSimilarBookmarks(ctx, queryEmbedding, "summary", 10)

// Search using full content embeddings
results, err := repo.SearchSimilarBookmarks(ctx, queryEmbedding, "full-content", 10)
```

### Best Practices

1. **Normalize Embeddings:** Ensure vectors are normalized before storage if using cosine similarity
2. **Strategy Consistency:** Use same strategy for indexing and querying
3. **Batch Operations:** Insert embeddings in batches for performance
4. **Index Management:** Create appropriate indexes based on data size and query patterns
5. **Error Handling:** Handle dimension mismatches and invalid vectors gracefully
6. **Version Control:** Store embedding model version with strategy to enable model updates

---

## Summary

The PostgreSQL adapter provides:

- **Robust Connection Management:** Connection pooling with pgxpool
- **Repository Pattern:** Clean separation of data access logic
- **Type Safety:** sqlc-generated queries with compile-time safety
- **Transaction Support:** Full transaction lifecycle management
- **Vector Search:** pgvector integration for semantic search
- **Type Conversions:** Comprehensive converter utilities
- **Flexible Queries:** Combination of generated and raw SQL
- **Batch Operations:** Optimized queries to avoid N+1 problems
- **JSON Support:** Native JSONB handling for flexible schemas

The adapter serves as the foundation for the Garden application's data persistence layer, providing reliable, performant, and type-safe database operations.
