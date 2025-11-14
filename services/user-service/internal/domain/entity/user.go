package entity

import "time"

type User struct {
	UserInfo

	Id        string
	Email     string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserInfo struct {
	FirstName string
	LastName  string
	CreatedAt time.Time
	UpdatedAt time.Time
}
