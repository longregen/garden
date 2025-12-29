# HTTP Handlers Documentation

## Overview

This document provides comprehensive documentation for all HTTP handlers in the Garden project. The handlers are organized as REST API endpoints using the Chi router framework. All handlers follow a clean architecture pattern with separation between HTTP layer, use cases, and domain logic.

**Base Path**: All endpoints are prefixed with `/api`

**Common Patterns**:
- Uses UUID for resource identifiers
- Supports pagination where applicable
- Returns JSON responses
- Uses standard HTTP status codes
- Includes backwards-compatible legacy endpoints where applicable

---

## Table of Contents

1. [Bookmark Handler](#bookmark-handler)
2. [Browser History Handler](#browser-history-handler)
3. [Category Handler](#category-handler)
4. [Configuration Handler](#configuration-handler)
5. [Contact Handler](#contact-handler)
6. [Dashboard Handler](#dashboard-handler)
7. [Entity Handler](#entity-handler)
8. [Item Handler](#item-handler)
9. [Logseq Handler](#logseq-handler)
10. [Message Handler](#message-handler)
11. [Note Handler](#note-handler)
12. [Observation Handler](#observation-handler)
13. [Room Handler](#room-handler)
14. [Search Handler](#search-handler)
15. [Session Handler](#session-handler)
16. [Social Post Handler](#social-post-handler)
17. [Tag Handler](#tag-handler)
18. [Utility Handler](#utility-handler)

---

## Bookmark Handler

**Purpose**: Manages web bookmarks with content fetching, processing, embedding generation, and Q&A functionality.

**Base Route**: `/api/bookmarks`

### Endpoints

#### List Bookmarks
- **Method**: `GET /api/bookmarks`
- **Description**: Get filtered and paginated bookmarks
- **Query Parameters**:
  - `categoryId` (string, optional) - Filter by category UUID
  - `searchQuery` (string, optional) - Search query text
  - `startCreationDate` (string, optional) - Start date (RFC3339)
  - `endCreationDate` (string, optional) - End date (RFC3339)
  - `page` (int, optional, default: 1) - Page number
  - `limit` (int, optional, default: 10) - Page size
- **Response**: `PaginatedResponse[BookmarkWithTitle]`
- **Status Codes**: 200 (success), 400 (invalid parameters), 500 (server error)

#### Get Random Bookmark
- **Method**: `GET /api/bookmarks/random`
- **Description**: Redirect to a random bookmark
- **Response**: 302 redirect
- **Status Codes**: 302 (redirect), 500 (server error)

#### Get Bookmark Details
- **Method**: `GET /api/bookmarks/{id}`
- **Description**: Get complete bookmark details with all relations
- **Path Parameters**:
  - `id` (UUID) - Bookmark ID
- **Response**: `BookmarkDetails`
- **Status Codes**: 200 (success), 400 (invalid ID), 500 (server error)

#### Search Bookmarks
- **Method**: `GET /api/bookmarks/search`
- **Description**: Perform vector similarity search on bookmarks
- **Query Parameters**:
  - `query` (string, required) - Search query
  - `strategy` (string, optional, default: "qa-v2-passage") - Search strategy
- **Response**: `[]BookmarkWithTitle`
- **Status Codes**: 200 (success), 400 (missing query), 500 (server error)

#### Get Missing HTTP Responses
- **Method**: `GET /api/bookmarks/missing/http`
- **Description**: Get bookmarks without HTTP responses
- **Response**: `[]Bookmark`
- **Status Codes**: 200 (success), 500 (server error)
- **Backwards-compatible alias**: `GET /api/bookmarks/missing-http`

#### Get Missing Reader Content
- **Method**: `GET /api/bookmarks/missing/reader`
- **Description**: Get bookmarks without reader-processed content
- **Response**: `[]Bookmark`
- **Status Codes**: 200 (success), 500 (server error)
- **Backwards-compatible alias**: `GET /api/bookmarks/missing-reader`

#### Update Question
- **Method**: `PUT /api/bookmarks/{id}/question`
- **Description**: Update a question and answer content reference
- **Path Parameters**:
  - `id` (UUID) - Bookmark ID
- **Request Body**: `UpdateQuestionInput`
- **Response**: `{"message": "Question updated successfully"}`
- **Status Codes**: 200 (success), 400 (invalid input), 500 (server error)
- **Backwards-compatible alias**: `PUT /api/bookmarks/{id}/update-question`

#### Delete Question
- **Method**: `DELETE /api/bookmarks/{id}/question/{refId}`
- **Description**: Delete a question and answer content reference
- **Path Parameters**:
  - `id` (UUID) - Bookmark ID
  - `refId` (UUID) - Reference ID
- **Request Body**: `DeleteQuestionInput`
- **Response**: `{"message": "Question deleted successfully"}`
- **Status Codes**: 200 (success), 400 (invalid input), 500 (server error)
- **Backwards-compatible alias**: `DELETE /api/bookmarks/{id}/delete-question`

#### Fetch Content
- **Method**: `POST /api/bookmarks/{id}/fetch`
- **Description**: Fetch and store HTTP content for a bookmark
- **Path Parameters**:
  - `id` (UUID) - Bookmark ID
- **Response**: `FetchResult`
- **Status Codes**: 200 (success), 400 (invalid ID), 500 (server error)

#### Process with Lynx
- **Method**: `POST /api/bookmarks/{id}/process/lynx`
- **Description**: Process bookmark content using lynx text browser
- **Path Parameters**:
  - `id` (UUID) - Bookmark ID
- **Response**: `ProcessingResult`
- **Status Codes**: 200 (success), 400 (invalid ID), 500 (server error)
- **Backwards-compatible alias**: `POST /api/bookmarks/{id}/lynx`

#### Process with Reader
- **Method**: `POST /api/bookmarks/{id}/process/reader`
- **Description**: Process bookmark content using reader mode
- **Path Parameters**:
  - `id` (UUID) - Bookmark ID
- **Response**: `ProcessingResult`
- **Status Codes**: 200 (success), 400 (invalid ID), 500 (server error)
- **Backwards-compatible alias**: `POST /api/bookmarks/{id}/reader`

#### Create Embeddings
- **Method**: `POST /api/bookmarks/{id}/embeddings`
- **Description**: Create chunked embeddings for bookmark content
- **Path Parameters**:
  - `id` (UUID) - Bookmark ID
- **Response**: `EmbeddingResult`
- **Status Codes**: 200 (success), 400 (invalid ID), 500 (server error)
- **Backwards-compatible alias**: `POST /api/bookmarks/{id}/embed-chunks`

#### Create Summary Embedding
- **Method**: `POST /api/bookmarks/{id}/summary-embedding`
- **Description**: Create a summary and its embedding
- **Path Parameters**:
  - `id` (UUID) - Bookmark ID
- **Response**: `SummaryEmbeddingResult`
- **Status Codes**: 200 (success), 400 (invalid ID), 500 (server error)
- **Backwards-compatible alias**: `POST /api/bookmarks/{id}/embed-summary`

#### Get Title
- **Method**: `GET /api/bookmarks/{id}/title`
- **Description**: Extract and store the bookmark title
- **Path Parameters**:
  - `id` (UUID) - Bookmark ID
- **Response**: `TitleExtractionResult`
- **Status Codes**: 200 (success), 400 (invalid ID), 500 (server error)

### Error Handling
- Invalid UUID: Returns 400 with "Invalid bookmark ID"
- Missing required parameters: Returns 400 with descriptive error
- Server errors: Returns 500 with error message

---

## Browser History Handler

**Purpose**: Manages browser history entries with filtering, search, and analytics.

**Base Route**: `/api/history`

### Endpoints

#### List History
- **Method**: `GET /api/history`
- **Description**: Get filtered and paginated browser history
- **Query Parameters**:
  - `q` (string, optional) - Search query
  - `start_date` (string, optional) - Start date (RFC3339)
  - `end_date` (string, optional) - End date (RFC3339)
  - `domain` (string, optional) - Filter by domain
  - `page` (int, optional, default: 1) - Page number
  - `page_size` (int, optional, default: 10) - Page size
- **Response**: `PaginatedResponse[BrowserHistory]`
- **Status Codes**: 200 (success), 500 (server error)

#### Top Domains
- **Method**: `GET /api/history/domains`
- **Description**: Get aggregated visit counts by domain
- **Query Parameters**:
  - `limit` (int, optional, default: 10) - Number of domains to return
- **Response**: `[]DomainVisitCount`
- **Status Codes**: 200 (success), 500 (server error)

#### Recent History
- **Method**: `GET /api/history/recent`
- **Description**: Get N most recent history entries
- **Query Parameters**:
  - `limit` (int, optional, default: 20) - Number of entries to return
- **Response**: `[]BrowserHistory`
- **Status Codes**: 200 (success), 500 (server error)

### Error Handling
- Silently handles invalid date parsing (uses defaults)
- Server errors: Returns 500 with error message

---

## Category Handler

**Purpose**: Manages categories with sources and supports merging operations.

**Base Route**: `/api/categories`

### Endpoints

#### List Categories
- **Method**: `GET /api/categories`
- **Description**: Get all categories with their sources
- **Response**: `CategoriesListResponse` with pagination metadata
- **Status Codes**: 200 (success), 500 (server error)

#### Get Category
- **Method**: `GET /api/categories/{id}`
- **Description**: Get a single category by ID
- **Path Parameters**:
  - `id` (UUID) - Category ID
- **Response**: `SimpleCategoryResponse`
- **Status Codes**: 200 (success), 400 (invalid ID), 404 (not found)

#### Update Category
- **Method**: `PUT /api/categories/{id}`
- **Description**: Update a category's name
- **Path Parameters**:
  - `id` (UUID) - Category ID
- **Request Body**: `{"name": "string"}`
- **Response**: 204 No Content
- **Status Codes**: 204 (success), 400 (invalid input), 500 (server error)

#### Merge Categories
- **Method**: `POST /api/categories/merge`
- **Description**: Merge two categories together
- **Request Body**: `{"source_id": "UUID", "target_id": "UUID"}`
- **Response**: 204 No Content
- **Status Codes**: 204 (success), 400 (invalid input), 500 (server error)

#### Create Source
- **Method**: `POST /api/categories/{id}/sources`
- **Description**: Add a source to a category
- **Path Parameters**:
  - `id` (UUID) - Category ID
- **Request Body**: `{"source_uri": "string", "raw_source": {}}`
- **Response**: `CategorySourceResponse` (201)
- **Status Codes**: 201 (created), 400 (invalid input), 500 (server error)

#### Update Source
- **Method**: `PUT /api/categories/sources/{id}`
- **Description**: Update a category source
- **Path Parameters**:
  - `id` (UUID) - Source ID
- **Request Body**: `{"source_uri": "string", "raw_source": {}}`
- **Response**: 204 No Content
- **Status Codes**: 204 (success), 400 (invalid input), 500 (server error)

#### Delete Source
- **Method**: `DELETE /api/categories/sources/{id}`
- **Description**: Delete a category source
- **Path Parameters**:
  - `id` (UUID) - Source ID
- **Response**: 204 No Content
- **Status Codes**: 204 (success), 400 (invalid ID), 500 (server error)

### Error Handling
- Invalid UUID: Returns 400 with "Invalid category ID" or "Invalid source ID"
- Not found: Returns 404 with "Category not found"
- Server errors: Returns 500 with "Failed to..." message

---

## Configuration Handler

**Purpose**: Key-value configuration storage with support for secrets, prefixes, and type parsing.

**Base Route**: `/api/configurations`

### Endpoints

#### List Configurations
- **Method**: `GET /api/configurations`
- **Description**: Get all configurations with optional filtering
- **Query Parameters**:
  - `prefix` (string, optional) - Key prefix filter
  - `include_secrets` (bool, optional, default: true) - Include secret configurations
- **Response**: `[]Configuration`
- **Status Codes**: 200 (success), 400 (invalid params), 500 (server error)

#### Get Configuration
- **Method**: `GET /api/configurations/{key}`
- **Description**: Get a single configuration by its key
- **Path Parameters**:
  - `key` (string) - Configuration key
- **Response**: `Configuration`
- **Status Codes**: 200 (success), 404 (not found), 500 (server error)

#### Get by Prefix
- **Method**: `GET /api/configurations/prefix/{prefix}`
- **Description**: Get all configurations with a given key prefix
- **Path Parameters**:
  - `prefix` (string) - Key prefix
- **Query Parameters**:
  - `include_secrets` (bool, optional, default: true)
- **Response**: `[]Configuration`
- **Status Codes**: 200 (success), 400 (invalid params), 500 (server error)

#### Create Configuration
- **Method**: `POST /api/configurations`
- **Description**: Create a new configuration entry
- **Request Body**: `{"key": "string", "value": "string", "is_secret": false}`
- **Response**: `Configuration` (201)
- **Status Codes**: 201 (created), 400 (invalid input), 500 (server error)

#### Update Configuration
- **Method**: `PUT /api/configurations/{key}`
- **Description**: Update an existing configuration
- **Path Parameters**:
  - `key` (string) - Configuration key
- **Request Body**: `{"value": "string", "is_secret": false}`
- **Response**: `Configuration`
- **Status Codes**: 200 (success), 400 (invalid input), 404 (not found), 500 (server error)

#### Delete Configuration
- **Method**: `DELETE /api/configurations/{key}`
- **Description**: Delete a configuration by key
- **Path Parameters**:
  - `key` (string) - Configuration key
- **Response**: 204 No Content
- **Status Codes**: 204 (success), 500 (server error)

#### Set Configuration (Upsert)
- **Method**: `PUT /api/configurations/{key}/value`
- **Description**: Create or update a configuration value
- **Path Parameters**:
  - `key` (string) - Configuration key
- **Request Body**: `{"value": "string", "is_secret": false}`
- **Response**: `Configuration`
- **Status Codes**: 200 (success), 400 (invalid input), 500 (server error)

#### Get Value
- **Method**: `GET /api/configurations/{key}/value`
- **Description**: Get just the value of a configuration
- **Path Parameters**:
  - `key` (string) - Configuration key
- **Response**: `{"value": "string"}`
- **Status Codes**: 200 (success), 404 (not found), 500 (server error)

#### Get Bool Value
- **Method**: `GET /api/configurations/{key}/bool`
- **Description**: Get a configuration value parsed as a boolean
- **Path Parameters**:
  - `key` (string) - Configuration key
- **Query Parameters**:
  - `default` (bool, optional, default: false) - Default value if not found
- **Response**: `{"value": true/false}`
- **Status Codes**: 200 (success), 400 (invalid default), 500 (server error)

#### Get Number Value
- **Method**: `GET /api/configurations/{key}/number`
- **Description**: Get a configuration value parsed as a number
- **Path Parameters**:
  - `key` (string) - Configuration key
- **Query Parameters**:
  - `default` (number, optional, default: 0) - Default value if not found
- **Response**: `{"value": 0.0}`
- **Status Codes**: 200 (success), 400 (invalid default), 500 (server error)

#### Get JSON Value
- **Method**: `GET /api/configurations/{key}/json`
- **Description**: Get a configuration value parsed as JSON
- **Path Parameters**:
  - `key` (string) - Configuration key
- **Response**: Raw JSON object
- **Status Codes**: 200 (success), 404 (not found), 500 (server error)

### Legacy Endpoints (Backwards-Compatible)

#### Legacy Get Configuration
- **Method**: `GET /api/configuration/get?key={key}`
- **Description**: Legacy endpoint for getting configuration
- **Query Parameters**:
  - `key` (string, required) - Configuration key
- **Response**: `Configuration`

#### Legacy Get Configuration (POST)
- **Method**: `POST /api/configuration/get`
- **Description**: Legacy endpoint for getting configuration via POST
- **Request Body**: `{"key": "string"}`
- **Response**: `Configuration`

#### Legacy Set Configuration
- **Method**: `GET /api/configuration/set?key={key}&value={value}`
- **Description**: Legacy endpoint for setting configuration
- **Query Parameters**:
  - `key` (string, required)
  - `value` (string, required)
- **Response**: `Configuration`

#### Legacy Set Configuration (POST)
- **Method**: `POST /api/configuration/set`
- **Description**: Legacy endpoint for setting configuration via POST
- **Request Body**: `{"key": "string", "value": "string"}`
- **Response**: `Configuration`

### Error Handling
- Missing key parameter: Returns 400 with "key parameter is required"
- Not found: Returns 404
- Invalid boolean/number: Returns 400 with parse error
- Server errors: Returns 500

---

## Contact Handler

**Purpose**: Comprehensive contact management with tags, evaluations, sources, and relationship tracking.

**Base Route**: `/api/contacts`

### Endpoints

#### Get Contact
- **Method**: `GET /api/contacts/{id}`
- **Description**: Get a single contact with all its relations
- **Path Parameters**:
  - `id` (UUID) - Contact ID
- **Response**: `FullContactResponse` (includes tags, evaluation, known names, rooms, sources)
- **Status Codes**: 200 (success), 400 (invalid ID), 404 (not found)

#### List Contacts
- **Method**: `GET /api/contacts`
- **Description**: Get paginated list of contacts with optional search
- **Query Parameters**:
  - `page` (int, optional, default: 1) - Page number
  - `pageSize` (int, optional, default: 20) - Page size
  - `search` or `searchQuery` (string, optional) - Search query
- **Response**: `PaginatedResponse[ContactListItemResponse]`
- **Status Codes**: 200 (success), 500 (server error)

#### Create Contact
- **Method**: `POST /api/contacts`
- **Description**: Create a new contact
- **Request Body**: `CreateContactInput` (name required)
- **Response**: `Contact` (201)
- **Status Codes**: 201 (created), 400 (invalid input), 500 (server error)

#### Update Contact
- **Method**: `PUT /api/contacts/{id}`
- **Description**: Update a contact's basic information
- **Path Parameters**:
  - `id` (UUID) - Contact ID
- **Request Body**: `UpdateContactInput`
- **Response**: `{"message": "Contact updated successfully"}`
- **Status Codes**: 200 (success), 400 (invalid input), 500 (server error)

#### Delete Contact
- **Method**: `DELETE /api/contacts/{id}`
- **Description**: Delete a contact
- **Path Parameters**:
  - `id` (UUID) - Contact ID
- **Response**: `{"message": "Contact deleted successfully"}`
- **Status Codes**: 200 (success), 400 (invalid ID), 500 (server error)

#### Update Evaluation
- **Method**: `PUT /api/contacts/{id}/evaluation`
- **Description**: Update or create a contact's evaluation scores (closeness, importance, fondness)
- **Path Parameters**:
  - `id` (UUID) - Contact ID
- **Request Body**: `UpdateEvaluationInput`
- **Response**: `{"message": "Contact evaluation updated successfully"}`
- **Status Codes**: 200 (success), 400 (invalid input), 500 (server error)

#### Merge Contacts
- **Method**: `POST /api/contacts/merge`
- **Description**: Merge two contacts together
- **Request Body**: `{"source_contact_id": "UUID", "target_contact_id": "UUID"}`
- **Response**: `{"message": "Contacts merged successfully"}`
- **Status Codes**: 200 (success), 400 (invalid input), 500 (server error)

#### Batch Update
- **Method**: `POST /api/contacts/batch`
- **Description**: Update multiple contacts atomically
- **Request Body**: `[]BatchContactUpdate`
- **Response**: `[]BatchUpdateResult`
- **Status Codes**: 200 (success), 400 (invalid input), 500 (server error)

#### Refresh Statistics
- **Method**: `POST /api/contacts/stats/refresh`
- **Description**: Update message statistics for all contacts
- **Response**: `{"message": "Stats refreshed successfully"}`
- **Status Codes**: 200 (success), 500 (server error)

#### Get Contact Tags
- **Method**: `GET /api/contacts/{id}/tags`
- **Description**: Get all tags for a contact
- **Path Parameters**:
  - `id` (UUID) - Contact ID
- **Response**: `[]ContactTag`
- **Status Codes**: 200 (success), 400 (invalid ID), 500 (server error)

#### Add Tag
- **Method**: `POST /api/contacts/{id}/tags`
- **Description**: Add a tag to a contact (creates tag if it doesn't exist)
- **Path Parameters**:
  - `id` (UUID) - Contact ID
- **Request Body**: `{"name": "string"}`
- **Response**: `ContactTag`
- **Status Codes**: 200 (success), 400 (invalid input), 500 (server error)

#### Remove Tag
- **Method**: `DELETE /api/contacts/{id}/tags/{tagId}`
- **Description**: Remove a tag from a contact
- **Path Parameters**:
  - `id` (UUID) - Contact ID
  - `tagId` (UUID) - Tag ID
- **Response**: `{"message": "Tag removed successfully"}`
- **Status Codes**: 200 (success), 400 (invalid ID), 500 (server error)

#### List All Tags
- **Method**: `GET /api/contact-tags`
- **Description**: Get all contact tag names in the system
- **Response**: `[]ContactTagName`
- **Status Codes**: 200 (success), 500 (server error)

### Contact Sources Endpoints

#### List All Contact Sources
- **Method**: `GET /api/contact-sources`
- **Description**: Get all contact sources in the system
- **Response**: `[]ContactSource`
- **Status Codes**: 200 (success), 500 (server error)

#### Create Contact Source
- **Method**: `POST /api/contact-sources`
- **Description**: Create a new contact source
- **Request Body**: `{"source_id": "string", "source_name": "string"}`
- **Response**: `ContactSource` (201)
- **Status Codes**: 201 (created), 400 (invalid input), 500 (server error)

#### Update Contact Source
- **Method**: `PUT /api/contact-sources/{id}`
- **Description**: Update a contact source
- **Path Parameters**:
  - `id` (UUID) - Contact Source ID
- **Request Body**: `{"source_id": "string", "source_name": "string"}`
- **Response**: `ContactSource`
- **Status Codes**: 200 (success), 400 (invalid input), 500 (server error)

#### Delete Contact Source
- **Method**: `DELETE /api/contact-sources/{id}`
- **Description**: Delete a contact source
- **Path Parameters**:
  - `id` (UUID) - Contact Source ID
- **Response**: `{"message": "Contact source deleted successfully"}`
- **Status Codes**: 200 (success), 400 (invalid ID), 500 (server error)

### Error Handling
- Invalid UUID: Returns 400 with "Invalid contact ID"
- Missing required fields: Returns 400 with specific error
- Not found: Returns 404 with "Contact not found"
- Server errors: Returns 500 with "Failed to..." message

---

## Dashboard Handler

**Purpose**: Provides aggregated statistics for the dashboard view.

**Base Route**: `/api/dashboard`

### Endpoints

#### Get Statistics
- **Method**: `GET /api/dashboard/stats`
- **Description**: Get comprehensive statistics for contacts, sessions, bookmarks, browser history, and recent items
- **Response**: `DashboardStats`
- **Status Codes**: 200 (success), 500 (server error)

### Error Handling
- Server errors: Returns 500 with error message

---

## Entity Handler

**Purpose**: Generic entity management with relationships and references, supporting Logseq integration.

**Base Route**: `/api/entities`

### Endpoints

#### List Entities
- **Method**: `GET /api/entities`
- **Description**: List all entities with optional filters
- **Query Parameters**:
  - `type` (string, optional) - Entity type filter
  - `updated_since` (string, optional) - Updated since timestamp
- **Response**: `[]EntityResponse`
- **Status Codes**: 200 (success), 500 (server error)

#### Get Entity
- **Method**: `GET /api/entities/{id}`
- **Description**: Get a single entity by ID
- **Path Parameters**:
  - `id` (UUID) - Entity ID
- **Response**: `EntityResponse`
- **Status Codes**: 200 (success), 400 (invalid ID), 404 (not found)

#### Create Entity
- **Method**: `POST /api/entities`
- **Description**: Create a new entity
- **Request Body**: `{"name": "string", "type": "string", "description": "string", "properties": {}}`
- **Response**: `EntityResponse` (201)
- **Status Codes**: 201 (created), 400 (invalid input), 500 (server error)

#### Update Entity
- **Method**: `PUT /api/entities/{id}`
- **Description**: Update an entity's fields
- **Path Parameters**:
  - `id` (UUID) - Entity ID
- **Request Body**: `UpdateEntityRequest`
- **Response**: `EntityResponse`
- **Status Codes**: 200 (success), 400 (invalid input), 500 (server error)

#### Delete Entity
- **Method**: `DELETE /api/entities/{id}`
- **Description**: Soft delete an entity
- **Path Parameters**:
  - `id` (UUID) - Entity ID
- **Response**: 204 No Content
- **Status Codes**: 204 (success), 400 (invalid ID), 500 (server error)

#### Search Entities
- **Method**: `GET /api/entities/search`
- **Description**: Search entities by name with optional type filter
- **Query Parameters**:
  - `q` or `query` (string, required) - Search query
  - `type` (string, optional) - Entity type filter
- **Response**: `[]EntityResponse`
- **Status Codes**: 200 (success), 400 (missing query), 500 (server error)

#### List Deleted Entities
- **Method**: `GET /api/entities/deleted`
- **Description**: List all soft-deleted entities
- **Response**: `[]EntityResponse`
- **Status Codes**: 200 (success), 500 (server error)

### Entity Relationships

#### Get Entity Relationships
- **Method**: `GET /api/entities/{id}/relationships`
- **Description**: Get relationships for an entity with optional type filters
- **Path Parameters**:
  - `id` (UUID) - Entity ID
- **Query Parameters**:
  - `related_type` (string, optional) - Related type filter
  - `relationship_type` (string, optional) - Relationship type filter
- **Response**: `[]EntityRelationshipResponse`
- **Status Codes**: 200 (success), 400 (invalid ID), 500 (server error)

#### Create Entity Relationship
- **Method**: `POST /api/entities/relationship`
- **Description**: Create a new relationship between an entity and another resource
- **Request Body**: `{"entity_id": "UUID", "related_type": "string", "related_id": "UUID", "relationship_type": "string", "metadata": {}}`
- **Response**: `EntityRelationshipResponse` (201)
- **Status Codes**: 201 (created), 400 (invalid input), 404 (entity not found), 500 (server error)

#### Delete Entity Relationship
- **Method**: `DELETE /api/entities/relationship/{id}`
- **Description**: Delete an entity relationship by ID
- **Path Parameters**:
  - `id` (UUID) - Relationship ID
- **Response**: 204 No Content
- **Status Codes**: 204 (success), 400 (invalid ID), 500 (server error)

### Entity References

#### Get Entity References
- **Method**: `GET /api/entity-references`
- **Description**: Get references to an entity with optional source type filter
- **Query Parameters**:
  - `entity_id` (UUID, required) - Entity ID
  - `source_type` (string, optional) - Source type filter
- **Response**: `[]EntityReferenceResponse`
- **Status Codes**: 200 (success), 400 (missing/invalid entity_id), 500 (server error)

#### Parse Entity References
- **Method**: `POST /api/entity-references/parse`
- **Description**: Parse [[entity]] references from content text
- **Request Body**: `{"content": "string"}`
- **Response**: `[]ParsedReferenceResponse`
- **Status Codes**: 200 (success), 400 (invalid input), 500 (server error)

### Error Handling
- Invalid UUID: Returns 400 with "Invalid entity ID"
- Missing required fields: Returns 400 with "Name and type are required"
- Not found: Returns 404 with "Entity not found"
- Server errors: Returns 500 with "Failed to..." message

---

## Item Handler

**Purpose**: Manages items (generic content units) with tags and vector search capabilities.

**Base Route**: `/api/items`

### Endpoints

#### List Items
- **Method**: `GET /api/items`
- **Description**: Get paginated list of items
- **Query Parameters**:
  - `page` (int, optional, default: 1) - Page number
  - `limit` (int, optional, default: 10, max: 100) - Page size
- **Response**: `ItemsListResponse`
- **Status Codes**: 200 (success), 500 (server error)

#### Get Item
- **Method**: `GET /api/items/{id}`
- **Description**: Get a single item with tags
- **Path Parameters**:
  - `id` (UUID) - Item ID
- **Response**: `ItemResponse`
- **Status Codes**: 200 (success), 400 (invalid ID), 404 (not found)

#### Create Item
- **Method**: `POST /api/items`
- **Description**: Create a new item with tags
- **Request Body**: `{"title": "string", "contents": "string", "tags": ["string"]}`
- **Response**: `{"data": ItemResponse, "message": "Item created successfully"}` (201)
- **Status Codes**: 201 (created), 400 (validation error), 500 (server error)
- **Validation**: Title, contents, and at least one tag are required

#### Update Item
- **Method**: `PUT /api/items/{id}`
- **Description**: Update an item's information
- **Path Parameters**:
  - `id` (UUID) - Item ID
- **Request Body**: `{"title": "string", "contents": "string"}`
- **Response**: `ItemResponse` with message
- **Status Codes**: 200 (success), 400 (invalid input), 404 (not found)

#### Delete Item
- **Method**: `DELETE /api/items/{id}`
- **Description**: Delete an item and all related data
- **Path Parameters**:
  - `id` (UUID) - Item ID
- **Response**: `{"message": "Item deleted successfully"}`
- **Status Codes**: 200 (success), 400 (invalid ID), 500 (server error)

#### Search Items
- **Method**: `GET /api/items/search`
- **Description**: Perform vector similarity search on items
- **Query Parameters**:
  - `q` (string, required) - Search query (embedding as JSON array)
- **Response**: `{"data": []ItemListItemResponse}`
- **Status Codes**: 200 (success), 400 (validation error), 500 (server error)

### Tag Management

#### Get Item Tags
- **Method**: `GET /api/items/{id}/tags`
- **Description**: Get all tags for a specific item
- **Path Parameters**:
  - `id` (UUID) - Item ID
- **Response**: `[]string`
- **Status Codes**: 200 (success), 400 (invalid ID), 404 (not found)

#### Update Item Tags
- **Method**: `PUT /api/items/{id}/tags`
- **Description**: Update the tags for a specific item
- **Path Parameters**:
  - `id` (UUID) - Item ID
- **Request Body**: `[]string` (array of tag names)
- **Response**: `{"message": "Tags updated successfully"}`
- **Status Codes**: 200 (success), 400 (invalid input), 500 (server error)

#### Add Tag (Legacy)
- **Method**: `POST /api/items/{id}/tags`
- **Description**: Backwards-compatible endpoint for adding a single tag
- **Path Parameters**:
  - `id` (UUID) - Item ID
- **Request Body**: `{"tag": "string"}`
- **Response**: `{"message": "Tag added successfully"}`
- **Status Codes**: 200 (success), 400 (invalid input), 500 (server error)

#### Remove Tag (Legacy)
- **Method**: `DELETE /api/items/{id}/tags?tagname={name}`
- **Description**: Backwards-compatible endpoint for removing a tag
- **Path Parameters**:
  - `id` (UUID) - Item ID
- **Query Parameters**:
  - `tagname` (string, required) - Tag name
- **Response**: `{"message": "Tag removed successfully"}`
- **Status Codes**: 200 (success), 400 (invalid input), 500 (server error)

### Error Handling
- Invalid UUID: Returns 400 with "Invalid item ID"
- Validation errors: Returns 400 with specific error details
- Not found: Returns 404 with "Item not found"
- Server errors: Returns 500 with "Failed to..." message

---

## Logseq Handler

**Purpose**: Synchronizes Logseq git repository with the database, managing bidirectional sync of markdown files and entities.

**Base Route**: `/api/sync/logseq`

### Endpoints

#### Synchronize
- **Method**: `POST /api/sync/logseq`
- **Description**: Trigger full Logseq sync between git repository and database
- **Response**: `SyncStatsResponse` (pages and entities processed, created, updated, skipped, errors)
- **Status Codes**: 200 (success), 500 (server error)

#### Perform Hard Sync Check
- **Method**: `GET /api/sync/logseq/check`
- **Description**: Compare all files in the Logseq folder with all entries in the database
- **Response**: `SyncCheckResponse` (missing_in_db, missing_in_git, out_of_sync items)
- **Status Codes**: 200 (success), 500 (server error)

#### Force Update File from DB
- **Method**: `POST /api/sync/logseq/force-git`
- **Description**: Force update of a git file with data from the database for an entity
- **Request Body**: `{"entity_id": "UUID"}`
- **Response**: `{"message": "File updated successfully"}`
- **Status Codes**: 200 (success), 400 (invalid input), 500 (server error)

#### Force Update DB from File
- **Method**: `POST /api/sync/logseq/force-db`
- **Description**: Force update of database entry with data from a git file
- **Request Body**: `{"page_path": "string"}`
- **Response**: `EntityResponse`
- **Status Codes**: 200 (success), 400 (invalid input), 500 (server error)

#### Force Update DB from File by UUID (Legacy)
- **Method**: `POST /api/sync/logseq/force-db/{uuid}`
- **Description**: Backwards-compatible endpoint using entity UUID instead of page_path
- **Path Parameters**:
  - `uuid` (UUID) - Entity UUID
- **Response**: `EntityResponse`
- **Status Codes**: 200 (success), 400 (invalid UUID or missing page_path), 404 (entity not found), 500 (server error)

### Error Handling
- Invalid UUID format: Returns 400 with "Invalid UUID format"
- Missing page_path: Returns 400 with "page_path is required"
- Entity not found: Returns 404 with "Entity not found"
- Server errors: Returns 500 with error message

---

## Message Handler

**Purpose**: Manages chat messages with search and retrieval capabilities.

**Base Route**: `/api/messages`

### Endpoints

#### Get Message
- **Method**: `GET /api/messages/{id}`
- **Description**: Get a single message by ID
- **Path Parameters**:
  - `id` (UUID) - Message ID
- **Response**: `Message`
- **Status Codes**: 200 (success), 400 (invalid ID), 404 (not found), 500 (server error)

#### Get All Message Contents
- **Method**: `GET /api/messages/content`
- **Description**: Get all message contents
- **Response**: `{"data": []MessageContent}`
- **Status Codes**: 200 (success), 500 (server error)

#### Get Message Text Representations
- **Method**: `GET /api/messages/{id}/text-representations`
- **Description**: Get text representations for a message
- **Path Parameters**:
  - `id` (UUID) - Message ID
- **Response**: `[]TextRepresentation`
- **Status Codes**: 200 (success), 400 (invalid ID), 500 (server error)

#### Search Messages
- **Method**: `GET /api/messages/search`
- **Description**: Search messages by query
- **Query Parameters**:
  - `q` (string, required) - Search query
  - `page` (int, optional, default: 1) - Page number
  - `pageSize` (int, optional, default: 50) - Page size
- **Response**: `PaginatedResponse[Message]`
- **Status Codes**: 200 (success), 400 (missing query), 500 (server error)

### Room Messages

#### Get Messages by Room ID
- **Method**: `GET /api/rooms/{roomId}/messages`
- **Description**: Get messages for a specific room
- **Path Parameters**:
  - `roomId` (UUID) - Room ID
- **Query Parameters**:
  - `page` (int, optional, default: 1) - Page number
  - `pageSize` (int, optional, default: 100) - Page size
- **Response**: `PaginatedResponse[Message]`
- **Status Codes**: 200 (success), 400 (invalid room ID), 500 (server error)

### Error Handling
- Invalid UUID: Returns 400 with error
- Missing query parameter: Returns 400 with "Query parameter 'q' is required"
- Not found: Returns 404
- Server errors: Returns 500

---

## Note Handler

**Purpose**: Manages notes with tags, entity relationships, and vector search.

**Base Route**: `/api/notes`

### Endpoints

#### List Notes
- **Method**: `GET /api/notes`
- **Description**: Get paginated list of notes with optional search
- **Query Parameters**:
  - `page` (int, optional, default: 1) - Page number
  - `pageSize` (int, optional, default: 12) - Page size
  - `searchQuery` (string, optional) - Search query
- **Response**: `NotesListResponse`
- **Status Codes**: 200 (success), 500 (server error)

#### Get Note
- **Method**: `GET /api/notes/{id}`
- **Description**: Get a single note with tags and processed content
- **Path Parameters**:
  - `id` (UUID) - Note ID
- **Response**: `NoteResponse` (includes processed content and entity_id)
- **Status Codes**: 200 (success), 400 (invalid ID), 404 (not found)

#### Create Note
- **Method**: `POST /api/notes`
- **Description**: Create a new note with tags and entity relationships
- **Request Body**: `CreateNoteInput` (title required)
- **Response**: `NoteResponse` (201)
- **Status Codes**: 201 (created), 400 (invalid input), 500 (server error)

#### Update Note
- **Method**: `PUT /api/notes/{id}`
- **Description**: Update a note's information and tags
- **Path Parameters**:
  - `id` (UUID) - Note ID
- **Request Body**: `UpdateNoteInput`
- **Response**: `NoteResponse`
- **Status Codes**: 200 (success), 400 (invalid input), 500 (server error)

#### Delete Note
- **Method**: `DELETE /api/notes/{id}`
- **Description**: Delete a note and all related data
- **Path Parameters**:
  - `id` (UUID) - Note ID
- **Response**: `{"message": "Note deleted successfully"}`
- **Status Codes**: 200 (success), 400 (invalid ID), 500 (server error)

#### Search Notes
- **Method**: `GET /api/notes/search`
- **Description**: Perform vector similarity search on notes
- **Query Parameters**:
  - `q` (string, required) - Search query
  - `strategy` (string, optional, default: "qa-v2-passage") - Search strategy
- **Response**: `[]NoteListItemResponse`
- **Status Codes**: 200 (success), 400 (missing query), 500 (server error)

#### List Tags
- **Method**: `GET /api/tags`
- **Description**: Get all available tags in the system
- **Response**: `[]NoteTag`
- **Status Codes**: 200 (success), 500 (server error)

### Error Handling
- Invalid UUID: Returns 400 with "Invalid note ID"
- Missing required fields: Returns 400 with specific error
- Not found: Returns 404 with "Note not found"
- Server errors: Returns 500 with "Failed to..." message

---

## Observation Handler

**Purpose**: Stores feedback observations for Q&A interactions on bookmarks.

**Base Route**: `/api/observations`

### Endpoints

#### Store Feedback
- **Method**: `POST /api/observations/feedback`
- **Description**: Store feedback for Q&A and optionally delete the content reference
- **Request Body**: `StoreFeedbackInput`
- **Response**: `Observation` (201)
- **Status Codes**: 201 (created), 400 (invalid input), 500 (server error)

#### Get Feedback Statistics
- **Method**: `GET /api/bookmarks/{id}/feedback`
- **Description**: Get aggregated feedback statistics for a bookmark
- **Path Parameters**:
  - `id` (UUID) - Bookmark ID
- **Response**: `FeedbackStats`
- **Status Codes**: 200 (success), 400 (invalid ID), 500 (server error)

### Error Handling
- Invalid UUID: Returns 400
- Invalid input: Returns 400
- Server errors: Returns 500

---

## Room Handler

**Purpose**: Manages chat rooms with message retrieval and search.

**Base Route**: `/api/rooms`

### Endpoints

#### List Rooms
- **Method**: `GET /api/rooms`
- **Description**: Get paginated list of rooms
- **Query Parameters**:
  - `page` (int, optional, default: 1) - Page number
  - `pageSize` (int, optional, default: 10) - Page size
  - `searchText` (string, optional) - Search text
- **Response**: `PaginatedResponse[Room]`
- **Status Codes**: 200 (success), 500 (server error)

#### Get Room Details
- **Method**: `GET /api/rooms/{id}`
- **Description**: Get detailed information for a room
- **Path Parameters**:
  - `id` (UUID) - Room ID
- **Response**: `RoomDetails`
- **Status Codes**: 200 (success), 400 (invalid ID), 404 (not found), 500 (server error)

#### Get Room Messages
- **Method**: `GET /api/rooms/{id}/messages`
- **Description**: Get messages for a room
- **Path Parameters**:
  - `id` (UUID) - Room ID
- **Query Parameters**:
  - `page` (int, optional, default: 1) - Page number
  - `pageSize` (int, optional, default: 50) - Page size
  - `beforeMessageId` (UUID, optional) - Get messages before this message ID
- **Response**: `PaginatedResponse[Message]`
- **Status Codes**: 200 (success), 400 (invalid ID), 500 (server error)

#### Search Room Messages
- **Method**: `GET /api/rooms/{id}/messages/search`
- **Description**: Search messages within a room
- **Path Parameters**:
  - `id` (UUID) - Room ID
- **Query Parameters**:
  - `searchText` (string, required) - Search text
  - `page` (int, optional, default: 1) - Page number
  - `pageSize` (int, optional, default: 50) - Page size
- **Response**: `PaginatedResponse[Message]`
- **Status Codes**: 200 (success), 400 (invalid ID or missing search text), 500 (server error)

#### Set Room Name
- **Method**: `PUT /api/rooms/{id}/name`
- **Description**: Update the name of a room
- **Path Parameters**:
  - `id` (UUID) - Room ID
- **Request Body**: `{"name": "string"}`
- **Response**: `{"status": "success"}`
- **Status Codes**: 200 (success), 400 (invalid input), 500 (server error)

#### Get Sessions Count
- **Method**: `GET /api/rooms/{id}/sessions/count`
- **Description**: Get the number of sessions in a room
- **Path Parameters**:
  - `id` (UUID) - Room ID
- **Response**: `{"count": 0}`
- **Status Codes**: 200 (success), 400 (invalid ID), 500 (server error)

### Error Handling
- Invalid UUID: Returns 400 with "Invalid room ID"
- Missing search text: Returns 400 with "Search text is required"
- Not found: Returns 404 with "Room not found"
- Server errors: Returns 500 with "Failed to..." message

---

## Search Handler

**Purpose**: Provides unified and advanced LLM-powered search across all content types.

**Base Route**: `/api/search`

### Endpoints

#### Search All
- **Method**: `GET /api/search`
- **Description**: Unified search across contacts, conversations, bookmarks, browser history, and notes
- **Query Parameters**:
  - `q` or `query` (string, required) - Search query (supports both parameter names)
  - `limit` (int, optional, default: 50) - Result limit
  - `exact_match_weight` (float, optional, default: 5.0) - Exact match scoring weight
  - `similarity_weight` or `levenshteinWeight` (float, optional, default: 2.0) - Similarity scoring weight (supports both parameter names)
  - `recency_weight` (float, optional, default: 1.0) - Recency scoring weight
- **Response**: `[]UnifiedSearchResult`
- **Status Codes**: 200 (success), 400 (missing query), 500 (server error)

#### Advanced Search
- **Method**: `POST /api/search/advanced`
- **Description**: LLM-powered search that uses vector similarity on bookmarks and synthesizes an answer
- **Request Body**: `{"query": "string"}` or `{"query": {}}`
- **Response**: `AdvancedSearchResult`
- **Status Codes**: 200 (success), 400 (invalid input), 500 (server error)

### Error Handling
- Missing query: Returns 400 with "Search query is required"
- Invalid query format: Returns 400 with specific error
- Server errors: Returns 500 with error message

---

## Session Handler

**Purpose**: Manages conversation sessions with search and timeline features.

**Base Route**: `/api/sessions`

### Endpoints

#### Search Sessions
- **Method**: `GET /api/sessions/search`
- **Description**: Search sessions by query
- **Query Parameters**:
  - `q` (string, required) - Search query
  - `limit` (int, optional, default: 10) - Result limit
- **Response**: `[]SessionSearchResult`
- **Status Codes**: 200 (success), 400 (missing query), 500 (server error)

#### Get Session Messages
- **Method**: `GET /api/sessions/{id}/messages`
- **Description**: Get messages for a session
- **Path Parameters**:
  - `id` (UUID) - Session ID
- **Response**: `SessionMessagesResponse`
- **Status Codes**: 200 (success), 400 (invalid ID), 500 (server error)

### Room Sessions

#### Get Room Sessions
- **Method**: `GET /api/rooms/{id}/sessions`
- **Description**: Get all sessions for a room
- **Path Parameters**:
  - `id` (UUID) - Room ID
- **Response**: `[]Session`
- **Status Codes**: 200 (success), 400 (invalid ID), 500 (server error)

#### Get Timeline
- **Method**: `GET /api/rooms/{id}/timeline`
- **Description**: Get timeline view of sessions for a room
- **Path Parameters**:
  - `id` (UUID) - Room ID
- **Response**: `Timeline`
- **Status Codes**: 200 (success), 400 (invalid ID), 500 (server error)

### Contact Sessions

#### Search Contact Sessions
- **Method**: `GET /api/contacts/{id}/sessions/search`
- **Description**: Search sessions for a specific contact
- **Path Parameters**:
  - `id` (UUID) - Contact ID
- **Query Parameters**:
  - `q` (string, required) - Search query
- **Response**: `[]SessionSearchResult`
- **Status Codes**: 200 (success), 400 (invalid ID or missing query), 500 (server error)

### Error Handling
- Invalid UUID: Returns 400 with "invalid room ID" or "invalid contact ID"
- Missing query: Returns 400 with "query parameter 'q' is required"
- Server errors: Returns 500

---

## Social Post Handler

**Purpose**: Manages social media posts to Twitter and Bluesky with OAuth authentication.

**Base Route**: `/api/social` (primary) and `/api/microlog` (legacy)

### Endpoints

#### List Posts
- **Method**: `GET /api/social/posts`
- **Description**: Get paginated social posts with optional status filter
- **Query Parameters**:
  - `status` (string, optional) - Filter by status (pending, completed, partial, failed)
  - `page` (int, optional, default: 1) - Page number
  - `limit` (int, optional, default: 10) - Items per page
- **Response**: `PaginatedResponse[SocialPost]`
- **Status Codes**: 200 (success), 400 (invalid params), 500 (server error)
- **Legacy alias**: `GET /api/microlog`

#### Get Post
- **Method**: `GET /api/social/posts/{id}`
- **Description**: Get a single social post by its ID
- **Path Parameters**:
  - `id` (UUID) - Post ID
- **Response**: `SocialPost`
- **Status Codes**: 200 (success), 400 (invalid ID), 404 (not found), 500 (server error)
- **Legacy alias**: `GET /api/microlog/{id}`

#### Create Post
- **Method**: `POST /api/social/posts`
- **Description**: Create a new social post and publish to Twitter and Bluesky
- **Request Body**: `CreateSocialPostInput`
- **Response**: `PostResult` (201)
- **Status Codes**: 201 (created), 400 (invalid input), 500 (server error)
- **Legacy alias**: `POST /api/microlog`

#### Update Status
- **Method**: `PUT /api/social/posts/{id}/status`
- **Description**: Update the status of a social post
- **Path Parameters**:
  - `id` (UUID) - Post ID
- **Request Body**: `UpdateStatusInput`
- **Response**: 200
- **Status Codes**: 200 (success), 400 (invalid input), 500 (server error)
- **Legacy alias**: `PUT /api/microlog/{id}`

#### Delete Post
- **Method**: `DELETE /api/social/posts/{id}`
- **Description**: Delete a social post by ID
- **Path Parameters**:
  - `id` (UUID) - Post ID
- **Response**: 204 No Content
- **Status Codes**: 204 (success), 400 (invalid ID), 500 (server error)
- **Legacy alias**: `DELETE /api/microlog/{id}`

### Credentials and Authentication

#### Check Credentials
- **Method**: `GET /api/social/credentials`
- **Description**: Verify Twitter and Bluesky credentials are valid
- **Response**: `CredentialsStatus`
- **Status Codes**: 200 (success), 500 (server error)
- **Legacy alias**: `GET /api/microlog/status`

#### Update Twitter Tokens
- **Method**: `PUT /api/social/twitter/tokens`
- **Description**: Store Twitter OAuth tokens from authorization flow
- **Request Body**: `TwitterTokens`
- **Response**: 200
- **Status Codes**: 200 (success), 400 (invalid input), 500 (server error)

#### Initiate Twitter Auth
- **Method**: `POST /api/social/twitter/auth`
- **Description**: Generate Twitter OAuth authorization URL
- **Response**: `TwitterAuthURL`
- **Status Codes**: 200 (success), 500 (server error)
- **Legacy alias**: `POST /api/microlog/twitter-auth`

#### Handle Twitter Callback
- **Method**: `POST /api/social/twitter/callback`
- **Description**: Process OAuth callback and exchange code for tokens
- **Request Body**: `TwitterCallbackInput`
- **Response**: `{"status": "success"}`
- **Status Codes**: 200 (success), 400 (invalid input), 500 (server error)
- **Legacy alias**: `POST /api/microlog/twitter-callback`

### Error Handling
- Invalid UUID: Returns 400
- Invalid input: Returns 400
- Not found: Returns 404
- Server errors: Returns 500

---

## Tag Handler

**Purpose**: Manages tags for items with usage tracking and statistics.

**Base Route**: `/api/tags`

### Endpoints

#### List Tags
- **Method**: `GET /api/tags`
- **Description**: Get all tags, optionally with usage counts
- **Query Parameters**:
  - `includeUsage` (bool, optional) - Include usage counts
- **Response**: `TagsListResponse`
- **Status Codes**: 200 (success), 500 (server error)

#### Get Tag
- **Method**: `GET /api/tags/{name}`
- **Description**: Get a single tag by name
- **Path Parameters**:
  - `name` (string) - Tag name
- **Response**: `TagResponse`
- **Status Codes**: 200 (success), 404 (not found)

#### Get Items by Tag
- **Method**: `GET /api/tags/{name}/items`
- **Description**: Get all items with a specific tag
- **Path Parameters**:
  - `name` (string) - Tag name
- **Query Parameters**:
  - `page` (int, optional, default: 1) - Page number
  - `limit` (int, optional, default: 20) - Items per page
- **Response**: `ItemsListResponse`
- **Status Codes**: 200 (success), 500 (server error)

#### Add Tag
- **Method**: `POST /api/tags/{itemId}/{tagName}`
- **Description**: Add a tag to an item
- **Path Parameters**:
  - `itemId` (UUID) - Item ID
  - `tagName` (string) - Tag name
- **Response**: `{"message": "Tag added successfully"}`
- **Status Codes**: 200 (success), 400 (invalid item ID), 500 (server error)

#### Remove Tag
- **Method**: `DELETE /api/tags/{itemId}/{tagName}`
- **Description**: Remove a tag from an item
- **Path Parameters**:
  - `itemId` (UUID) - Item ID
  - `tagName` (string) - Tag name
- **Response**: `{"message": "Tag removed successfully"}`
- **Status Codes**: 200 (success), 400 (invalid item ID), 500 (server error)

#### Delete Tag
- **Method**: `DELETE /api/tags/{name}`
- **Description**: Delete a tag and all its associations
- **Path Parameters**:
  - `name` (string) - Tag name
- **Response**: `{"message": "Tag deleted successfully"}`
- **Status Codes**: 200 (success), 500 (server error)

### Error Handling
- Invalid UUID: Returns 400 with "Invalid item ID"
- Not found: Returns 404 with "Tag not found"
- Server errors: Returns 500 with "Failed to..." message

---

## Utility Handler

**Purpose**: Provides system utilities for debugging and maintenance.

**Base Routes**: Various `/api` endpoints

### Endpoints

#### Get Debug Info
- **Method**: `GET /api/debug`
- **Description**: Returns system debug information
- **Response**: `DebugInfo`
- **Status Codes**: 200 (success), 500 (server error)

#### Cleanup Stale Conversations
- **Method**: `POST /api/conversations/cleanup`
- **Description**: Remove stale sessions
- **Response**: `{"deleted_count": 0}`
- **Status Codes**: 200 (success), 500 (server error)

#### Get Messages Content
- **Method**: `POST /api/messages/content`
- **Description**: Retrieve messages by their IDs
- **Request Body**: `{"message_ids": ["UUID"]}`
- **Response**: `[]Message`
- **Status Codes**: 200 (success), 400 (invalid input), 500 (server error)

### Error Handling
- Invalid UUID: Returns 400 with "Invalid message ID: {id}"
- Invalid request body: Returns 400 with "Invalid request body"
- Server errors: Returns 500 with "Failed to..." message

---

## Common Error Handling Patterns

All handlers follow consistent error handling patterns:

1. **400 Bad Request**: Invalid input, missing required parameters, invalid UUID format
2. **404 Not Found**: Resource not found
3. **500 Internal Server Error**: Server-side errors with descriptive messages

## Authentication and Authorization

**Note**: The current implementation does not include authentication or authorization middleware. All endpoints are publicly accessible. It is recommended to implement authentication middleware before deploying to production.

## Response Format

All JSON responses use `application/json` content type. Pagination responses follow this structure:

```json
{
  "data": [...],
  "page": 1,
  "totalPages": 10,
  "limit": 10,
  "total": 100
}
```

## Request Body Format

All POST/PUT requests expect JSON bodies with `Content-Type: application/json`.

## UUID Format

All resource identifiers use UUID v4 format. Invalid UUIDs return 400 errors.

## Backwards Compatibility

Many handlers include backwards-compatible aliases for legacy endpoints to support existing clients during migration.
