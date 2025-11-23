package entity

type CategoryID string

type Category struct {
	ID          CategoryID `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
}
