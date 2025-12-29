# API Reference

## Overview

This document provides comprehensive documentation for the Garden API. The API follows RESTful principles and returns JSON responses.

### Base URL

```
http://localhost:8080
```

The default port is 8080, but can be configured using the `PORT` environment variable.

### Content Type

All requests and responses use `application/json` content type unless otherwise specified.

### CORS

The API supports Cross-Origin Resource Sharing (CORS) with the following configuration:
- **Allowed Origins**: `*` (all origins)
- **Allowed Methods**: `GET`, `POST`, `PUT`, `DELETE`, `OPTIONS`
- **Allowed Headers**: `Accept`, `Authorization`, `Content-Type`, `X-Request-ID`

### Request Timeout

All requests have a 60-second timeout.

---

## Authentication

> **Note**: Authentication middleware should be configured based on your deployment requirements. The handlers are designed to work with standard HTTP authentication mechanisms.

---

## Common Response Formats

### Success Response

```json
{
  "data": { ... },
  "message": "Operation successful"
}
```

### Error Response

```json
{
  "error": "Error type",
  "message": "Detailed error message"
}
```

### Paginated Response

```json
{
  "data": [ ... ],
  "pagination": {
    "page": 1,
    "totalPages": 10,
    "limit": 20,
    "totalItems": 200
  }
}
```

---

## Health Check

### Get Health Status

**Endpoint**: `GET /health`

**Description**: Returns the health status of the API.

**Response**: `200 OK`
```json
{
  "status": "ok"
}
```

---

## Bookmarks API

Manage web bookmarks with content fetching, processing, and vector search capabilities.

### List Bookmarks

**Endpoint**: `GET /api/bookmarks`

**Description**: Get filtered and paginated bookmarks.

**Query Parameters**:
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `categoryId` | UUID | No | - | Filter by category ID |
| `searchQuery` | string | No | - | Search query |
| `startCreationDate` | string | No | - | Start date (RFC3339 format) |
| `endCreationDate` | string | No | - | End date (RFC3339 format) |
| `page` | integer | No | 1 | Page number |
| `limit` | integer | No | 10 | Items per page |

**Response**: `200 OK`
```json
{
  "data": [
    {
      "bookmark_id": "uuid",
      "url": "https://example.com",
      "title": "Example Page",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "totalPages": 5,
    "limit": 10,
    "totalItems": 50
  }
}
```

### Get Random Bookmark

**Endpoint**: `GET /api/bookmarks/random`

**Description**: Redirect to a random bookmark.

**Response**: `302 Found` (redirects to `/api/bookmarks/{id}`)

### Search Bookmarks

**Endpoint**: `GET /api/bookmarks/search`

**Description**: Perform vector similarity search on bookmarks.

**Query Parameters**:
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `query` | string | Yes | - | Search query |
| `strategy` | string | No | qa-v2-passage | Search strategy |

**Response**: `200 OK`
```json
[
  {
    "bookmark_id": "uuid",
    "url": "https://example.com",
    "title": "Example Page",
    "similarity_score": 0.95
  }
]
```

### Get Bookmark Details

**Endpoint**: `GET /api/bookmarks/{id}`

**Description**: Get complete bookmark details with all relations.

**Path Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | UUID | Yes | Bookmark ID |

**Response**: `200 OK`
```json
{
  "bookmark_id": "uuid",
  "url": "https://example.com",
  "title": "Example Page",
  "category": { ... },
  "content": { ... },
  "embeddings": [ ... ],
  "questions": [ ... ]
}
```

### Get Bookmarks Missing HTTP Responses

**Endpoint**: `GET /api/bookmarks/missing/http`

**Description**: Get bookmarks that don't have HTTP responses.

**Response**: `200 OK`
```json
[
  {
    "bookmark_id": "uuid",
    "url": "https://example.com"
  }
]
```

### Get Bookmarks Missing Reader Content

**Endpoint**: `GET /api/bookmarks/missing/reader`

**Description**: Get bookmarks that don't have reader-processed content.

**Response**: `200 OK`
```json
[
  {
    "bookmark_id": "uuid",
    "url": "https://example.com"
  }
]
```

### Update Bookmark Question

**Endpoint**: `PUT /api/bookmarks/{id}/question`

**Description**: Update a question and answer content reference.

**Request Body**:
```json
{
  "question": "What is this about?",
  "answer": "This is about...",
  "content_reference_id": "uuid"
}
```

**Response**: `200 OK`
```json
{
  "message": "Question updated successfully"
}
```

### Delete Bookmark Question

**Endpoint**: `DELETE /api/bookmarks/{id}/question/{refId}`

**Description**: Delete a question and answer content reference.

**Response**: `200 OK`
```json
{
  "message": "Question deleted successfully"
}
```

### Fetch Bookmark Content

**Endpoint**: `POST /api/bookmarks/{id}/fetch`

**Description**: Fetch and store HTTP content for a bookmark.

**Response**: `200 OK`
```json
{
  "success": true,
  "status_code": 200,
  "content_length": 12345,
  "content_type": "text/html"
}
```

### Process with Lynx

**Endpoint**: `POST /api/bookmarks/{id}/process/lynx`

**Description**: Process bookmark content using lynx text browser.

**Response**: `200 OK`
```json
{
  "success": true,
  "content_length": 5000,
  "processing_time_ms": 250
}
```

### Process with Reader

**Endpoint**: `POST /api/bookmarks/{id}/process/reader`

**Description**: Process bookmark content using reader mode (Mozilla Readability).

**Response**: `200 OK`
```json
{
  "success": true,
  "title": "Article Title",
  "content_length": 5000,
  "processing_time_ms": 250
}
```

### Create Embeddings

**Endpoint**: `POST /api/bookmarks/{id}/embeddings`

**Description**: Create chunked embeddings for bookmark content.

**Response**: `200 OK`
```json
{
  "success": true,
  "chunks_created": 15,
  "embedding_model": "nomic-embed-text"
}
```

### Create Summary Embedding

**Endpoint**: `POST /api/bookmarks/{id}/summary-embedding`

**Description**: Create a summary and its embedding using AI.

**Response**: `200 OK`
```json
{
  "success": true,
  "summary": "This article discusses...",
  "embedding_created": true
}
```

### Get Bookmark Title

**Endpoint**: `GET /api/bookmarks/{id}/title`

**Description**: Extract and store the bookmark title from HTML.

**Response**: `200 OK`
```json
{
  "title": "Example Page Title",
  "source": "html_title_tag"
}
```

---

## Browser History API

Manage and search browser history entries.

### List Browser History

**Endpoint**: `GET /api/history`

**Description**: Get filtered and paginated browser history.

**Query Parameters**:
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `q` | string | No | - | Search query |
| `start_date` | string | No | - | Start date (RFC3339) |
| `end_date` | string | No | - | End date (RFC3339) |
| `domain` | string | No | - | Filter by domain |
| `page` | integer | No | 1 | Page number |
| `page_size` | integer | No | 10 | Items per page |

**Response**: `200 OK`
```json
{
  "data": [
    {
      "id": "uuid",
      "url": "https://example.com",
      "title": "Example Page",
      "visit_time": "2024-01-01T00:00:00Z",
      "visit_count": 5
    }
  ],
  "pagination": { ... }
}
```

### Get Top Domains

**Endpoint**: `GET /api/history/domains`

**Description**: Get most visited domains with aggregated counts.

**Query Parameters**:
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `limit` | integer | No | 10 | Number of domains to return |

**Response**: `200 OK`
```json
[
  {
    "domain": "example.com",
    "visit_count": 125
  }
]
```

### Get Recent History

**Endpoint**: `GET /api/history/recent`

**Description**: Get N most recent history entries.

**Query Parameters**:
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `limit` | integer | No | 20 | Number of entries to return |

**Response**: `200 OK`
```json
[
  {
    "id": "uuid",
    "url": "https://example.com",
    "title": "Example Page",
    "visit_time": "2024-01-01T00:00:00Z"
  }
]
```

---

## Categories API

Manage bookmark categories and their sources.

### List Categories

**Endpoint**: `GET /api/categories`

**Description**: Get all categories with their sources.

**Response**: `200 OK`
```json
{
  "data": [
    {
      "category_id": "uuid",
      "name": "Technology",
      "sources": [
        {
          "id": "uuid",
          "category_id": "uuid",
          "source_uri": "https://hn.algolia.com/",
          "raw_source": { ... }
        }
      ]
    }
  ],
  "pagination": { ... }
}
```

### Get Category

**Endpoint**: `GET /api/categories/{id}`

**Description**: Get a single category by ID.

**Response**: `200 OK`
```json
{
  "category_id": "uuid",
  "name": "Technology"
}
```

### Update Category

**Endpoint**: `PUT /api/categories/{id}`

**Description**: Update a category's name.

**Request Body**:
```json
{
  "name": "New Category Name"
}
```

**Response**: `204 No Content`

### Merge Categories

**Endpoint**: `POST /api/categories/merge`

**Description**: Merge two categories together.

**Request Body**:
```json
{
  "source_id": "uuid",
  "target_id": "uuid"
}
```

**Response**: `204 No Content`

### Create Category Source

**Endpoint**: `POST /api/categories/{id}/sources`

**Description**: Add a source to a category.

**Request Body**:
```json
{
  "source_uri": "https://example.com/feed",
  "raw_source": { ... }
}
```

**Response**: `201 Created`
```json
{
  "id": "uuid",
  "category_id": "uuid",
  "source_uri": "https://example.com/feed",
  "raw_source": { ... }
}
```

### Update Category Source

**Endpoint**: `PUT /api/categories/sources/{id}`

**Description**: Update a category source.

**Request Body**:
```json
{
  "source_uri": "https://example.com/newfeed",
  "raw_source": { ... }
}
```

**Response**: `204 No Content`

### Delete Category Source

**Endpoint**: `DELETE /api/categories/sources/{id}`

**Description**: Delete a category source.

**Response**: `204 No Content`

---

## Configurations API

Manage application configuration key-value pairs.

### List Configurations

**Endpoint**: `GET /api/configurations`

**Description**: Get all configurations with optional filtering.

**Query Parameters**:
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `prefix` | string | No | - | Key prefix filter |
| `include_secrets` | boolean | No | true | Include secret configurations |

**Response**: `200 OK`
```json
[
  {
    "key": "api.timeout",
    "value": "30s",
    "is_secret": false,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

### Get Configuration

**Endpoint**: `GET /api/configurations/{key}`

**Description**: Get a single configuration by its key.

**Response**: `200 OK`
```json
{
  "key": "api.timeout",
  "value": "30s",
  "is_secret": false
}
```

### Get Configurations by Prefix

**Endpoint**: `GET /api/configurations/prefix/{prefix}`

**Description**: Get all configurations with a given key prefix.

**Query Parameters**:
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `include_secrets` | boolean | No | true | Include secret configurations |

**Response**: `200 OK`
```json
[
  {
    "key": "api.timeout",
    "value": "30s",
    "is_secret": false
  }
]
```

### Create Configuration

**Endpoint**: `POST /api/configurations`

**Description**: Create a new configuration entry.

**Request Body**:
```json
{
  "key": "api.timeout",
  "value": "30s",
  "is_secret": false
}
```

**Response**: `201 Created`
```json
{
  "key": "api.timeout",
  "value": "30s",
  "is_secret": false,
  "created_at": "2024-01-01T00:00:00Z"
}
```

### Update Configuration

**Endpoint**: `PUT /api/configurations/{key}`

**Description**: Update an existing configuration.

**Request Body**:
```json
{
  "value": "60s",
  "is_secret": false
}
```

**Response**: `200 OK`
```json
{
  "key": "api.timeout",
  "value": "60s",
  "is_secret": false
}
```

### Delete Configuration

**Endpoint**: `DELETE /api/configurations/{key}`

**Description**: Delete a configuration by key.

**Response**: `204 No Content`

### Set Configuration (Upsert)

**Endpoint**: `PUT /api/configurations/{key}/value`

**Description**: Create or update a configuration value.

**Request Body**:
```json
{
  "value": "30s",
  "is_secret": false
}
```

**Response**: `200 OK`
```json
{
  "key": "api.timeout",
  "value": "30s",
  "is_secret": false
}
```

### Get Configuration Value

**Endpoint**: `GET /api/configurations/{key}/value`

**Description**: Get just the value of a configuration.

**Response**: `200 OK`
```json
{
  "value": "30s"
}
```

### Get Boolean Value

**Endpoint**: `GET /api/configurations/{key}/bool`

**Description**: Get a configuration value parsed as a boolean.

**Query Parameters**:
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `default` | boolean | No | false | Default value if not found |

**Response**: `200 OK`
```json
{
  "value": true
}
```

### Get Number Value

**Endpoint**: `GET /api/configurations/{key}/number`

**Description**: Get a configuration value parsed as a number.

**Query Parameters**:
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `default` | number | No | 0 | Default value if not found |

**Response**: `200 OK`
```json
{
  "value": 42
}
```

### Get JSON Value

**Endpoint**: `GET /api/configurations/{key}/json`

**Description**: Get a configuration value parsed as JSON.

**Response**: `200 OK`
```json
{
  "nested": {
    "key": "value"
  }
}
```

---

## Contacts API

Manage contacts with evaluations, tags, and relationships.

### List Contacts

**Endpoint**: `GET /api/contacts`

**Description**: Get paginated list of contacts with optional search.

**Query Parameters**:
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `page` | integer | No | 1 | Page number |
| `pageSize` | integer | No | 20 | Items per page |
| `search` | string | No | - | Search query |

**Response**: `200 OK`
```json
{
  "data": [
    {
      "contact_id": "uuid",
      "name": "John Doe",
      "email": "john@example.com",
      "phone": "+1234567890",
      "tags": ["friend", "colleague"],
      "importance": 5,
      "closeness": 8,
      "fondness": 7
    }
  ],
  "pagination": { ... }
}
```

### Get Contact

**Endpoint**: `GET /api/contacts/{id}`

**Description**: Get a single contact with all its relations.

**Response**: `200 OK`
```json
{
  "contact_id": "uuid",
  "name": "John Doe",
  "email": "john@example.com",
  "phone": "+1234567890",
  "birthday": "1990-01-01",
  "notes": "Met at conference",
  "extras": { ... },
  "evaluation": {
    "closeness": 8,
    "importance": 5,
    "fondness": 7
  },
  "tags": [ ... ],
  "known_names": [ ... ],
  "rooms": [ ... ],
  "sources": [ ... ]
}
```

### Create Contact

**Endpoint**: `POST /api/contacts`

**Description**: Create a new contact.

**Request Body**:
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "phone": "+1234567890",
  "birthday": "1990-01-01",
  "notes": "Met at conference"
}
```

**Response**: `201 Created`
```json
{
  "contact_id": "uuid",
  "name": "John Doe",
  "email": "john@example.com"
}
```

### Update Contact

**Endpoint**: `PUT /api/contacts/{id}`

**Description**: Update a contact's basic information.

**Request Body**:
```json
{
  "name": "John Doe Jr.",
  "email": "john.jr@example.com",
  "notes": "Updated notes"
}
```

**Response**: `200 OK`
```json
{
  "message": "Contact updated successfully"
}
```

### Delete Contact

**Endpoint**: `DELETE /api/contacts/{id}`

**Description**: Delete a contact.

**Response**: `200 OK`
```json
{
  "message": "Contact deleted successfully"
}
```

### Update Contact Evaluation

**Endpoint**: `PUT /api/contacts/{id}/evaluation`

**Description**: Update or create a contact's evaluation scores.

**Request Body**:
```json
{
  "closeness": 8,
  "importance": 5,
  "fondness": 7
}
```

**Response**: `200 OK`
```json
{
  "message": "Contact evaluation updated successfully"
}
```

### Merge Contacts

**Endpoint**: `POST /api/contacts/merge`

**Description**: Merge two contacts together.

**Request Body**:
```json
{
  "source_contact_id": "uuid",
  "target_contact_id": "uuid"
}
```

**Response**: `200 OK`
```json
{
  "message": "Contacts merged successfully"
}
```

### Batch Update Contacts

**Endpoint**: `POST /api/contacts/batch`

**Description**: Update multiple contacts atomically.

**Request Body**:
```json
[
  {
    "contact_id": "uuid",
    "name": "Updated Name",
    "email": "new@example.com"
  }
]
```

**Response**: `200 OK`
```json
[
  {
    "contact_id": "uuid",
    "success": true
  }
]
```

### Refresh Contact Statistics

**Endpoint**: `POST /api/contacts/stats/refresh`

**Description**: Update message statistics for all contacts.

**Response**: `200 OK`
```json
{
  "message": "Stats refreshed successfully"
}
```

### Get Contact Tags

**Endpoint**: `GET /api/contacts/{id}/tags`

**Description**: Get all tags for a contact.

**Response**: `200 OK`
```json
[
  {
    "id": "uuid",
    "name": "friend",
    "created_at": "2024-01-01T00:00:00Z"
  }
]
```

### Add Tag to Contact

**Endpoint**: `POST /api/contacts/{id}/tags`

**Description**: Add a tag to a contact (creates tag if it doesn't exist).

**Request Body**:
```json
{
  "name": "colleague"
}
```

**Response**: `200 OK`
```json
{
  "id": "uuid",
  "name": "colleague"
}
```

### Remove Tag from Contact

**Endpoint**: `DELETE /api/contacts/{id}/tags/{tagId}`

**Description**: Remove a tag from a contact.

**Response**: `200 OK`
```json
{
  "message": "Tag removed successfully"
}
```

### List All Contact Tags

**Endpoint**: `GET /api/contact-tags`

**Description**: Get all contact tag names in the system.

**Response**: `200 OK`
```json
[
  {
    "id": "uuid",
    "name": "friend"
  }
]
```

### List Contact Sources

**Endpoint**: `GET /api/contact-sources`

**Description**: Get all contact sources in the system.

**Response**: `200 OK`
```json
[
  {
    "id": "uuid",
    "source_id": "telegram_12345",
    "source_name": "Telegram",
    "contact_id": "uuid"
  }
]
```

### Create Contact Source

**Endpoint**: `POST /api/contact-sources`

**Description**: Create a new contact source.

**Request Body**:
```json
{
  "source_id": "telegram_12345",
  "source_name": "Telegram",
  "contact_id": "uuid"
}
```

**Response**: `201 Created`
```json
{
  "id": "uuid",
  "source_id": "telegram_12345",
  "source_name": "Telegram"
}
```

### Update Contact Source

**Endpoint**: `PUT /api/contact-sources/{id}`

**Description**: Update a contact source.

**Request Body**:
```json
{
  "source_id": "telegram_67890",
  "source_name": "Telegram"
}
```

**Response**: `200 OK`

### Delete Contact Source

**Endpoint**: `DELETE /api/contact-sources/{id}`

**Description**: Delete a contact source.

**Response**: `200 OK`
```json
{
  "message": "Contact source deleted successfully"
}
```

---

## Dashboard API

Get aggregated statistics and insights.

### Get Dashboard Statistics

**Endpoint**: `GET /api/dashboard/stats`

**Description**: Get comprehensive statistics for contacts, sessions, bookmarks, browser history, and recent items.

**Response**: `200 OK`
```json
{
  "contacts": {
    "total": 150,
    "recent": 10
  },
  "sessions": {
    "total": 500,
    "today": 5
  },
  "bookmarks": {
    "total": 1000,
    "unprocessed": 25
  },
  "browser_history": {
    "total": 5000,
    "today": 50
  },
  "recent_items": [ ... ]
}
```

---

## Entities API

Manage structured entities with relationships and references.

### List Entities

**Endpoint**: `GET /api/entities`

**Description**: List all entities with optional filters.

**Query Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `type` | string | No | Entity type filter |
| `updated_since` | string | No | Updated since timestamp |

**Response**: `200 OK`
```json
[
  {
    "entity_id": "uuid",
    "name": "Example Entity",
    "type": "concept",
    "description": "Description here",
    "properties": { ... },
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

### Create Entity

**Endpoint**: `POST /api/entities`

**Description**: Create a new entity.

**Request Body**:
```json
{
  "name": "Example Entity",
  "type": "concept",
  "description": "Description here",
  "properties": {
    "custom_field": "value"
  }
}
```

**Response**: `201 Created`
```json
{
  "entity_id": "uuid",
  "name": "Example Entity",
  "type": "concept"
}
```

### Get Entity

**Endpoint**: `GET /api/entities/{id}`

**Description**: Get a single entity by ID.

**Response**: `200 OK`
```json
{
  "entity_id": "uuid",
  "name": "Example Entity",
  "type": "concept",
  "description": "Description here",
  "properties": { ... },
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### Update Entity

**Endpoint**: `PUT /api/entities/{id}`

**Description**: Update an entity's fields.

**Request Body**:
```json
{
  "name": "Updated Entity",
  "description": "New description",
  "properties": { ... }
}
```

**Response**: `200 OK`

### Delete Entity

**Endpoint**: `DELETE /api/entities/{id}`

**Description**: Soft delete an entity.

**Response**: `204 No Content`

### Search Entities

**Endpoint**: `GET /api/entities/search`

**Description**: Search entities by name with optional type filter.

**Query Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `q` or `query` | string | Yes | Search query |
| `type` | string | No | Entity type filter |

**Response**: `200 OK`
```json
[
  {
    "entity_id": "uuid",
    "name": "Example Entity",
    "type": "concept"
  }
]
```

### List Deleted Entities

**Endpoint**: `GET /api/entities/deleted`

**Description**: List all soft-deleted entities.

**Response**: `200 OK`

### Get Entity References

**Endpoint**: `GET /api/entity-references`

**Description**: Get references to an entity with optional source type filter.

**Query Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `entity_id` | UUID | Yes | Entity ID |
| `source_type` | string | No | Source type filter |

**Response**: `200 OK`
```json
[
  {
    "id": "uuid",
    "source_type": "note",
    "source_id": "uuid",
    "entity_id": "uuid",
    "reference_text": "[[entity-name]]",
    "position": 42,
    "created_at": "2024-01-01T00:00:00Z"
  }
]
```

### Parse Entity References

**Endpoint**: `POST /api/entity-references/parse`

**Description**: Parse [[entity]] references from content text.

**Request Body**:
```json
{
  "content": "This mentions [[entity-name]] and [[another-entity]]"
}
```

**Response**: `200 OK`
```json
[
  {
    "original": "[[entity-name]]",
    "entity_name": "entity-name",
    "display_text": "entity-name"
  }
]
```

### Create Entity Relationship

**Endpoint**: `POST /api/entities/relationship`

**Description**: Create a new relationship between an entity and another resource.

**Request Body**:
```json
{
  "entity_id": "uuid",
  "related_type": "note",
  "related_id": "uuid",
  "relationship_type": "references",
  "metadata": { ... }
}
```

**Response**: `201 Created`

### Get Entity Relationships

**Endpoint**: `GET /api/entities/{id}/relationships`

**Description**: Get relationships for an entity with optional type filters.

**Query Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `related_type` | string | No | Related type filter |
| `relationship_type` | string | No | Relationship type filter |

**Response**: `200 OK`
```json
[
  {
    "id": "uuid",
    "entity_id": "uuid",
    "related_type": "note",
    "related_id": "uuid",
    "relationship_type": "references",
    "metadata": { ... }
  }
]
```

### Delete Entity Relationship

**Endpoint**: `DELETE /api/entities/relationship/{id}`

**Description**: Delete an entity relationship by ID.

**Response**: `204 No Content`

---

## Items API

Manage generic items with tags and embeddings.

### List Items

**Endpoint**: `GET /api/items`

**Description**: Get paginated list of items.

**Query Parameters**:
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `page` | integer | No | 1 | Page number |
| `limit` | integer | No | 10 | Items per page (max 100) |

**Response**: `200 OK`
```json
{
  "data": [
    {
      "id": "uuid",
      "title": "Item Title",
      "tags": ["tag1", "tag2"],
      "created": 1234567890,
      "modified": 1234567890
    }
  ],
  "pagination": { ... }
}
```

### Create Item

**Endpoint**: `POST /api/items`

**Description**: Create a new item with tags.

**Request Body**:
```json
{
  "title": "Item Title",
  "contents": "Item contents here",
  "tags": ["tag1", "tag2"]
}
```

**Response**: `201 Created`
```json
{
  "data": {
    "id": "uuid",
    "title": "Item Title",
    "contents": "Item contents here",
    "tags": ["tag1", "tag2"],
    "created": 1234567890,
    "modified": 1234567890
  },
  "message": "Item created successfully"
}
```

### Get Item

**Endpoint**: `GET /api/items/{id}`

**Description**: Get a single item with tags.

**Response**: `200 OK`
```json
{
  "id": "uuid",
  "title": "Item Title",
  "contents": "Item contents here",
  "tags": ["tag1", "tag2"],
  "created": 1234567890,
  "modified": 1234567890
}
```

### Update Item

**Endpoint**: `PUT /api/items/{id}`

**Description**: Update an item's information.

**Request Body**:
```json
{
  "title": "Updated Title",
  "contents": "Updated contents"
}
```

**Response**: `200 OK`

### Delete Item

**Endpoint**: `DELETE /api/items/{id}`

**Description**: Delete an item and all related data.

**Response**: `200 OK`
```json
{
  "message": "Item deleted successfully"
}
```

### Search Items

**Endpoint**: `GET /api/items/search`

**Description**: Perform vector similarity search on items.

**Query Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `q` | string | Yes | Search query (embedding as JSON array) |

**Response**: `200 OK`
```json
{
  "data": [
    {
      "id": "uuid",
      "title": "Item Title",
      "tags": ["tag1"]
    }
  ]
}
```

### Get Item Tags

**Endpoint**: `GET /api/items/{id}/tags`

**Description**: Get all tags for a specific item.

**Response**: `200 OK`
```json
["tag1", "tag2", "tag3"]
```

### Update Item Tags

**Endpoint**: `PUT /api/items/{id}/tags`

**Description**: Update the tags for a specific item.

**Request Body**:
```json
["new-tag1", "new-tag2"]
```

**Response**: `200 OK`
```json
{
  "message": "Tags updated successfully"
}
```

---

## Messages API

Query and search messages from conversations.

### Get Message

**Endpoint**: `GET /api/messages/{id}`

**Description**: Get a single message by ID.

**Response**: `200 OK`
```json
{
  "message_id": "uuid",
  "content": "Message text",
  "sender_id": "uuid",
  "room_id": "uuid",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### Get Messages by Room

**Endpoint**: `GET /api/rooms/{roomId}/messages`

**Description**: Get paginated messages for a specific room.

**Query Parameters**:
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `page` | integer | No | 1 | Page number |
| `pageSize` | integer | No | 100 | Items per page |

**Response**: `200 OK`
```json
{
  "data": [ ... ],
  "pagination": { ... }
}
```

### Get All Message Contents

**Endpoint**: `GET /api/messages/content`

**Description**: Get all message contents (for indexing/processing).

**Response**: `200 OK`
```json
{
  "data": [
    {
      "message_id": "uuid",
      "content": "Message text"
    }
  ]
}
```

### Get Message Text Representations

**Endpoint**: `GET /api/messages/{id}/text-representations`

**Description**: Get different text representations of a message.

**Response**: `200 OK`
```json
[
  {
    "type": "plain",
    "content": "Message text"
  }
]
```

### Search Messages

**Endpoint**: `GET /api/messages/search`

**Description**: Search messages by query.

**Query Parameters**:
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `q` | string | Yes | - | Search query |
| `page` | integer | No | 1 | Page number |
| `pageSize` | integer | No | 50 | Items per page |

**Response**: `200 OK`
```json
{
  "data": [ ... ],
  "pagination": { ... }
}
```

### Get Messages by IDs

**Endpoint**: `POST /api/messages/content`

**Description**: Retrieve multiple messages by their IDs.

**Request Body**:
```json
{
  "message_ids": ["uuid1", "uuid2", "uuid3"]
}
```

**Response**: `200 OK`
```json
[
  {
    "message_id": "uuid1",
    "content": "Message text"
  }
]
```

---

## Notes API

Manage personal notes with tags and entity relationships.

### List Notes

**Endpoint**: `GET /api/notes`

**Description**: Get paginated list of notes with optional search.

**Query Parameters**:
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `page` | integer | No | 1 | Page number |
| `pageSize` | integer | No | 12 | Items per page |
| `searchQuery` | string | No | - | Search query |

**Response**: `200 OK`
```json
{
  "notes": [
    {
      "id": "uuid",
      "title": "Note Title",
      "tags": ["tag1", "tag2"],
      "created": 1234567890,
      "modified": 1234567890
    }
  ],
  "totalPages": 5
}
```

### Create Note

**Endpoint**: `POST /api/notes`

**Description**: Create a new note with tags and entity relationships.

**Request Body**:
```json
{
  "title": "Note Title",
  "contents": "Note contents with [[entity-reference]]",
  "tags": ["tag1", "tag2"]
}
```

**Response**: `201 Created`
```json
{
  "id": "uuid",
  "title": "Note Title",
  "contents": "Note contents",
  "processedContents": "Processed with entity links",
  "tags": ["tag1", "tag2"],
  "created": 1234567890,
  "modified": 1234567890,
  "entity_id": "uuid"
}
```

### Get Note

**Endpoint**: `GET /api/notes/{id}`

**Description**: Get a single note with tags and processed content.

**Response**: `200 OK`
```json
{
  "id": "uuid",
  "title": "Note Title",
  "contents": "Note contents",
  "processedContents": "Processed with entity links",
  "tags": ["tag1", "tag2"],
  "created": 1234567890,
  "modified": 1234567890
}
```

### Update Note

**Endpoint**: `PUT /api/notes/{id}`

**Description**: Update a note's information and tags.

**Request Body**:
```json
{
  "title": "Updated Title",
  "contents": "Updated contents",
  "tags": ["new-tag"]
}
```

**Response**: `200 OK`

### Delete Note

**Endpoint**: `DELETE /api/notes/{id}`

**Description**: Delete a note and all related data.

**Response**: `200 OK`
```json
{
  "message": "Note deleted successfully"
}
```

### Search Notes

**Endpoint**: `GET /api/notes/search`

**Description**: Perform vector similarity search on notes.

**Query Parameters**:
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `q` | string | Yes | - | Search query |
| `strategy` | string | No | qa-v2-passage | Search strategy |

**Response**: `200 OK`
```json
[
  {
    "id": "uuid",
    "title": "Note Title",
    "tags": ["tag1"]
  }
]
```

### List All Tags

**Endpoint**: `GET /api/tags`

**Description**: Get all available tags in the system.

**Response**: `200 OK`
```json
[
  {
    "id": "uuid",
    "name": "tag1",
    "created": 1234567890
  }
]
```

---

## Observations API

Store and query Q&A feedback and observations.

### Store Feedback

**Endpoint**: `POST /api/observations/feedback`

**Description**: Store feedback for Q&A and optionally delete the content reference.

**Request Body**:
```json
{
  "bookmark_id": "uuid",
  "question": "What is this about?",
  "answer": "This is about...",
  "feedback": "helpful",
  "content_reference_id": "uuid",
  "delete_reference": false
}
```

**Response**: `201 Created`
```json
{
  "observation_id": "uuid",
  "created_at": "2024-01-01T00:00:00Z"
}
```

### Get Feedback Statistics

**Endpoint**: `GET /api/bookmarks/{id}/feedback`

**Description**: Get aggregated feedback statistics for a bookmark.

**Response**: `200 OK`
```json
{
  "helpful_count": 15,
  "not_helpful_count": 2,
  "neutral_count": 3,
  "total_feedback": 20
}
```

---

## Rooms API

Manage conversation rooms and their messages.

### List Rooms

**Endpoint**: `GET /api/rooms`

**Description**: Get paginated list of rooms.

**Query Parameters**:
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `page` | integer | No | 1 | Page number |
| `pageSize` | integer | No | 10 | Items per page |
| `searchText` | string | No | - | Search query |

**Response**: `200 OK`
```json
{
  "data": [
    {
      "room_id": "uuid",
      "name": "Room Name",
      "participant_count": 5,
      "message_count": 150,
      "last_message_at": "2024-01-01T00:00:00Z"
    }
  ],
  "pagination": { ... }
}
```

### Get Room Details

**Endpoint**: `GET /api/rooms/{id}`

**Description**: Get detailed information about a room.

**Response**: `200 OK`
```json
{
  "room_id": "uuid",
  "name": "Room Name",
  "participants": [ ... ],
  "message_count": 150,
  "created_at": "2024-01-01T00:00:00Z"
}
```

### Get Room Messages

**Endpoint**: `GET /api/rooms/{id}/messages`

**Description**: Get paginated messages for a room.

**Query Parameters**:
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `page` | integer | No | 1 | Page number |
| `pageSize` | integer | No | 50 | Items per page |
| `beforeMessageId` | UUID | No | - | Get messages before this ID |

**Response**: `200 OK`
```json
{
  "data": [ ... ],
  "pagination": { ... }
}
```

### Search Room Messages

**Endpoint**: `GET /api/rooms/{id}/messages/search`

**Description**: Search messages within a room.

**Query Parameters**:
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `searchText` | string | Yes | - | Search query |
| `page` | integer | No | 1 | Page number |
| `pageSize` | integer | No | 50 | Items per page |

**Response**: `200 OK`

### Set Room Name

**Endpoint**: `PUT /api/rooms/{id}/name`

**Description**: Update a room's name.

**Request Body**:
```json
{
  "name": "New Room Name"
}
```

**Response**: `200 OK`
```json
{
  "status": "success"
}
```

### Get Room Sessions Count

**Endpoint**: `GET /api/rooms/{id}/sessions/count`

**Description**: Get the number of sessions in a room.

**Response**: `200 OK`
```json
{
  "count": 42
}
```

---

## Search API

Unified search across all content types.

### Search All

**Endpoint**: `GET /api/search`

**Description**: Search across contacts, conversations, bookmarks, browser history, and notes.

**Query Parameters**:
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `q` or `query` | string | Yes | - | Search query |
| `limit` | integer | No | 50 | Result limit |
| `exact_match_weight` | float | No | 5.0 | Exact match weight |
| `similarity_weight` | float | No | 2.0 | Similarity weight |
| `recency_weight` | float | No | 1.0 | Recency weight |

**Response**: `200 OK`
```json
[
  {
    "type": "bookmark",
    "id": "uuid",
    "title": "Result Title",
    "score": 0.95,
    "snippet": "Matching content..."
  }
]
```

### Advanced Search

**Endpoint**: `POST /api/search/advanced`

**Description**: LLM-powered search that synthesizes answers from bookmarks.

**Request Body**:
```json
{
  "query": "What are the best practices for API design?"
}
```

**Response**: `200 OK`
```json
{
  "query": "What are the best practices for API design?",
  "answer": "Based on the bookmarks, best practices include...",
  "sources": [
    {
      "bookmark_id": "uuid",
      "title": "API Design Guide",
      "url": "https://example.com"
    }
  ],
  "processing_time_ms": 1500
}
```

---

## Sessions API

Query conversation sessions and timelines.

### Search Sessions

**Endpoint**: `GET /api/sessions/search`

**Description**: Search sessions by query.

**Query Parameters**:
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `q` | string | Yes | - | Search query |
| `limit` | integer | No | 10 | Result limit |

**Response**: `200 OK`
```json
[
  {
    "session_id": "uuid",
    "room_id": "uuid",
    "start_time": "2024-01-01T00:00:00Z",
    "message_count": 25
  }
]
```

### Get Room Sessions

**Endpoint**: `GET /api/rooms/{id}/sessions`

**Description**: Get all sessions for a room.

**Response**: `200 OK`
```json
[
  {
    "session_id": "uuid",
    "start_time": "2024-01-01T00:00:00Z",
    "end_time": "2024-01-01T01:00:00Z",
    "message_count": 25
  }
]
```

### Search Contact Sessions

**Endpoint**: `GET /api/contacts/{id}/sessions/search`

**Description**: Search sessions involving a specific contact.

**Query Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `q` | string | Yes | Search query |

**Response**: `200 OK`

### Get Timeline

**Endpoint**: `GET /api/rooms/{id}/timeline`

**Description**: Get chronological timeline of sessions in a room.

**Response**: `200 OK`
```json
[
  {
    "session_id": "uuid",
    "start_time": "2024-01-01T00:00:00Z",
    "duration_minutes": 30,
    "message_count": 25
  }
]
```

### Get Session Messages

**Endpoint**: `GET /api/sessions/{id}/messages`

**Description**: Get all messages in a session.

**Response**: `200 OK`
```json
{
  "session_id": "uuid",
  "messages": [ ... ]
}
```

---

## Social Posts API

Manage and publish social media posts to Twitter and Bluesky.

### List Social Posts

**Endpoint**: `GET /api/social/posts`

**Description**: Get paginated social posts with optional status filter.

**Query Parameters**:
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `status` | string | No | - | Filter by status (pending, completed, partial, failed) |
| `page` | integer | No | 1 | Page number |
| `limit` | integer | No | 10 | Items per page |

**Response**: `200 OK`
```json
{
  "data": [
    {
      "post_id": "uuid",
      "content": "Post content here",
      "status": "completed",
      "twitter_id": "123456789",
      "bluesky_uri": "at://...",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "pagination": { ... }
}
```

### Get Social Post

**Endpoint**: `GET /api/social/posts/{id}`

**Description**: Get a single social post by its ID.

**Response**: `200 OK`
```json
{
  "post_id": "uuid",
  "content": "Post content",
  "status": "completed",
  "twitter_id": "123456789",
  "bluesky_uri": "at://..."
}
```

### Create Social Post

**Endpoint**: `POST /api/social/posts`

**Description**: Create a new social post and publish to Twitter and Bluesky.

**Request Body**:
```json
{
  "content": "This is a new post!",
  "platforms": ["twitter", "bluesky"]
}
```

**Response**: `201 Created`
```json
{
  "post_id": "uuid",
  "status": "completed",
  "twitter": {
    "success": true,
    "id": "123456789"
  },
  "bluesky": {
    "success": true,
    "uri": "at://..."
  }
}
```

### Update Post Status

**Endpoint**: `PUT /api/social/posts/{id}/status`

**Description**: Update the status of a social post.

**Request Body**:
```json
{
  "status": "completed",
  "twitter_id": "123456789",
  "bluesky_uri": "at://..."
}
```

**Response**: `200 OK`

### Delete Social Post

**Endpoint**: `DELETE /api/social/posts/{id}`

**Description**: Delete a social post by ID.

**Response**: `204 No Content`

### Check Credentials

**Endpoint**: `GET /api/social/credentials`

**Description**: Verify Twitter and Bluesky credentials are valid.

**Response**: `200 OK`
```json
{
  "twitter": {
    "configured": true,
    "valid": true
  },
  "bluesky": {
    "configured": true,
    "valid": true
  }
}
```

### Update Twitter Tokens

**Endpoint**: `PUT /api/social/twitter/tokens`

**Description**: Store Twitter OAuth tokens from authorization flow.

**Request Body**:
```json
{
  "access_token": "...",
  "refresh_token": "...",
  "expires_at": "2024-12-31T23:59:59Z"
}
```

**Response**: `200 OK`

### Initiate Twitter Auth

**Endpoint**: `POST /api/social/twitter/auth`

**Description**: Generate Twitter OAuth authorization URL.

**Response**: `200 OK`
```json
{
  "auth_url": "https://twitter.com/i/oauth2/authorize?...",
  "state": "random_state_value"
}
```

### Handle Twitter Callback

**Endpoint**: `POST /api/social/twitter/callback`

**Description**: Process OAuth callback and exchange code for tokens.

**Request Body**:
```json
{
  "code": "oauth_code",
  "state": "random_state_value"
}
```

**Response**: `200 OK`
```json
{
  "status": "success"
}
```

---

## Tags API

Manage tags for items.

### List Tags

**Endpoint**: `GET /api/tags`

**Description**: Get all tags, optionally with usage counts.

**Query Parameters**:
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `includeUsage` | boolean | No | false | Include usage counts |

**Response**: `200 OK`
```json
{
  "tags": [
    {
      "id": "uuid",
      "name": "technology",
      "created": 1234567890,
      "modified": 1234567890,
      "lastActivity": 1234567890,
      "usageCount": 42
    }
  ]
}
```

### Get Tag

**Endpoint**: `GET /api/tags/{name}`

**Description**: Get a single tag by name.

**Response**: `200 OK`
```json
{
  "id": "uuid",
  "name": "technology",
  "created": 1234567890,
  "modified": 1234567890
}
```

### Get Items by Tag

**Endpoint**: `GET /api/tags/{name}/items`

**Description**: Get all items with a specific tag.

**Query Parameters**:
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `page` | integer | No | 1 | Page number |
| `limit` | integer | No | 20 | Items per page |

**Response**: `200 OK`
```json
{
  "data": [ ... ],
  "pagination": { ... }
}
```

### Add Tag to Item

**Endpoint**: `POST /api/tags/{itemId}/{tagName}`

**Description**: Add a tag to an item.

**Response**: `200 OK`
```json
{
  "message": "Tag added successfully"
}
```

### Remove Tag from Item

**Endpoint**: `DELETE /api/tags/{itemId}/{tagName}`

**Description**: Remove a tag from an item.

**Response**: `200 OK`
```json
{
  "message": "Tag removed successfully"
}
```

### Delete Tag

**Endpoint**: `DELETE /api/tags/{name}`

**Description**: Delete a tag and all its associations.

**Response**: `200 OK`
```json
{
  "message": "Tag deleted successfully"
}
```

---

## Logseq Sync API

Synchronize with Logseq git repository.

### Synchronize

**Endpoint**: `POST /api/sync/logseq`

**Description**: Trigger full Logseq sync between git repository and database.

**Response**: `200 OK`
```json
{
  "pages_processed": 150,
  "pages_created": 10,
  "pages_updated": 25,
  "pages_skipped": 115,
  "entities_processed": 200,
  "entities_created": 15,
  "entities_updated": 30,
  "entities_skipped": 155,
  "errors": []
}
```

### Perform Sync Check

**Endpoint**: `GET /api/sync/logseq/check`

**Description**: Compare all files in Logseq folder with database entries.

**Response**: `200 OK`
```json
{
  "missing_in_db": [
    {
      "entity_id": "uuid",
      "name": "Entity Name",
      "type": "concept"
    }
  ],
  "missing_in_git": [ ... ],
  "out_of_sync": [ ... ]
}
```

### Force Update Git from DB

**Endpoint**: `POST /api/sync/logseq/force-git`

**Description**: Force update a git file with data from the database.

**Request Body**:
```json
{
  "entity_id": "uuid"
}
```

**Response**: `200 OK`
```json
{
  "message": "File updated successfully"
}
```

### Force Update DB from Git

**Endpoint**: `POST /api/sync/logseq/force-db`

**Description**: Force update database entry with data from a git file.

**Request Body**:
```json
{
  "page_path": "pages/concept.md"
}
```

**Response**: `200 OK`
```json
{
  "entity_id": "uuid",
  "name": "Concept",
  "type": "concept"
}
```

---

## Utility API

System utilities and debugging.

### Get Debug Info

**Endpoint**: `GET /api/debug`

**Description**: Returns system debug information.

**Response**: `200 OK`
```json
{
  "version": "1.0.0",
  "database": {
    "connected": true,
    "pool_size": 10
  },
  "services": {
    "ollama": "connected",
    "postgres": "connected"
  }
}
```

### Cleanup Stale Conversations

**Endpoint**: `POST /api/conversations/cleanup`

**Description**: Remove stale sessions.

**Response**: `200 OK`
```json
{
  "deleted_count": 42
}
```

---

## Error Codes

The API uses standard HTTP status codes:

### Success Codes

| Code | Description |
|------|-------------|
| `200 OK` | Request succeeded |
| `201 Created` | Resource created successfully |
| `204 No Content` | Request succeeded with no response body |
| `302 Found` | Temporary redirect |

### Client Error Codes

| Code | Description |
|------|-------------|
| `400 Bad Request` | Invalid request body or parameters |
| `404 Not Found` | Resource not found |
| `422 Unprocessable Entity` | Validation error |

### Server Error Codes

| Code | Description |
|------|-------------|
| `500 Internal Server Error` | Server error occurred |
| `503 Service Unavailable` | Service temporarily unavailable |

### Error Response Format

```json
{
  "error": "Error type",
  "message": "Detailed error message",
  "details": [
    {
      "field": "email",
      "message": "Invalid email format"
    }
  ]
}
```

---

## Rate Limiting

> **Note**: Rate limiting should be implemented at the reverse proxy or API gateway level based on your requirements.

---

## Backward Compatibility

The API maintains backward compatibility with legacy endpoints:

- `/api/configuration/get` → `/api/configurations/{key}`
- `/api/configuration/set` → `/api/configurations/{key}/value`
- `/api/microlog/*` → `/api/social/posts/*`
- Legacy query parameters like `levenshteinWeight` are supported alongside new names

---

## Changelog

### Version 1.0.0
- Initial API release
- Support for bookmarks, contacts, notes, items, and more
- Vector search capabilities
- Social media integration
- Logseq synchronization
