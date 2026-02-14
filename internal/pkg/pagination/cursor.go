package pagination

import (
	"encoding/base64"
	"fmt"
)

type CursorParams struct {
	Limit          int
	StartingAfter  string
	EndingBefore   string
}

type CursorMeta struct {
	HasMore    bool   `json:"has_more"`
	NextCursor string `json:"next_cursor,omitempty"`
}

func EncodeCursor(id string) string {
	return base64.URLEncoding.EncodeToString([]byte(id))
}

func DecodeCursor(cursor string) (string, error) {
	decoded, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return "", fmt.Errorf("invalid cursor: %w", err)
	}
	return string(decoded), nil
}

func ValidateLimit(limit int) int {
	if limit <= 0 {
		return 20
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func NewCursorMeta(hasMore bool, lastID string) CursorMeta {
	meta := CursorMeta{
		HasMore: hasMore,
	}
	if hasMore && lastID != "" {
		meta.NextCursor = EncodeCursor(lastID)
	}
	return meta
}
