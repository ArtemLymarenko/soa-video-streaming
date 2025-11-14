package entity

import "time"

type MediaContentType string

const (
	MediaContentTypeMovie  MediaContentType = "movie"
	MediaContentTypeSeries MediaContentType = "series"
)

type MediaContent struct {
	ID          string
	Name        string
	Description string
	Type        MediaContentType
	Duration    int
	Categories  []CategoryID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
