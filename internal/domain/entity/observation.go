package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Observation represents a generic data collection entry
type Observation struct {
	ObservationID uuid.UUID       `json:"observation_id"`
	Data          json.RawMessage `json:"data"`
	Type          *string         `json:"type,omitempty"`
	Source        *string         `json:"source,omitempty"`
	Tags          *string         `json:"tags,omitempty"`
	Parent        *uuid.UUID      `json:"parent,omitempty"`
	Ref           *uuid.UUID      `json:"ref,omitempty"`
	CreationDate  int64           `json:"creation_date"`
}

// FeedbackType represents the type of feedback
type FeedbackType string

const (
	FeedbackUpvote   FeedbackType = "upvote"
	FeedbackDownvote FeedbackType = "downvote"
	FeedbackTrash    FeedbackType = "trash"
)

// FeedbackData represents the data structure for Q&A feedback
type FeedbackData struct {
	Question     string       `json:"question"`
	Answer       string       `json:"answer"`
	BookmarkID   string       `json:"bookmarkId"`
	UserQuestion string       `json:"userQuestion"`
	Similarity   float64      `json:"similarity"`
	FeedbackType FeedbackType `json:"feedbackType"`
	Timestamp    time.Time    `json:"timestamp"`
}

// StoreFeedbackInput represents input for storing feedback
type StoreFeedbackInput struct {
	Question      string       `json:"question"`
	Answer        string       `json:"answer"`
	BookmarkID    uuid.UUID    `json:"bookmark_id"`
	ReferenceID   *uuid.UUID   `json:"reference_id,omitempty"`
	UserQuestion  string       `json:"user_question"`
	Similarity    float64      `json:"similarity"`
	FeedbackType  FeedbackType `json:"feedback_type"`
	DeleteRef     bool         `json:"delete_ref"`
}

// FeedbackStats represents aggregated feedback statistics
type FeedbackStats struct {
	Upvotes   int64 `json:"upvotes"`
	Downvotes int64 `json:"downvotes"`
	Trash     int64 `json:"trash"`
}
