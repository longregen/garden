# Garden

Garden is a personal knowledge management system written in Go. It aggregates, stores, and serves data from multiple sources—bookmarks, notes, chat messages, contacts, and social media—through a unified REST API. The system uses vector embeddings for semantic search and maintains a knowledge graph to connect related concepts across different data types.

## Architecture

The application follows hexagonal architecture (ports and adapters), separating business logic from external concerns:

```
cmd/server/main.go          → Application composition and dependency injection
internal/domain/entity/     → Pure data structures (no dependencies)
internal/domain/service/    → Business logic implementing use case interfaces
internal/port/input/        → Use case interfaces (what the app does)
internal/port/output/       → Repository and service interfaces (what the app needs)
internal/adapter/primary/   → HTTP handlers (Chi router)
internal/adapter/secondary/ → PostgreSQL, Ollama, HTTP fetching, social APIs
```

## Data Storage

All data is stored in PostgreSQL 17+ with several extensions:

| Extension | Purpose |
|-----------|---------|
| **pgvector** | Vector similarity search using HNSW indexes |
| **pg_trgm** | Trigram-based fuzzy text matching |
| **fuzzystrmatch** | String distance calculations (Levenshtein) |
| **unaccent** | Text normalization for search |
| **uuid-ossp** | UUID generation |

SQL queries are type-checked at compile time using [sqlc](https://sqlc.dev/).

---

## Data Model

### Bookmarks

Bookmarks store web pages with progressive content enrichment:

```
bookmarks
├── bookmark_id (UUID, primary key)
├── url (TEXT)
└── creation_date (TIMESTAMP)

bookmark_http_responses     → Raw HTTP response (status, headers, body)
bookmark_titles             → Extracted page title
processed_contents          → Text extraction via Lynx or go-readability
bookmark_content_references → Chunked content with vector embeddings
bookmark_questions          → Generated Q&A pairs for retrieval
bookmark_category           → Category assignment
bookmark_sources            → Origin tracking (e.g., which message shared this URL)
bookmark_evaluations        → Quality scores
```

Content is processed in stages: fetch HTTP → extract text → chunk content → generate embeddings. Each stage stores its output separately, allowing partial processing and reprocessing.

**Embedding strategies** determine how content is chunked and vectorized:

- `chunked-reader` — Fixed-size chunks from readable content
- `qa-v2-passage` — Question-answer pair embeddings
- `summary-reader` — Summary-based embeddings

### Notes

Markdown notes with tagging and entity linking:

```
notes
├── id (UUID)
├── title (TEXT)
├── slug (TEXT)
├── contents (TEXT, markdown)
├── created (BIGINT, unix timestamp)
└── modified (BIGINT, unix timestamp)

note_tags → Many-to-many relationship with tags table
```

Notes support wiki-style entity references using `[[entity name]]` syntax. These references are parsed and stored in `entity_references`, creating bidirectional links between notes and the knowledge graph.

### Knowledge Graph

Entities represent named concepts that can be referenced across the system:

```
entities
├── entity_id (UUID)
├── name (TEXT)
├── type (TEXT) — person, place, concept, group_chat, etc.
├── description (TEXT)
├── properties (JSONB)
├── created_at, updated_at, deleted_at (soft delete)

entity_relationships
├── entity_id → source entity
├── related_type (TEXT) → target type (contact, room, note, etc.)
├── related_id (UUID) → target record
├── relationship_type (TEXT) → identity, mentions, related_to, etc.
└── metadata (JSONB)

entity_references
├── source_type (TEXT) → note, message, etc.
├── source_id (UUID)
├── entity_id (UUID)
├── reference_text (TEXT) — the original [[bracketed]] text
└── position (INT)
```

Entities bridge different data types. A `person` entity might link to:
- A `contact` record (relationship_type: `identity`)
- Multiple `message` records where they're mentioned
- `note` records that reference them

Database triggers automatically create entities when certain records are inserted (e.g., contacts automatically get a corresponding `person` entity).

### Messages and Sessions

Messages come from Matrix chat rooms and are automatically grouped into sessions:

```
messages
├── message_id (UUID)
├── event_id (TEXT) — Matrix event ID
├── room_id (UUID)
├── sender_contact_id (UUID)
├── event_datetime (TIMESTAMP)
├── body, formatted_body (TEXT)
├── msgtype (TEXT) — m.text, m.image, m.audio, etc.
├── message_type, message_classification (TEXT)
├── is_edited, is_reply (BOOLEAN)
└── reply_to_event_id (TEXT)

sessions
├── session_id (UUID)
├── room_id (UUID)
├── first_message_id, last_message_id (UUID)
├── first_date_time, last_date_time (TIMESTAMP)

session_message → Links messages to sessions
session_summaries → AI-generated session summaries
```

**Automatic session grouping**: A database trigger groups messages into sessions based on time gaps. When a message arrives more than 2 hours after the last message in a room, a new session begins. This happens transparently via the `add_message_to_session` stored procedure.

Supporting tables for messages:
- `messages_edit_history` — Previous versions of edited messages
- `messages_media` — Attached files with Matrix content URIs
- `messages_mentions` — Extracted @mentions
- `messages_reactions` — Emoji reactions
- `messages_relations` — Reply threading
- `message_text_representation` — Searchable text with GIN indexes

### Rooms

Chat rooms with participant tracking:

```
rooms
├── room_id (UUID)
├── matrix_room_id (TEXT)
├── display_name, user_defined_name (TEXT)
├── avatar (TEXT)
├── is_direct (BOOLEAN)
├── is_space (BOOLEAN)
├── participant_count (INT)
└── last_activity (TIMESTAMP)

room_participants
├── room_id, contact_id (UUID)
└── known_last_presence (TIMESTAMP)

room_known_names → Historical room names
room_known_avatars → Historical avatars
room_state → Matrix room state events (JSONB)
```

### Contacts

People from chat systems with metadata and evaluation:

```
contacts
├── contact_id (UUID)
├── name (TEXT)
├── display_name (TEXT)
├── source (TEXT)
├── avatar (TEXT)
└── created_at, updated_at (TIMESTAMP)

contact_evals → Subjective scores (importance, closeness, fondness)
contact_sources → External identifiers (Matrix user ID, etc.)
contact_known_names → Alternative names
contact_known_avatars → Historical avatars
contact_tags → Many-to-many with contact_tagnames
contact_stats → Aggregated message counts
```

### Supporting Data

**Categories and Tags**:
```
categories → Hierarchical organization for bookmarks
tags → Freeform labels for items and notes
contact_tagnames → Separate tag namespace for contacts
```

**Browser History**:
```
browser_history
├── id (UUID)
├── url, title (TEXT)
├── visit_time (TIMESTAMP)
└── browser (TEXT)
```

**Social Posts**:
```
social_posts
├── social_post_id (UUID)
├── content (TEXT)
├── twitter_id, bluesky_id (TEXT)
└── created_at, posted_at (TIMESTAMP)
```

**Observations** — System events and extracted insights:
```
observations
├── observation_id (UUID)
├── type (TEXT) — transcription, analysis, etc.
├── source (TEXT)
├── ref (UUID) — Reference to source record
├── data (JSONB)
└── tags (TEXT[])
```

**Configurations** — Key-value settings:
```
configurations
├── key (TEXT, primary key)
├── value (JSONB)
└── created_at, updated_at (TIMESTAMP)
```

---

## API Structure

The HTTP API is organized by resource type. All endpoints accept and return JSON.

### Bookmarks
```
GET    /api/bookmarks              → List with filters (category, date range, search)
GET    /api/bookmarks/{id}         → Full details with processed content
GET    /api/bookmarks/search       → Vector similarity search
GET    /api/bookmarks/random       → Redirect to random bookmark URL
GET    /api/bookmarks/missing/http → Bookmarks without fetched content
GET    /api/bookmarks/missing/reader → Bookmarks without readable content
POST   /api/bookmarks/{id}/fetch   → Fetch HTTP content from URL
POST   /api/bookmarks/{id}/process/lynx → Extract text via Lynx
POST   /api/bookmarks/{id}/process/reader → Extract via readability
POST   /api/bookmarks/{id}/embeddings → Generate vector embeddings
POST   /api/bookmarks/{id}/summary-embedding → Generate summary + embedding
GET    /api/bookmarks/{id}/title   → Extract title from content
PUT    /api/bookmarks/{id}/question → Add/update Q&A pair
DELETE /api/bookmarks/{id}/question/{refId} → Remove Q&A pair
```

### Notes
```
GET    /api/notes        → List all notes with tags
GET    /api/notes/{id}   → Full note with processed entity references
POST   /api/notes        → Create note with title, contents, tags
PUT    /api/notes/{id}   → Update note
DELETE /api/notes/{id}   → Delete note
GET    /api/notes/tags   → List all tags
```

### Entities
```
GET    /api/entities                    → List entities (filterable by type)
GET    /api/entities/{id}               → Get entity with relationships
POST   /api/entities                    → Create entity
PUT    /api/entities/{id}               → Update entity
DELETE /api/entities/{id}               → Soft delete
GET    /api/entities/{id}/relationships → List relationships
POST   /api/entities/{id}/references    → Create entity reference
POST   /api/entities/parse-references   → Parse [[entity]] syntax from text
```

### Sessions and Messages
```
GET    /api/sessions                    → List sessions for a room
GET    /api/sessions/{id}/messages      → Messages in session with contacts
POST   /api/sessions/{id}/search        → Search sessions by content
GET    /api/timeline                    → Aggregated timeline visualization
GET    /api/messages                    → Search messages
POST   /api/messages                    → Create message
```

### Contacts
```
GET    /api/contacts              → List with filters
GET    /api/contacts/{id}         → Full contact with relations
POST   /api/contacts              → Create contact
PUT    /api/contacts/{id}         → Update contact
DELETE /api/contacts/{id}         → Delete contact
GET    /api/contacts/{id}/tags    → Get contact's tags
POST   /api/contacts/{id}/tags/{name} → Add tag
DELETE /api/contacts/{id}/tags/{name} → Remove tag
PATCH  /api/contacts/{id}/merge/{targetId} → Merge contacts
POST   /api/contacts/batch        → Batch update
POST   /api/contacts/refresh-stats → Recalculate statistics
```

### Search
```
POST   /api/search/all      → Unified search across content types
POST   /api/search/advanced → LLM-powered contextual search
```

### Other Resources
```
/api/categories      → CRUD for bookmark categories
/api/tags            → CRUD for tags
/api/items           → Generic items with tags
/api/browser-history → Browser history records
/api/social-posts    → Social media posting (Twitter, Bluesky)
/api/observations    → System observations
/api/configuration   → Key-value configuration
/api/dashboard       → Aggregated statistics
/api/rooms           → Chat rooms
```

---

## External Services

### Ollama

The system uses Ollama for:

- **Embeddings**: Converts text to vectors using `nomic-embed-text` (or configurable model). Text is chunked into ~8000 character segments before embedding.
- **Summarization**: Generates concise summaries of bookmark content.
- **LLM queries**: Powers advanced search with contextual understanding.

### Content Processing

Two strategies for extracting readable content from web pages:

- **Lynx**: Shell out to the Lynx browser for text rendering
- **go-readability**: Pure Go implementation of Mozilla's Readability algorithm

### Social Media

Optional integration with Twitter and Bluesky APIs for cross-posting.

---

## Data Flow Examples

### Bookmark Ingestion

1. Bookmark created with URL
2. `POST /fetch` retrieves HTTP response → stored in `http_responses`
3. `POST /process/reader` extracts article → stored in `processed_contents`
4. `POST /embeddings` chunks content and generates vectors → stored in `bookmark_content_references`
5. `POST /summary-embedding` creates summary → stored as `summary-reader` strategy

Each step is independent and idempotent. Failed steps can be retried without affecting earlier stages.

### Message → Session Flow

1. Message inserted into `messages` table
2. Trigger `process_new_message_into_session_trigger` fires
3. Stored procedure checks last session for this room
4. If within 2-hour gap: add to existing session
5. If gap exceeded: create new session
6. Session's `last_message_id` and `last_date_time` updated

### Entity Reference Resolution

1. Note created with content containing `[[John Smith]]`
2. Service parses content for `[[...]]` patterns
3. For each reference, looks up or creates matching entity
4. Creates `entity_reference` linking note → entity
5. When note is retrieved, references are converted to markdown links

---

## Data Relationships Diagram

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│  bookmarks  │────▶│ http_resp.   │     │   notes     │
│             │────▶│ proc_content │     │             │
│             │────▶│ content_refs │◀────│             │
│             │────▶│ questions    │     │             │◀───┐
│             │────▶│ titles       │     │             │    │
└─────────────┘     └──────────────┘     └──────▲──────┘    │
       │                                        │           │
       │ category_id                   note_tags│           │
       ▼                                        │           │
┌─────────────┐                          ┌──────┴──────┐    │
│ categories  │                          │    tags     │    │
└─────────────┘                          └─────────────┘    │
                                                            │
┌─────────────┐     ┌──────────────┐     ┌─────────────┐    │
│  contacts   │────▶│ contact_evals│     │  entities   │◀───┘
│             │────▶│ known_names  │◀────│             │  entity_refs
│             │────▶│ sources      │     │             │────────────┐
│             │◀────│              │     │             │            │
└──────┬──────┘     └──────────────┘     └──────┬──────┘            │
       │                                        │                   │
       │ sender_contact_id              entity_relationships        │
       │                                        │                   │
       ▼                                        ▼                   │
┌─────────────┐     ┌──────────────┐     ┌─────────────┐            │
│  messages   │────▶│ edit_history │     │   rooms     │            │
│             │────▶│ media        │◀────│             │            │
│             │────▶│ mentions     │     │             │────────────┘
│             │────▶│ reactions    │     │ participants│
└──────┬──────┘     └──────────────┘     └─────────────┘
       │
       │ session_message
       ▼
┌─────────────┐     ┌──────────────┐
│  sessions   │────▶│  summaries   │
└─────────────┘     └──────────────┘
```

The knowledge graph (`entities` + `entity_relationships`) acts as a hub connecting disparate data types. A person entity can link to their contact record, messages mentioning them, notes about them, and rooms they participate in—enabling queries like "show everything related to John Smith" across the entire system.
