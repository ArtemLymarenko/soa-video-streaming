package entity

import "time"

type MediaContentType string

const (
	MediaContentTypeMovie  MediaContentType = "movie"
	MediaContentTypeSeries MediaContentType = "series"
)

type MediaContent struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Type        MediaContentType `json:"type"`
	Duration    int              `json:"duration"`
	Categories  []Category       `json:"categories"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}
