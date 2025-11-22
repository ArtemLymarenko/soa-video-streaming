package entity

type CategoryID string

type Category struct {
	ID          CategoryID
	Name        string
	Description string
}
