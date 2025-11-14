package entity

import (
	"github.com/go-playground/validator/v10"
	"time"
)

var validate = validator.New()

type User struct {
	UserInfo

	Id        string
	Email     string `validate:"required,email"`
	Password  string `validate:"required,min=6,max=64"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (u *User) Validate() error {
	return validate.Struct(u)
}

type UserInfo struct {
	FirstName string `validate:"required,min=3,max=50"`
	LastName  string `validate:"required,min=3,max=50"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
