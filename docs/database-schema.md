# Database Schema Documentation

## Overview

This document provides comprehensive documentation for the PostgreSQL database schema used in the Garden project. The database uses PostgreSQL 17.7 with several extensions to support advanced features including vector embeddings, full-text search, and fuzzy string matching.

## Database Extensions

The schema leverages the following PostgreSQL extensions:

- **vector**: Provides vector data type and HNSW/IVFFlat access methods for semantic search with embeddings
- **pg_trgm**: Text similarity measurement and trigram-based index searching
- **fuzzystrmatch**: Determines similarities and distance between strings
- **unaccent**: Text search dictionary that removes accents
- **uuid-ossp**: Generates universally unique identifiers (UUIDs)

## Table Groups

The database schema is organized into several functional domains:

1. **Messaging System** - Messages, rooms, sessions, and related data
2. **Contacts & Social** - Contact management, tags, and social relationships
3. **Bookmarks & Content** - Web bookmarks, content processing, and browser history
4. **Knowledge Management** - Items (notes), tags, and semantic indexing
5. **Entities & Relationships** - Generic entity system for cross-domain relationships
6. **Dispatch System** - Audio transcription and analysis workflow
7. **AI Conversations** - Alicia conversation tracking
8. **Configuration & Observations** - System configuration and data observations
9. **Social Posts** - Cross-platform social media posting

---

## 1. Messaging System

### messages

Core table for storing chat messages from various platforms (primarily Matrix).

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| message_id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique message identifier |
| event_id | TEXT | NOT NULL, UNIQUE | External platform event ID |
| event_datetime | TIMESTAMP | - | When the message event occurred |
| origin_server_ts | BIGINT | - | Original server timestamp |
| sender_contact_id | UUID | NOT NULL, FK → contacts(contact_id) | Message sender |
| room_id | UUID | NOT NULL, FK → rooms(room_id) | Room where message was sent |
| message_type | TEXT | - | Type classification of message |
| message_classification | TEXT | - | Classification category |
| body | TEXT | - | Plain text message body |
| formatted_body | TEXT | - | Formatted/HTML message body |
| format | TEXT | - | Format type (e.g., 'org.matrix.custom.html') |
| msgtype | TEXT | - | Message type (e.g., 'm.text', 'm.image') |
| is_edited | BOOLEAN | DEFAULT false | Whether message has been edited |
| is_reply | BOOLEAN | DEFAULT false | Whether message is a reply |
| reply_to_event_id | TEXT | - | Event ID of message being replied to |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Record creation time |
| updated_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Record last update time |

**Indexes:**
- `idx_messages_event_id` (btree on event_id)
- `idx_messages_room_id` (btree on room_id)
- `idx_messages_sender_contact_id` (btree on sender_contact_id)
- `idx_messages_body_gin` (GIN for full-text search on body)

**Triggers:**
- `process_new_message_into_session_trigger` - Automatically adds new messages to sessions
- `process_new_message_type_trigger` - Evaluates and classifies message types

**Notes:**
- Table has `REPLICA IDENTITY FULL` for replication support
- Full-text search enabled via GIN index on message body

### message_text_representation

Stores searchable text representations of messages with full-text search vectors.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique identifier |
| message_id | UUID | NOT NULL, FK → messages(message_id) ON DELETE CASCADE | Associated message |
| text_content | TEXT | NOT NULL | Searchable text content |
| source_type | TEXT | NOT NULL | Source of text representation |
| search_vector | TSVECTOR | - | Full-text search vector |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Creation time |
| updated_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Last update time |

**Indexes:**
- `idx_message_text_search` (GIN on search_vector)

**Triggers:**
- `message_text_search_update_trigger` - Automatically updates search_vector on insert/update

### messages_edit_history

Tracks edit history for messages.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| edit_id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique edit record ID |
| message_id | UUID | NOT NULL, FK → messages(message_id) ON DELETE CASCADE | Message being edited |
| previous_body | TEXT | - | Previous plain text body |
| previous_formatted_body | TEXT | - | Previous formatted body |
| edit_timestamp | TIMESTAMP | - | When the edit occurred |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Record creation time |

**Indexes:**
- `idx_messages_edit_history_message_id` (btree on message_id)

### messages_media

Stores media attachments associated with messages.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| media_id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique media identifier |
| message_id | UUID | NOT NULL, FK → messages(message_id) ON DELETE CASCADE | Associated message |
| url | TEXT | - | Media URL |
| mimetype | TEXT | - | MIME type of media |
| size | INTEGER | - | Size in bytes |
| width | INTEGER | - | Width in pixels (for images/video) |
| height | INTEGER | - | Height in pixels (for images/video) |
| duration | INTEGER | - | Duration in milliseconds (for audio/video) |
| filename | TEXT | - | Original filename |
| is_encrypted | BOOLEAN | - | Whether media is encrypted |
| thumbnail_url | TEXT | - | Thumbnail URL |
| geo_uri | TEXT | - | Geographic URI for location media |
| location_description | TEXT | - | Human-readable location description |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Creation time |

### messages_mentions

Tracks @mentions in messages.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| mention_id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique mention ID |
| message_id | UUID | NOT NULL, FK → messages(message_id) ON DELETE CASCADE | Message containing mention |
| contact_id | UUID | FK → contacts(contact_id) | Mentioned contact |
| room_mention | BOOLEAN | - | Whether this is a room-wide mention |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Creation time |

**Indexes:**
- `idx_messages_mentions_contact_id` (btree on contact_id)

### messages_reactions

Stores emoji reactions to messages.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| reaction_id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique reaction ID |
| message_id | UUID | NOT NULL, FK → messages(message_id) ON DELETE CASCADE | Message being reacted to |
| target_event_id | TEXT | NOT NULL | Target event ID |
| sender_contact_id | UUID | NOT NULL, FK → contacts(contact_id) | User who reacted |
| key | TEXT | NOT NULL | Reaction emoji/key |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Creation time |

### messages_relations

Tracks relationships between messages (replies, edits, etc.).

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| relation_id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique relation ID |
| source_message_id | UUID | NOT NULL, FK → messages(message_id) ON DELETE CASCADE | Source message |
| target_event_id | TEXT | NOT NULL | Target event ID |
| relation_type | TEXT | NOT NULL | Type of relation (e.g., 'm.replace', 'm.thread') |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Creation time |

**Indexes:**
- `idx_messages_relations_target` (btree on target_event_id)

### raw_messages

Stores raw message data before processing.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique ID |
| external_id | TEXT | NOT NULL, UNIQUE | External platform message ID |
| content | JSONB | NOT NULL | Raw message content |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Creation time |

**Triggers:**
- `new_raw_message_to_message_trigger` - Processes raw messages into structured messages table

### rooms

Represents chat rooms or conversations.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| room_id | UUID | PRIMARY KEY, DEFAULT uuid_generate_v4() | Unique room identifier |
| source_id | TEXT | NOT NULL, UNIQUE | External platform room ID |
| display_name | TEXT | - | Room display name |
| user_defined_name | TEXT | - | User-customized room name |
| last_activity | TIMESTAMP | - | Last activity timestamp |

**Triggers:**
- `after_room_insert` - Creates entity for rooms with >3 participants

**Notes:**
- Table has `REPLICA IDENTITY FULL` for replication support

### room_state

Maintains current state and metadata for rooms.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| room_id | UUID | PRIMARY KEY, FK → rooms(room_id) | Room identifier |
| room_type | TEXT | - | Type of room |
| room_platform | TEXT | - | Platform (e.g., 'matrix') |
| display_name | TEXT | - | Current display name |
| subtitle | TEXT | - | Room subtitle |
| avatar | TEXT | - | Avatar URL |
| participant_count | INTEGER | - | Number of participants |
| unread_counts | INTEGER | - | Unread message count |
| unread_highlights | INTEGER | - | Unread mentions/highlights |
| last_activity | TIMESTAMP | - | Last activity time |
| last_message_id | UUID | FK → messages(message_id) ON DELETE SET NULL | Last message |
| last_message_text | TEXT | - | Last message preview text |
| is_muted | BOOLEAN | - | Whether room is muted |
| is_hidden | BOOLEAN | - | Whether room is hidden |
| is_favorite | BOOLEAN | - | Whether room is favorited |
| extra | JSONB | - | Additional metadata |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Creation time |
| updated_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Last update time |

**Indexes:**
- `room_state_last_activity_idx` (btree on last_activity DESC)
- `room_state_active_idx` (btree on last_activity DESC WHERE is_hidden = false)

### room_participants

Junction table tracking room membership.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique ID |
| room_id | UUID | NOT NULL, FK → rooms(room_id) | Room |
| contact_id | UUID | NOT NULL, FK → contacts(contact_id) | Participant |
| known_last_presence | TIMESTAMP | - | Last known presence time |
| known_last_exit | TIMESTAMP | - | Last known exit time |

**Unique Constraints:**
- (room_id, contact_id)

**Indexes:**
- `idx_room_participants_room_contact` (btree on room_id, contact_id)

### room_known_names

Tracks historical room names.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique ID |
| room_id | UUID | NOT NULL, FK → rooms(room_id) | Room |
| name | TEXT | NOT NULL | Known name |
| last_time | TIMESTAMP | - | Last time this name was used |

**Unique Constraints:**
- (room_id, name)

### room_known_avatars

Tracks historical room avatars.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique ID |
| room_id | UUID | FK → rooms(room_id) | Room |
| avatar | TEXT | - | Avatar URL |
| earliest_date | TIMESTAMP | - | Earliest date this avatar was used |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Creation time |

**Unique Constraints:**
- (room_id, avatar)

### sessions

Groups messages into conversation sessions based on time gaps.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| session_id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique session ID |
| room_id | UUID | NOT NULL, FK → rooms(room_id) | Room |
| first_date_time | TIMESTAMP | NOT NULL | Session start time |
| first_message_id | UUID | - | First message in session |
| last_message_id | UUID | - | Last message in session |
| last_date_time | TIMESTAMP | - | Session end time |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Creation time |
| updated_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Last update time |

**Indexes:**
- `idx_sessions_room_id` (btree on room_id)
- `idx_sessions_first_date_time` (btree on first_date_time)
- `idx_sessions_last_date_time` (btree on last_date_time)

### session_message

Junction table linking sessions to messages.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| session_id | UUID | PRIMARY KEY, FK → sessions(session_id) ON DELETE CASCADE | Session |
| message_id | UUID | PRIMARY KEY | Message |

**Primary Key:** (session_id, message_id)

### session_summaries

Stores AI-generated summaries of conversation sessions with semantic embeddings.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique ID |
| session_id | UUID | NOT NULL, FK → sessions(session_id) ON DELETE CASCADE | Associated session |
| summary | TEXT | - | Summary text |
| embedding | vector(1024) | - | **Semantic embedding vector** |
| strategy | TEXT | - | Strategy used to generate summary |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Creation time |

**pgvector Usage:**
- 1024-dimensional embeddings for semantic search of session summaries

### message_view

View providing a denormalized view of messages with room and sender names.

**Columns:**
- message_id
- room_id
- room (display name)
- from_id (sender contact_id)
- from (sender name)
- when (event_datetime)
- content (JSONB with message details)

---

## 2. Contacts & Social

### contacts

Core table for managing contact information.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| contact_id | UUID | PRIMARY KEY, DEFAULT uuid_generate_v4() | Unique contact identifier |
| name | TEXT | NOT NULL | Contact name |
| email | TEXT | - | Email address |
| phone | TEXT | - | Phone number |
| creation_date | TIMESTAMP | NOT NULL, DEFAULT now() | When contact was created |
| last_update | TIMESTAMP | NOT NULL, DEFAULT now() | Last update time |
| birthday | TEXT | - | Birthday |
| notes | TEXT | - | Notes about contact |
| extras | JSONB | DEFAULT '{}' | Additional metadata |

**Triggers:**
- `after_contact_insert` - Creates corresponding person entity
- `before_contact_delete` - Soft deletes corresponding entity

**Notes:**
- Table has `REPLICA IDENTITY FULL` for replication support

### contact_known_names

Tracks alternative/historical names for contacts.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique ID |
| contact_id | UUID | NOT NULL, FK → contacts(contact_id) | Contact |
| name | TEXT | NOT NULL | Known name |

**Unique Constraints:**
- (contact_id, name)

### contact_known_avatars

Tracks historical avatars for contacts.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique ID |
| contact_id | UUID | NOT NULL, FK → contacts(contact_id) | Contact |
| avatar | TEXT | NOT NULL | Avatar URL |
| earliest_date | TIMESTAMP | - | Earliest date this avatar was used |

**Unique Constraints:**
- (contact_id, avatar)

### contact_sources

Links contacts to external platform sources.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique ID |
| contact_id | UUID | FK → contacts(contact_id) ON DELETE CASCADE | Contact |
| source_id | TEXT | NOT NULL | External platform user ID |
| source_name | TEXT | NOT NULL | Platform name (e.g., 'matrix') |

**Indexes:**
- `idx_contact_sources_contact_id` (btree on contact_id)
- `idx_contact_sources_source_id` (btree on source_id)

### contact_evals

Stores subjective evaluations/ratings of contacts.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique ID |
| contact_id | UUID | NOT NULL, UNIQUE, FK → contacts(contact_id) ON DELETE CASCADE | Contact |
| importance | INTEGER | - | Importance rating |
| closeness | INTEGER | - | Closeness rating |
| fondness | INTEGER | - | Fondness rating |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Creation time |
| updated_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Last update time |

**Indexes:**
- `idx_contact_evals_contact_id` (btree on contact_id)

### contact_stats

Stores computed statistics about contacts.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique ID |
| contact_id | UUID | NOT NULL, FK → contacts(contact_id) ON DELETE CASCADE | Contact |
| last_week_messages | INTEGER | - | Messages sent in last week |
| groups_in_common | INTEGER | - | Number of shared groups |

### contact_tagnames

Defines available tags for contacts.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| tag_id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique tag ID |
| name | TEXT | NOT NULL, UNIQUE | Tag name |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Creation time |

### contact_tags

Junction table linking contacts to tags.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique ID |
| contact_id | UUID | NOT NULL, FK → contacts(contact_id) ON DELETE CASCADE | Contact |
| tag_id | UUID | NOT NULL, FK → contact_tagnames(tag_id) ON DELETE CASCADE | Tag |

**Unique Constraints:**
- (contact_id, tag_id)

**Indexes:**
- `idx_contact_tags_contact_id` (btree on contact_id)
- `idx_contact_tags_tag_id` (btree on tag_id)

---

## 3. Bookmarks & Content

### bookmarks

Core table for web bookmarks.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| bookmark_id | UUID | PRIMARY KEY, DEFAULT uuid_generate_v4() | Unique bookmark identifier |
| url | TEXT | NOT NULL | Bookmark URL |
| creation_date | TIMESTAMP | NOT NULL | When bookmark was created |

**Triggers:**
- `new_bookmark_trigger` - Notifies system of new bookmarks for processing

### bookmark_titles

Stores titles associated with bookmarks from various sources.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique ID |
| bookmark_id | UUID | FK → bookmarks(bookmark_id) ON DELETE CASCADE | Bookmark |
| title | TEXT | - | Title text |
| source | TEXT | - | Source of title (e.g., 'html', 'og_tag') |

### bookmark_sources

Stores source data for bookmarks.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| source_id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique source ID |
| bookmark_id | UUID | FK → bookmarks(bookmark_id) ON DELETE CASCADE | Bookmark |
| source_uri | TEXT | - | Source URI |
| raw_source | BYTEA | NOT NULL | Raw source data |

### bookmark_content_references

Stores processed content chunks with semantic embeddings for bookmarks.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique ID |
| bookmark_id | UUID | FK → bookmarks(bookmark_id) | Bookmark |
| content | TEXT | - | Content chunk text |
| strategy | TEXT | - | Processing strategy used |
| embedding | vector(1024) | - | **Semantic embedding vector** |
| created_at | TIMESTAMP | DEFAULT now() | Creation time |
| extra | JSONB | DEFAULT '{}' | Additional metadata |

**pgvector Usage:**
- 1024-dimensional embeddings enable semantic search across bookmark content chunks
- Table has `REPLICA IDENTITY FULL` for replication support

### bookmark_evaluations

Stores AI evaluations and assessments of bookmarks.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique ID |
| bookmark_id | UUID | NOT NULL, FK → bookmarks(bookmark_id) ON DELETE CASCADE | Bookmark |
| failed_fetch | BOOLEAN | - | Whether fetch failed |
| summary_id | TEXT | - | ID of summary |
| summary_comments | TEXT | - | Comments on summary |
| summary_eval | INTEGER | - | Summary evaluation score |
| questions_eval | INTEGER | - | Questions evaluation score |
| questions_ids | JSONB | NOT NULL | IDs of generated questions |
| created_at | TIMESTAMPTZ | DEFAULT CURRENT_TIMESTAMP | Creation time |

**Indexes:**
- `bookmark_evaluations_bookmark_id_idx` (btree on bookmark_id)

### processed_contents

Stores processed/extracted content from bookmarks.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| processed_content_id | UUID | PRIMARY KEY, DEFAULT uuid_generate_v4() | Unique ID |
| bookmark_id | UUID | FK → bookmarks(bookmark_id) ON DELETE CASCADE | Bookmark |
| strategy_used | TEXT | - | Processing strategy |
| processed_content | TEXT | - | Processed text content |

### http_responses

Caches HTTP responses from bookmark fetches.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| response_id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique response ID |
| bookmark_id | UUID | FK → bookmarks(bookmark_id) ON DELETE CASCADE | Associated bookmark |
| status_code | INTEGER | - | HTTP status code |
| headers | TEXT | - | Response headers |
| content | BYTEA | NOT NULL | Response body |
| fetch_date | TIMESTAMP | - | When response was fetched |

### categories

Defines bookmark categories for organization.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| category_id | UUID | PRIMARY KEY, DEFAULT uuid_generate_v4() | Unique category ID |
| name | TEXT | NOT NULL | Category name |

### category_sources

Stores source data for categories.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique ID |
| category_id | UUID | FK → categories(category_id) ON DELETE CASCADE | Category |
| source_uri | TEXT | - | Source URI |
| raw_source | BYTEA | NOT NULL | Raw source data |

### bookmark_category

Junction table linking bookmarks to categories.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique ID |
| bookmark_id | UUID | FK → bookmarks(bookmark_id) ON DELETE CASCADE | Bookmark |
| category_id | UUID | FK → categories(category_id) ON DELETE CASCADE | Category |

### browser_history

Stores browser history data, primarily imported from Firefox.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PRIMARY KEY | Unique ID |
| url | TEXT | NOT NULL | Visited URL |
| title | TEXT | - | Page title |
| visit_date | TIMESTAMP | NOT NULL | Visit timestamp |
| typed | BOOLEAN | - | Whether URL was typed |
| hidden | BOOLEAN | - | Whether visit is hidden |
| imported_from_firefox_place_id | INTEGER | - | Firefox place ID |
| imported_from_firefox_visit_id | INTEGER | - | Firefox visit ID (UNIQUE) |
| domain | TEXT | - | Extracted domain |
| created_at | TIMESTAMPTZ | DEFAULT CURRENT_TIMESTAMP | Record creation time |

**Indexes:**
- `idx_browser_history_domain` (btree on domain)
- `idx_browser_history_visit_date` (btree on visit_date)

---

## 4. Knowledge Management (Items & Tags)

### items

Core table for notes and knowledge items.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY, DEFAULT uuid_generate_v4() | Unique item identifier |
| title | TEXT | - | Item title |
| slug | TEXT | - | URL-friendly slug |
| contents | TEXT | - | Item content/body |
| created | BIGINT | DEFAULT EXTRACT(epoch FROM CURRENT_TIMESTAMP) | Creation timestamp (Unix epoch) |
| modified | BIGINT | DEFAULT EXTRACT(epoch FROM CURRENT_TIMESTAMP) | Last modified timestamp (Unix epoch) |

**Triggers:**
- `notify_new_item_trigger` - Notifies system of new items
- `trigger_item_name_to_slug` - Auto-generates slug from title
- `update_modified_before_update` - Updates modified timestamp on changes

### item_semantic_index

Stores semantic embeddings for items to enable vector similarity search.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY, DEFAULT uuid_generate_v4() | Unique ID |
| item_id | UUID | FK → items(id) | Associated item |
| embedding | vector(1024) | - | **Semantic embedding vector** |

**pgvector Usage:**
- 1024-dimensional embeddings for semantic search across notes/items
- Table has `REPLICA IDENTITY FULL` for replication support

### tags

Defines tags for organizing items.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY, DEFAULT uuid_generate_v4() | Unique tag ID |
| name | TEXT | NOT NULL, UNIQUE | Tag name |
| created | BIGINT | DEFAULT EXTRACT(epoch FROM CURRENT_TIMESTAMP) | Creation timestamp |
| modified | BIGINT | DEFAULT EXTRACT(epoch FROM CURRENT_TIMESTAMP) | Last modified timestamp |
| last_activity | BIGINT | - | Last activity timestamp |

**Triggers:**
- `update_tag_modified_before_update` - Updates modified timestamp
- `update_tag_last_activity_after_insert_or_delete` - Updates last_activity on tag usage

### item_tags

Junction table linking items to tags.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| item_id | UUID | PRIMARY KEY, FK → items(id) | Item |
| tag_id | UUID | PRIMARY KEY, FK → tags(id) | Tag |

**Primary Key:** (item_id, tag_id)

### full_items

View joining items with their tags as JSONB array.

**Columns:**
- id
- title
- contents
- created
- modified
- tags (JSONB array of tag names)

---

## 5. Entities & Relationships

### entities

Generic entity system for representing people, groups, organizations, etc.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| entity_id | UUID | PRIMARY KEY, DEFAULT uuid_generate_v4() | Unique entity identifier |
| name | TEXT | NOT NULL | Entity name |
| type | TEXT | NOT NULL | Entity type (e.g., 'person', 'group_chat', 'organization') |
| description | TEXT | - | Entity description |
| properties | JSONB | DEFAULT '{}' | Additional properties |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Creation time |
| updated_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Last update time |
| deleted_at | TIMESTAMP | - | Soft deletion time |

**Indexes:**
- `idx_entities_type` (btree on type)

**Purpose:**
- Unified system for cross-domain entity tracking
- Automatically created for contacts (as 'person' entities)
- Automatically created for large group chats (>3 participants)

### entity_references

Tracks references to entities in various content sources.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique ID |
| source_type | TEXT | NOT NULL | Type of source (e.g., 'message', 'note') |
| source_id | UUID | NOT NULL | ID of source record |
| entity_id | UUID | NOT NULL, FK → entities(entity_id) | Referenced entity |
| reference_text | TEXT | NOT NULL | Text that references the entity |
| position | INTEGER | - | Position in source text |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Creation time |

**Indexes:**
- `idx_entity_references_entity_id` (btree on entity_id)
- `idx_entity_references_source` (btree on source_type, source_id)

### entity_relationships

Defines relationships between entities.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique ID |
| entity_id | UUID | NOT NULL, FK → entities(entity_id) | Source entity |
| related_type | TEXT | NOT NULL | Type of related record |
| related_id | UUID | NOT NULL | ID of related record |
| relationship_type | TEXT | NOT NULL | Relationship type (e.g., 'identity', 'member_of') |
| metadata | JSONB | - | Relationship metadata |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Creation time |
| updated_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Last update time |

**Indexes:**
- `idx_entity_relationships_entity_id` (btree on entity_id)
- `idx_entity_relationships_related_id` (btree on related_id)
- `idx_entity_relationships_types` (btree on related_type, relationship_type)

---

## 6. Dispatch System (Audio Transcription & Analysis)

The Dispatch system handles audio recording, transcription, and AI-powered analysis.

### dispatch_audio_raw

Stores raw audio files.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PRIMARY KEY | Unique audio ID |
| file_data | BYTEA | NOT NULL | Raw audio file data |
| file_name | TEXT | NOT NULL | Original filename |
| file_type | TEXT | NOT NULL | File MIME type |
| file_size | INTEGER | NOT NULL | Size in bytes |
| duration | REAL | - | Duration in seconds |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Upload time |
| metadata | JSONB | NOT NULL | Additional metadata |

### dispatch_transcription

Stores transcriptions of audio files.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PRIMARY KEY | Unique transcription ID |
| audio_id | INTEGER | NOT NULL, FK → dispatch_audio_raw(id) ON DELETE CASCADE | Source audio |
| transcription_text | TEXT | NOT NULL | Transcribed text |
| is_edited | BOOLEAN | NOT NULL | Whether transcription was manually edited |
| parent_id | INTEGER | FK → dispatch_transcription(id) ON DELETE SET NULL | Previous version if edited |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Creation time |
| extra | JSONB | NOT NULL | Additional data |

**Indexes:**
- `idx_dispatch_transcription_audio_id` (btree on audio_id)

### dispatch_analysis

Stores AI analysis of transcriptions.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PRIMARY KEY | Unique analysis ID |
| transcription_id | INTEGER | NOT NULL, FK → dispatch_transcription(id) ON DELETE CASCADE | Source transcription |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Creation time |
| is_edited | BOOLEAN | NOT NULL | Whether analysis was edited |
| parent_id | INTEGER | FK → dispatch_analysis(id) ON DELETE SET NULL | Previous version if edited |
| prompt_used | TEXT | NOT NULL | AI prompt used |
| raw_response | TEXT | NOT NULL | Raw AI response |
| extra | JSONB | NOT NULL | Additional data |

**Indexes:**
- `idx_dispatch_analysis_transcription_id` (btree on transcription_id)

### dispatch_bulletpoint

Stores individual bullet points extracted from analysis.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PRIMARY KEY | Unique bulletpoint ID |
| analysis_id | INTEGER | NOT NULL, FK → dispatch_analysis(id) ON DELETE CASCADE | Parent analysis |
| from_position | INTEGER | NOT NULL | Start position in transcription |
| to_position | INTEGER | NOT NULL | End position in transcription |
| info | TEXT | NOT NULL | Bullet point text |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Creation time |
| order_index | INTEGER | NOT NULL | Display order |

**Indexes:**
- `idx_dispatch_bulletpoint_analysis_id` (btree on analysis_id)

### dispatch_extracted_entity

Stores entities (people, places, things) extracted from bullet points.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PRIMARY KEY | Unique entity ID |
| bulletpoint_id | INTEGER | NOT NULL, FK → dispatch_bulletpoint(id) ON DELETE CASCADE | Source bulletpoint |
| entity_text | TEXT | NOT NULL | Extracted entity text |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Creation time |

**Indexes:**
- `idx_dispatch_extracted_entity_bulletpoint_id` (btree on bulletpoint_id)

### dispatch_entity_merge

Tracks merging of duplicate extracted entities.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PRIMARY KEY | Unique merge ID |
| entity_id | INTEGER | NOT NULL, FK → dispatch_extracted_entity(id) ON DELETE CASCADE | Primary entity |
| merged_entity_id | INTEGER | NOT NULL, FK → dispatch_extracted_entity(id) ON DELETE CASCADE | Merged entity |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Merge time |
| created_by | TEXT | - | User who performed merge |
| notes | TEXT | - | Merge notes |

### dispatch_job

Tracks the complete workflow from audio to analysis.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | INTEGER | PRIMARY KEY | Unique job ID |
| title | VARCHAR | - | Job title |
| created_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Creation time |
| updated_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Last update time |
| audio_id | INTEGER | FK → dispatch_audio_raw(id) ON DELETE SET NULL | Associated audio |
| transcription_id | INTEGER | FK → dispatch_transcription(id) ON DELETE SET NULL | Associated transcription |
| analysis_id | INTEGER | FK → dispatch_analysis(id) ON DELETE SET NULL | Associated analysis |
| status | VARCHAR | NOT NULL, DEFAULT 'pending' | Job status |
| metadata | JSONB | NOT NULL, DEFAULT '{}' | Job metadata |

**Constraints:**
- status CHECK: must be one of ('pending', 'in_progress', 'completed', 'failed')

---

## 7. AI Conversations (Alicia)

Tables for tracking conversations with the Alicia AI assistant.

### alicia_conversations

Represents conversation threads with Alicia.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | TEXT | PRIMARY KEY | Unique conversation ID |
| created_at | TIMESTAMP | - | Creation time |

### alicia_message

Individual messages in Alicia conversations.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | TEXT | PRIMARY KEY | Unique message ID |
| role | TEXT | NOT NULL | Message role (e.g., 'user', 'assistant') |
| contents | TEXT | NOT NULL | Message content |
| previous_id | TEXT | FK → alicia_message(id) | Previous message in thread |
| conversation_id | TEXT | FK → alicia_conversations(id) | Parent conversation |
| created_at | TIMESTAMP | - | Creation time |

### alicia_meta

Metadata and references for Alicia messages.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | TEXT | PRIMARY KEY | Unique metadata ID |
| ref | TEXT | FK → alicia_message(id) | Referenced message |
| contents | BYTEA | NOT NULL | Metadata content |
| conversation_id | TEXT | FK → alicia_conversations(id) | Parent conversation |
| created_at | TIMESTAMP | - | Creation time |

---

## 8. Configuration & Observations

### configurations

Stores application configuration key-value pairs.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| config_id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique config ID |
| key | TEXT | NOT NULL, UNIQUE | Configuration key |
| value | TEXT | NOT NULL | Configuration value |
| is_secret | BOOLEAN | NOT NULL | Whether value is sensitive |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Creation time |
| updated_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Last update time |

**Indexes:**
- `idx_configurations_key` (UNIQUE btree on key)
- `idx_configurations_is_secret` (btree on is_secret)

### observations

Generic storage for observational data and events.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| observation_id | UUID | PRIMARY KEY, DEFAULT uuid_generate_v4() | Unique observation ID |
| data | JSONB | NOT NULL | Observation data |
| type | TEXT | - | Observation type |
| source | TEXT | - | Data source |
| tags | TEXT | - | Space-separated tags |
| parent | UUID | - | Parent observation |
| ref | UUID | - | Referenced observation |
| creation_date | BIGINT | NOT NULL, DEFAULT EXTRACT(epoch FROM CURRENT_TIMESTAMP) | Creation timestamp (Unix epoch) |

---

## 9. Social Posts

### social_posts

Tracks cross-platform social media posts.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| post_id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique post ID |
| content | TEXT | NOT NULL | Post content/text |
| twitter_post_id | TEXT | - | Twitter/X post ID if posted |
| bluesky_post_id | TEXT | - | Bluesky post ID if posted |
| created_at | TIMESTAMPTZ | DEFAULT CURRENT_TIMESTAMP | Creation time |
| updated_at | TIMESTAMPTZ | DEFAULT CURRENT_TIMESTAMP | Last update time |
| status | TEXT | NOT NULL | Post status |
| error_message | TEXT | - | Error message if failed |

**Indexes:**
- `idx_social_posts_created_at` (btree on created_at DESC NULLS LAST)

---

## Legacy Tables

### messages_old

Legacy messages table, superseded by the current messages table.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| message_id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique message ID |
| sender_id | UUID | NOT NULL, FK → contacts(contact_id) | Sender |
| room_id | UUID | NOT NULL, FK → rooms(room_id) | Room |
| content | JSONB | NOT NULL | Message content |
| external_id | TEXT | NOT NULL, UNIQUE | External platform ID |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Creation time |

**Indexes:**
- `idx_messages_created_at` (btree on created_at)
- `idx_messages_external_id` (btree on external_id)
- `idx_messages_old_room_id` (btree on room_id)

---

## Vector Embeddings (pgvector)

The database uses the **pgvector** extension to store and query high-dimensional vector embeddings for semantic search and similarity matching.

### Tables with Vector Columns

| Table | Column | Dimensions | Purpose |
|-------|--------|------------|---------|
| **bookmark_content_references** | embedding | 1024 | Semantic search across bookmark content chunks |
| **item_semantic_index** | embedding | 1024 | Semantic search across notes/knowledge items |
| **session_summaries** | embedding | 1024 | Similarity search for conversation summaries |

### Vector Search Capabilities

All vector columns use 1024-dimensional embeddings, likely generated from text using models like:
- OpenAI's text-embedding-3-small (1536 dims, potentially truncated)
- Custom embedding models
- Sentence transformers

**Common Operations:**
```sql
-- Find similar content using cosine similarity
SELECT * FROM bookmark_content_references
ORDER BY embedding <=> $1
LIMIT 10;

-- Find similar notes
SELECT * FROM item_semantic_index
ORDER BY embedding <-> $1
LIMIT 5;
```

**Distance Operators:**
- `<->` : L2 distance (Euclidean)
- `<#>` : Inner product (negative, for maximum inner product search)
- `<=>` : Cosine distance (1 - cosine similarity)

---

## Indexes Summary

### Full-Text Search Indexes

- **messages.body**: GIN index on `to_tsvector('english', body)` for fast text search
- **message_text_representation.search_vector**: GIN index for full-text search on processed message text

### Performance Indexes

**Messages & Rooms:**
- room_id, sender_contact_id, event_id for fast message lookups
- last_activity indexes on room_state for active room queries

**Contacts:**
- source_id, source_name for external platform lookups
- contact_id indexes on all related tables (tags, stats, evals)

**Sessions:**
- room_id, first_date_time, last_date_time for temporal queries

**Entities:**
- type for filtering by entity type
- Combined indexes on (source_type, source_id) and (related_type, relationship_type)

**Bookmarks:**
- domain for browser history grouping
- visit_date for temporal queries

### Unique Constraints

Key unique constraints ensure data integrity:
- **messages.event_id**: Prevents duplicate message imports
- **rooms.source_id**: Prevents duplicate room records
- **tags.name**: Tag name uniqueness
- **configurations.key**: Configuration key uniqueness
- **Composite uniques**: Many junction tables use composite unique constraints (e.g., contact_id + tag_id)

---

## Foreign Key Relationships

### Cascade Policies

**ON DELETE CASCADE** (child records deleted when parent is deleted):
- All junction tables (item_tags, contact_tags, bookmark_category, etc.)
- Message relationships (messages_media, messages_mentions, messages_reactions, etc.)
- Bookmark relationships (bookmark_titles, bookmark_sources, http_responses, etc.)
- Session relationships (session_message, session_summaries)
- Dispatch workflow (transcription → analysis → bulletpoints → entities)

**ON DELETE SET NULL** (child records preserved, FK set to NULL):
- dispatch_job references (audio_id, transcription_id, analysis_id)
- room_state.last_message_id
- Versioning parent_id columns (dispatch_analysis.parent_id, dispatch_transcription.parent_id)

---

## Database Functions & Procedures

The schema includes several PL/pgSQL functions and procedures:

### Tag Management
- `add_tag(item_id, tagname)`: Adds tag to item, creating tag if needed
- `add_contact_tag(contact_id, tagname)`: Adds tag to contact

### Message Processing
- `add_message_to_session(message_id, time_gap)`: Groups messages into sessions based on time gaps
- `add_raw_message(raw_id, input_json, input_date)`: Processes raw message imports
- `process_raw_message_to_message()`: Trigger function to parse raw messages

### Entity Management
- `create_contact_entity()`: Auto-creates entity when contact is inserted
- `create_room_entity()`: Auto-creates entity for group chats (>3 participants)
- `delete_contact_entity()`: Soft-deletes entity when contact is deleted

### Search & Utility
- `encode_zbase32(num)`: Encodes numbers using zbase32 encoding
- `message_text_search_update()`: Updates full-text search vectors
- `update_modified_column()`: Updates modified timestamps
- `set_slug_from_name()`: Auto-generates URL slugs

---

## Triggers

### Data Processing Triggers

| Trigger | Table | Event | Function | Purpose |
|---------|-------|-------|----------|---------|
| new_bookmark_trigger | bookmarks | AFTER INSERT | notify_new_bookmark() | Queues bookmark for processing |
| notify_new_item_trigger | items | AFTER INSERT | notify_new_item() | Queues note for indexing |
| new_raw_message_to_message_trigger | raw_messages | AFTER INSERT | process_raw_message_to_message() | Parses raw message data |
| process_new_message_into_session_trigger | messages | AFTER INSERT | process_new_message_into_session() | Groups messages into sessions |
| process_new_message_type_trigger | messages | AFTER INSERT | process_message_type_eval() | Classifies message types |

### Entity Management Triggers

| Trigger | Table | Event | Function | Purpose |
|---------|-------|-------|----------|---------|
| after_contact_insert | contacts | AFTER INSERT | create_contact_entity() | Creates person entity |
| after_room_insert | rooms | AFTER INSERT | create_room_entity() | Creates group entity |
| before_contact_delete | contacts | BEFORE DELETE | delete_contact_entity() | Soft-deletes entity |

### Maintenance Triggers

| Trigger | Table | Event | Function | Purpose |
|---------|-------|-------|----------|---------|
| message_text_search_update_trigger | message_text_representation | BEFORE INSERT/UPDATE | message_text_search_update() | Updates search vectors |
| trigger_item_name_to_slug | items | BEFORE INSERT | set_slug_from_name() | Generates URL slugs |
| update_modified_before_update | items | BEFORE UPDATE | update_modified_column() | Updates timestamps |
| update_tag_modified_before_update | tags | BEFORE UPDATE | update_tag_modified_column() | Updates timestamps |
| update_tag_last_activity_after_insert_or_delete | item_tags | AFTER INSERT/DELETE | update_tag_last_activity() | Tracks tag usage |

---

## Migration Strategy

### Current Approach

The project uses **schema.sql** as the single source of truth for the database schema. This file is:

1. **Generated by pg_dump**: The schema.sql file is a complete PostgreSQL dump including:
   - Extension definitions
   - Custom functions and procedures
   - Table definitions
   - Indexes, constraints, and triggers
   - Initial data (if any)

2. **Code Generation with sqlc**: The project uses [sqlc](https://sqlc.dev/) to generate type-safe Go code from SQL:
   - Configuration: `/home/user/garden/sqlc.yaml`
   - Queries: `/home/user/garden/internal/adapter/secondary/postgres/queries/*.sql`
   - Generated models: `/home/user/garden/internal/adapter/secondary/postgres/generated/db/models.go`
   - Generated query functions: `/home/user/garden/internal/adapter/secondary/postgres/generated/db/*.sql.go`

### Recommended Migration Strategy

Since no formal migration system is currently in place, consider adopting one of these approaches:

#### Option 1: golang-migrate/migrate

```bash
# Install golang-migrate
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Create migration
migrate create -ext sql -dir migrations -seq initial_schema

# Run migrations
migrate -database "postgres://user:pass@localhost:5432/garden?sslmode=disable" \
        -path migrations up
```

**Directory structure:**
```
migrations/
  000001_initial_schema.up.sql     # Initial schema from schema.sql
  000001_initial_schema.down.sql   # Rollback (DROP everything)
  000002_add_feature.up.sql        # Future migrations
  000002_add_feature.down.sql
```

#### Option 2: Goose

```bash
# Install goose
go install github.com/pressly/goose/v3/cmd/goose@latest

# Create migration
goose -dir migrations create add_feature sql

# Run migrations
goose -dir migrations postgres "user=postgres dbname=garden" up
```

#### Option 3: Atlas (Schema-as-Code)

[Atlas](https://atlasgo.io/) provides declarative schema management with automatic migration generation:

```bash
# Install atlas
brew install ariga/tap/atlas  # or other installation methods

# Inspect current database
atlas schema inspect -u "postgres://localhost:5432/garden" > schema.hcl

# Generate migration from schema changes
atlas schema diff --from file://schema.sql --to file://new_schema.sql
```

### Migration Best Practices

1. **Version Control**: All migrations should be committed to git
2. **Idempotency**: Use `IF EXISTS` and `IF NOT EXISTS` where appropriate
3. **Rollback Support**: Always provide down migrations
4. **Data Migrations**: Separate schema changes from data migrations
5. **Testing**: Test migrations on staging before production
6. **Backup**: Always backup before running migrations in production

### Schema Maintenance Workflow

**Current workflow:**
```bash
# 1. Make changes directly to database
psql garden

# 2. Export updated schema
pg_dump --schema-only garden > schema.sql

# 3. Regenerate Go code
sqlc generate
```

**Recommended workflow with migrations:**
```bash
# 1. Create migration file
migrate create -ext sql -dir migrations -seq add_new_feature

# 2. Write migration SQL
vim migrations/000003_add_new_feature.up.sql

# 3. Run migration
migrate -database $DATABASE_URL -path migrations up

# 4. Update schema.sql (if maintaining it)
pg_dump --schema-only garden > schema.sql

# 5. Regenerate Go code
sqlc generate

# 6. Commit migration + generated code
git add migrations/ internal/adapter/secondary/postgres/generated/
git commit -m "Add new feature to database"
```

### Database Initialization

For new deployments:

```bash
# Using schema.sql directly
psql -U postgres -d garden -f schema.sql

# Or with migrations (recommended)
migrate -database $DATABASE_URL -path migrations up
```

### Replica Identity

Several tables use `REPLICA IDENTITY FULL`:
- messages
- contacts
- rooms
- bookmark_content_references
- item_semantic_index

This setting is important for logical replication and change data capture (CDC) systems. It ensures that all column values are included in the replication stream, not just the primary key.

---

## Performance Considerations

### Connection Pooling

Consider using PgBouncer or built-in pgx connection pooling for high-traffic applications.

### Vector Search Optimization

For large-scale vector search, consider creating HNSW or IVFFlat indexes:

```sql
-- Create HNSW index for approximate nearest neighbor search
CREATE INDEX ON bookmark_content_references
USING hnsw (embedding vector_cosine_ops);

-- Or IVFFlat index
CREATE INDEX ON item_semantic_index
USING ivfflat (embedding vector_l2_ops)
WITH (lists = 100);
```

### Partitioning Considerations

For tables that grow very large, consider partitioning:
- **messages**: Partition by created_at (monthly or quarterly)
- **browser_history**: Partition by visit_date
- **observations**: Partition by creation_date

### Vacuum and Analyze

Set up regular maintenance:
```sql
-- Auto-vacuum should be configured in postgresql.conf
ALTER TABLE messages SET (autovacuum_vacuum_scale_factor = 0.1);
ALTER TABLE browser_history SET (autovacuum_vacuum_scale_factor = 0.2);
```

---

## Security Considerations

1. **Sensitive Data**: The `configurations.is_secret` flag marks sensitive values
2. **JSONB Fields**: Review `extras`, `metadata`, and `properties` fields for sensitive data
3. **Access Control**: Implement row-level security (RLS) if multi-tenant
4. **Encryption**: Consider encrypting sensitive BYTEA fields (file_data, raw_source)
5. **Audit Logging**: Consider adding audit triggers for sensitive tables

---

## Future Enhancements

Potential improvements to consider:

1. **Temporal Tables**: Use PostgreSQL temporal tables for full audit history
2. **GraphQL Integration**: The entity system could power a GraphQL API
3. **Materialized Views**: Create materialized views for expensive dashboard queries
4. **Composite Indexes**: Add composite indexes based on query patterns
5. **Vector Index Tuning**: Benchmark and optimize vector index parameters
6. **Archival Strategy**: Implement archival for old messages and browser history
7. **Real-time Sync**: Use logical replication or LISTEN/NOTIFY for real-time features

---

## References

- PostgreSQL Documentation: https://www.postgresql.org/docs/
- pgvector: https://github.com/pgvector/pgvector
- sqlc: https://sqlc.dev/
- golang-migrate: https://github.com/golang-migrate/migrate
- Atlas: https://atlasgo.io/

---

**Last Updated**: 2025-12-29
**Schema Version**: PostgreSQL 17.7
**Database Dump Version**: pg_dump 17.7
