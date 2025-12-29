# Domain Entities Documentation

## Overview

The Garden project implements a comprehensive domain model for managing various types of personal information, including bookmarks, notes, contacts, messages, and more. The domain layer follows Domain-Driven Design (DDD) principles with entities representing core business concepts, value objects for data structures, and clear separation of concerns.

The domain model is organized around several key areas:
- **Knowledge Management**: Bookmarks, notes, items, and entities with semantic relationships
- **Communication**: Messages, rooms, sessions, and contacts
- **Organization**: Categories, tags, and entity references
- **External Integrations**: Browser history, Logseq pages, and social media posts
- **System Management**: Configuration, dashboard statistics, and search capabilities

## Common Patterns

### 1. UUID-based Identification
All primary entities use `uuid.UUID` for unique identification, providing globally unique identifiers across the system.

```go
type Bookmark struct {
    BookmarkID   uuid.UUID
    // ...
}
```

### 2. Timestamps
Most entities include timestamp fields for audit trails:
- `CreatedAt`, `UpdatedAt` - Using `time.Time` for RFC3339 timestamps
- `Created`, `Modified` - Using `int64` Unix timestamps (legacy pattern)
- `CreationDate` - Using `time.Time` for creation tracking

### 3. Nullable Fields
Extensive use of pointers for optional fields, allowing proper NULL representation in the database:
```go
Title        *string
Summary      *string
UpdatedAt    *time.Time
```

### 4. Input DTOs
Separate input structs for create and update operations, following the Command pattern:
- `Create*Input` - For entity creation with required fields
- `Update*Input` - For updates with all fields optional (pointers)

### 5. Rich Data Structures
Complex entities include companion types for different contexts:
- Base entity (core data)
- `*Details` (with relations loaded)
- `*WithTitle` or `*With*` (projections with specific fields)
- `Full*` (complete with all relations)

### 6. Filters and Search
Dedicated filter types for query operations:
```go
type BookmarkFilters struct {
    CategoryID        *uuid.UUID
    SearchQuery       *string
    Page              int32
    Limit             int32
}
```

### 7. JSON Raw Messages
Use of `json.RawMessage` for flexible, schema-less data storage:
```go
Properties  json.RawMessage  // Flexible property bag
Metadata    json.RawMessage  // Extensible metadata
```

## Core Entities

### Entity

**File**: `/home/user/garden/internal/domain/entity/entity.go`

The `Entity` type (defined in `note.go`) represents a node in the knowledge graph system.

```go
type Entity struct {
    EntityID    uuid.UUID
    Name        string
    Type        string
    Description *string
    Properties  json.RawMessage
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   *time.Time
}
```

**Purpose**: Represents named entities in the knowledge graph (people, places, concepts, etc.) that can be referenced across notes and other content.

**Key Fields**:
- `EntityID` - Unique identifier
- `Name` - Display name of the entity
- `Type` - Entity classification (person, place, concept, etc.)
- `Description` - Optional description
- `Properties` - Flexible JSON storage for type-specific properties
- `DeletedAt` - Soft deletion timestamp

**Related Types**:
- `CreateEntityInput` - Input for creating new entities
- `UpdateEntityInput` - Input for updating entities (all fields optional)

### EntityRelationship

Represents typed relationships between entities and other resources.

```go
type EntityRelationship struct {
    ID               uuid.UUID
    EntityID         uuid.UUID
    RelatedType      string
    RelatedID        uuid.UUID
    RelationshipType string
    Metadata         json.RawMessage
    CreatedAt        time.Time
    UpdatedAt        time.Time
}
```

**Purpose**: Models bidirectional relationships in the knowledge graph, allowing entities to be connected to other entities or resources with typed relationships.

**Key Fields**:
- `RelatedType` - Type of the related resource
- `RelationshipType` - Nature of the relationship (e.g., "mentions", "related_to")
- `Metadata` - Additional relationship-specific data

### EntityReference

**File**: `/home/user/garden/internal/domain/entity/entity_reference.go`

Tracks references to entities within content (notes, messages, etc.).

```go
type EntityReference struct {
    ID            uuid.UUID
    SourceType    string
    SourceID      uuid.UUID
    EntityID      uuid.UUID
    ReferenceText string
    Position      *int
    CreatedAt     time.Time
}
```

**Purpose**: Links content to entities, enabling backlinks and knowledge graph connections. Supports wiki-style `[[entity]]` references.

**Key Fields**:
- `SourceType` - Type of content containing the reference (note, message, etc.)
- `SourceID` - ID of the source content
- `ReferenceText` - The actual text used in the reference
- `Position` - Optional position in the source content

**Related Types**:
```go
type ParsedReference struct {
    Original    string
    EntityName  string
    DisplayText string
}
```

## Knowledge Management Entities

### Bookmark

**File**: `/home/user/garden/internal/domain/entity/bookmark.go`

Represents saved web pages with multiple levels of detail.

**Base Type**:
```go
type Bookmark struct {
    BookmarkID   uuid.UUID
    URL          string
    CreationDate time.Time
}
```

**Enhanced Types**:
- `BookmarkWithTitle` - Adds title and summary
- `BookmarkDetails` - Complete bookmark with all fetched content, categories, and Q&A
- `BookmarkContentReference` - Content chunks with embeddings for semantic search

**Purpose**: Manages saved URLs with rich metadata, content extraction, and AI-powered Q&A capabilities.

**Key Features**:
- HTTP content fetching and storage
- Multiple content representations (raw, Lynx text, reader mode)
- Question/answer pairs for semantic search
- Content chunking with embeddings
- Category association

**Related Types**:
- `BookmarkQuestion` - Q&A pairs associated with bookmarks
- `BookmarkFilters` - Query filtering options
- `ProcessingResult` - Results from processing operations
- `FetchResult` - HTTP fetch results
- `EmbeddingResult` - Embedding generation results
- `UpdateQuestionInput`, `DeleteQuestionInput` - Q&A management

### Note

**File**: `/home/user/garden/internal/domain/entity/note.go`

Personal notes with markdown support and entity references.

```go
type Note struct {
    ID       uuid.UUID
    Title    *string
    Slug     *string
    Contents *string
    Created  int64
    Modified int64
}
```

**Purpose**: Stores personal notes with wiki-style entity linking and tagging.

**Key Features**:
- Markdown content support
- Entity reference processing
- Tag associations
- URL-friendly slugs

**Related Types**:
- `FullNote` - Note with tags, entity ID, and processed content (entity references converted to markdown links)
- `NoteTag` - Tags associated with notes
- `NoteListItem` - Lightweight projection for list views
- `CreateNoteInput`, `UpdateNoteInput` - CRUD operations

### Item

**File**: `/home/user/garden/internal/domain/entity/item.go`

Generic content items (appears to be legacy/alternative to Note).

```go
type Item struct {
    ID       uuid.UUID
    Title    *string
    Slug     *string
    Contents *string
    Created  int64
    Modified int64
}
```

**Purpose**: Similar to Note, provides a generic content storage mechanism.

**Related Types**:
- `FullItem` - Item with tag list
- `ItemTag` - Tags associated with items
- `ItemListItem` - List view projection
- `CreateItemInput`, `UpdateItemInput` - CRUD operations

**Note**: Item and Note share very similar structures; this may indicate a migration or different use cases.

## Communication Entities

### Contact

**File**: `/home/user/garden/internal/domain/entity/contact.go`

Represents people in the system.

```go
type Contact struct {
    ContactID        uuid.UUID
    Name             string
    Email            *string
    Phone            *string
    Birthday         *string
    Notes            *string
    Extras           map[string]interface{}
    CreationDate     time.Time
    LastUpdate       time.Time
    LastWeekMessages *int32
    GroupsInCommon   *int32
}
```

**Purpose**: Manages contact information with relationship metrics and multi-source integration.

**Key Features**:
- Personal information storage
- Evaluation metrics (importance, closeness, fondness)
- Multiple known names
- External source references
- Tag associations
- Room memberships
- Activity tracking

**Related Types**:
- `ContactEvaluation` - Subjective relationship metrics
- `ContactTag` - Tag associations
- `ContactKnownName` - Alternative names/aliases
- `ContactSource` - External service references (Matrix, etc.)
- `ContactRoom` - Room/group memberships
- `FullContact` - Complete contact with all relations
- `ContactTagName` - Tag definitions
- `BatchUpdateResult` - Batch operation results
- Various input types for CRUD operations

### Message

**File**: `/home/user/garden/internal/domain/entity/message.go`

Individual messages from chat systems.

```go
type Message struct {
    MessageID       uuid.UUID
    SenderContactID uuid.UUID
    RoomID          uuid.UUID
    EventID         string
    EventDatetime   time.Time
    Body            *string
    FormattedBody   *string
    MessageType     *string
    SenderName      *string
    SenderEmail     *string
}
```

**Purpose**: Stores messages from external chat systems (likely Matrix) with full-text search support.

**Related Types**:
- `MessageSearchParams` - Search filtering options
- `MessageTextRepresentation` - Searchable text with vectors

### Room

**File**: `/home/user/garden/internal/domain/entity/room.go`

Chat rooms or group conversations.

```go
type Room struct {
    RoomID          uuid.UUID
    DisplayName     *string
    UserDefinedName *string
    SourceID        string
    LastActivity    *time.Time
    LastMessageTime *time.Time
    ParticipantCount int32
}
```

**Purpose**: Represents chat rooms with participant tracking and message history.

**Related Types**:
- `RoomParticipant` - Participant information with presence tracking
- `RoomDetails` - Room with participants, messages, and contacts
- `RoomMessage` - Enriched message with classifications and bookmarks
- `SearchResult` - Message search results with ranking
- `MessagesWithContacts` - Bundled messages and contact info

### Session

**File**: `/home/user/garden/internal/domain/entity/session.go`

Conversation groupings within rooms.

```go
type Session struct {
    SessionID     uuid.UUID
    RoomID        uuid.UUID
    FirstDateTime time.Time
    FirstMessageID *uuid.UUID
    LastMessageID  *uuid.UUID
    LastDateTime   *time.Time
    CreatedAt      time.Time
    UpdatedAt      *time.Time
}
```

**Purpose**: Groups messages into conversation sessions for better organization and analysis.

**Key Features**:
- Automatic session detection
- Summary generation
- Participant activity tracking
- Timeline visualization
- Semantic search via embeddings

**Related Types**:
- `SessionSummary` - Session with AI-generated summary
- `SessionSearchResult` - Search results with similarity scores
- `SessionMessage` - Message in session context
- `SessionMessageContact` - Contact information for senders
- `SessionMessagesResponse` - Messages with contacts
- `ParticipantActivity` - Per-participant statistics
- `SessionWithMessages` - Session with message count
- `TimelineMonth`, `TimelineSession`, `TimelineParticipant` - Timeline visualization
- `TimelineData` - Complete timeline data structure

## Organization Entities

### Category

**File**: `/home/user/garden/internal/domain/entity/category.go`

Classification categories for bookmarks.

```go
type Category struct {
    CategoryID uuid.UUID
    Name       string
}
```

**Purpose**: Organizes bookmarks into named categories with external source support.

**Related Types**:
- `CategorySource` - Source data for categories (RSS feeds, etc.)
- `CategoryWithSources` - Category with all sources
- `CreateCategorySourceInput`, `UpdateCategorySourceInput` - Source management
- `MergeCategoriesInput` - Category merging operations

### Tag

**File**: `/home/user/garden/internal/domain/entity/tag.go`

Flexible tagging system for items and notes.

```go
type Tag struct {
    ID           uuid.UUID
    Name         string
    Created      int64
    Modified     int64
    LastActivity *int64
}
```

**Purpose**: Provides folksonomy-style organization with activity tracking.

**Related Types**:
- `TagWithUsage` - Tag with usage count
- `AddTagInput`, `RemoveTagInput` - Tag operations

## Integration Entities

### BrowserHistory

**File**: `/home/user/garden/internal/domain/entity/browser_history.go`

Browser history import from Firefox.

```go
type BrowserHistory struct {
    ID                         int32
    URL                        string
    Title                      *string
    VisitDate                  *time.Time
    Typed                      *bool
    Hidden                     *bool
    ImportedFromFirefoxPlaceID *int32
    ImportedFromFirefoxVisitID *int32
    Domain                     *string
    CreatedAt                  *time.Time
}
```

**Purpose**: Imports and analyzes browser history for insights and bookmark suggestions.

**Related Types**:
- `BrowserHistoryFilters` - Query filtering
- `DomainVisitCount` - Aggregated statistics by domain

### Logseq

**File**: `/home/user/garden/internal/domain/entity/logseq.go`

Integration with Logseq knowledge management system.

```go
type LogseqPage struct {
    Path         string
    Filename     string
    Title        string
    Content      string
    Frontmatter  LogseqPageFrontmatter
    LastModified time.Time
}
```

**Purpose**: Syncs entities with Logseq markdown pages for external editing.

**Key Features**:
- YAML frontmatter parsing
- Bidirectional sync between database and git repository
- Conflict detection
- Force update capabilities

**Related Types**:
- `LogseqPageFrontmatter` - YAML metadata
- `SyncStats` - Sync operation statistics
- `SyncCheckResult` - Hard sync check results
- `OutOfSyncItem` - Items needing reconciliation
- `ForceUpdateRequest` - Force sync request

### SocialPost

**File**: `/home/user/garden/internal/domain/entity/social_post.go`

Cross-posting to social media platforms.

```go
type SocialPost struct {
    PostID         uuid.UUID
    Content        string
    TwitterPostID  *string
    BlueskyPostID  *string
    CreatedAt      time.Time
    UpdatedAt      *time.Time
    Status         string
    ErrorMessage   *string
}
```

**Purpose**: Manages posting to Twitter and Bluesky with credential management and OAuth flow.

**Related Types**:
- `SocialPostFilters` - Query filtering
- `CreateSocialPostInput` - Post creation
- `UpdateStatusInput` - Status updates
- `PostResult` - Posting results
- `CredentialsStatus` - Credential validation status
- `TwitterCredentials`, `BlueskyCredentials` - Platform-specific credentials
- `TwitterProfile`, `BlueskyProfile` - User profiles
- `TwitterTokens` - OAuth tokens
- `TwitterAuthURL`, `TwitterCallbackInput` - OAuth flow
- `TwitterOAuthState` - OAuth state management

## System Entities

### Configuration

**File**: `/home/user/garden/internal/domain/entity/configuration.go`

Key-value configuration storage.

```go
type Configuration struct {
    ConfigID  uuid.UUID
    Key       string
    Value     string
    IsSecret  bool
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

**Purpose**: Stores application configuration with secret protection.

**Related Types**:
- `NewConfiguration` - Configuration creation
- `ConfigurationUpdate` - Configuration updates
- `ConfigurationFilter` - Query filtering with secret inclusion control

### Dashboard

**File**: `/home/user/garden/internal/domain/entity/dashboard.go`

Dashboard statistics and metrics.

```go
type DashboardStats struct {
    Contacts    CategoryStats
    Sessions    CategoryStats
    Bookmarks   CategoryStats
    History     CategoryStats
    RecentItems []RecentItem
}
```

**Purpose**: Provides aggregated statistics for the dashboard UI.

**Related Types**:
- `CategoryStats` - Per-category statistics with growth metrics
- `RecentItem` - Recent activity across all categories

### Search

**File**: `/home/user/garden/internal/domain/entity/search.go`

Unified search across entity types.

**Types**:
- `UnifiedSearchResult` - Search results from multiple tables
- `SearchWeights` - Configurable scoring weights
- `RetrievedItem` - Bookmark Q&A search results
- `AdvancedSearchResult` - LLM-powered search with reasoning

**Purpose**: Provides semantic and keyword search across all entity types with configurable ranking.

**Key Features**:
- Vector similarity search
- Exact match boosting
- Recency weighting
- LLM-powered advanced search with thinking process

### Observation

**File**: `/home/user/garden/internal/domain/entity/observation.go`

Generic data collection mechanism.

```go
type Observation struct {
    ObservationID uuid.UUID
    Data          json.RawMessage
    Type          *string
    Source        *string
    Tags          *string
    Parent        *uuid.UUID
    Ref           *uuid.UUID
    CreationDate  int64
}
```

**Purpose**: Stores arbitrary observations with flexible schema for experimentation and data collection.

**Related Types**:
- `FeedbackType` - Enum for feedback types (upvote, downvote, trash)
- `FeedbackData` - Q&A feedback structure
- `StoreFeedbackInput` - Feedback submission
- `FeedbackStats` - Aggregated feedback statistics

### Utility

**File**: `/home/user/garden/internal/domain/entity/utility.go`

System utilities and debug information.

```go
type DebugInfo struct {
    DatabaseStatus string
    Version        string
    Uptime         string
    Config         map[string]string
}
```

**Purpose**: Provides system health and debugging information.

## Entity Relationships

### Primary Relationships

1. **Bookmark → Category** (Many-to-One)
   - Bookmarks are organized into categories
   - Categories can contain multiple bookmarks

2. **Bookmark → BookmarkQuestion** (One-to-Many)
   - Each bookmark can have multiple Q&A pairs
   - Questions are used for semantic search

3. **Contact → ContactTag** (Many-to-Many)
   - Contacts can have multiple tags
   - Tags can be applied to multiple contacts

4. **Contact → Room** (Many-to-Many via RoomParticipant)
   - Contacts participate in rooms
   - Rooms have multiple participants

5. **Room → Message** (One-to-Many)
   - Rooms contain messages
   - Messages belong to one room

6. **Room → Session** (One-to-Many)
   - Rooms are divided into sessions
   - Sessions group related messages

7. **Contact → Message** (One-to-Many)
   - Contacts send messages
   - Messages have a sender

8. **Note/Item → Tag** (Many-to-Many)
   - Notes and items can have multiple tags
   - Tags can be applied to multiple items

9. **Entity → EntityReference** (One-to-Many)
   - Entities can be referenced in multiple places
   - Each reference points to one entity

10. **Entity → EntityRelationship** (Many-to-Many)
    - Entities can have relationships with other resources
    - Relationships are typed and bidirectional

### Cross-Entity Relationships

- **Messages → Bookmarks**: Messages can reference bookmarks
- **Notes → Entities**: Notes can reference entities via `[[entity]]` syntax
- **Contacts → ContactSource**: Contacts linked to external systems (Matrix, etc.)
- **Entities → Logseq Pages**: Entities synced with Logseq pages

## Value Objects and Enums

### Timestamps
- `time.Time` - RFC3339 timestamps (modern pattern)
- `int64` - Unix timestamps (legacy pattern)

### FeedbackType
```go
const (
    FeedbackUpvote   FeedbackType = "upvote"
    FeedbackDownvote FeedbackType = "downvote"
    FeedbackTrash    FeedbackType = "trash"
)
```

### SearchWeights
```go
type SearchWeights struct {
    ExactMatchWeight float64  // Default: 5.0
    SimilarityWeight float64  // Default: 2.0
    RecencyWeight    float64  // Default: 1.0
}
```

### Embeddings
```go
type Embedding struct {
    Text      string
    Embedding []float32  // Vector representation
}
```

### JSON Structures
- `json.RawMessage` - Used for flexible, schema-less storage in:
  - Entity.Properties
  - EntityRelationship.Metadata
  - Category.RawSource
  - Observation.Data

### Map Types
- `map[string]interface{}` - Used for:
  - Contact.Extras
  - LogseqPageFrontmatter.Extra
  - RoomMessage.TranscriptionData

## Design Patterns Summary

1. **Rich Domain Model**: Entities encapsulate both data and behavior context
2. **DTO Pattern**: Separate input/output types for API boundaries
3. **Repository Pattern**: Entities designed for repository implementation
4. **Soft Deletion**: DeletedAt timestamps for reversible deletes
5. **Audit Trail**: Created/Updated timestamps on most entities
6. **Flexible Schema**: JSON fields for extensibility without migrations
7. **Semantic Search**: Vector embeddings for AI-powered search
8. **Multi-tenancy Ready**: UUID-based identification supports distributed systems
9. **External Integration**: Source tracking for imported data
10. **Event Sourcing Ready**: Observation pattern for generic data collection

## Best Practices

1. **Always use pointers for optional fields** to distinguish NULL from zero values
2. **Use UUID v4 for all entity IDs** for global uniqueness
3. **Include timestamps** on all entities for audit trails
4. **Separate create/update inputs** with required vs. optional fields
5. **Use json.RawMessage** for truly dynamic data, not for domain model
6. **Provide list projections** for efficient queries
7. **Use semantic versioning** for breaking changes to entity structures
8. **Document relationships** explicitly in related types
9. **Use enums (constants)** for fixed value sets
10. **Consider timezone handling** - store UTC, convert on display
