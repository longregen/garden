package repository

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// convertTimeToPgTimestamp converts time.Time to pgtype.Timestamp
func convertTimeToPgTimestamp(t time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{
		Time:  t,
		Valid: true,
	}
}

// convertPgTimestampToTime converts pgtype.Timestamp to time.Time
func convertPgTimestampToTime(ts pgtype.Timestamp) time.Time {
	if ts.Valid {
		return ts.Time
	}
	return time.Time{}
}

// convertPgTimestampToTimePtr converts pgtype.Timestamp to *time.Time
func convertPgTimestampToTimePtr(ts pgtype.Timestamp) *time.Time {
	if ts.Valid {
		return &ts.Time
	}
	return nil
}

// convertStringToPtr converts string to *string
func convertStringToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// convertInterfaceToStringSlice converts interface{} (array aggregation) to []string
func convertInterfaceToStringSlice(i interface{}) []string {
	if i == nil {
		return []string{}
	}

	tags := []string{}
	if arr, ok := i.([]interface{}); ok {
		for _, item := range arr {
			if str, ok := item.(string); ok && str != "" {
				tags = append(tags, str)
			}
		}
	}
	return tags
}

// convertUUIDToPgUUID converts uuid.UUID to pgtype.UUID
func convertUUIDToPgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: id,
		Valid: true,
	}
}

// convertPgUUIDToUUIDPtr converts pgtype.UUID to *uuid.UUID
func convertPgUUIDToUUIDPtr(id pgtype.UUID) *uuid.UUID {
	if id.Valid {
		uid := uuid.UUID(id.Bytes)
		return &uid
	}
	return nil
}

// convertMapToJSON converts map[string]interface{} to json.RawMessage
func convertMapToJSON(m map[string]interface{}) json.RawMessage {
	if m == nil {
		return json.RawMessage("{}")
	}
	b, err := json.Marshal(m)
	if err != nil {
		return json.RawMessage("{}")
	}
	return json.RawMessage(b)
}

// convertInterfaceToTimePtr converts interface{} (which may be pgtype.Timestamp) to *time.Time
func convertInterfaceToTimePtr(i interface{}) *time.Time {
	if i == nil {
		return nil
	}
	if ts, ok := i.(time.Time); ok {
		return &ts
	}
	return nil
}
