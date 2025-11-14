package dto

import (
	"sync"

	"github.com/go-playground/validator/v10"
)

var once sync.Once
var validate *validator.Validate

func GetValidator() *validator.Validate {
	once.Do(func() {
		validate = validator.New()
	})

	return validate
}

type SignUpRequest struct {
	FirstName string `json:"first_name" validate:"required,min=3,max=50"`
	LastName  string `json:"last_name" validate:"required,min=3,max=50"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=6,max=64"`
}

type SignUpResponse struct {
	AccessToken string `json:"access_token"`
}

type SignInRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=64"`
}

type SignInResponse struct {
	AccessToken string `json:"access_token"`
}

type AddPreferenceCategoriesRequest struct {
	CategoryIDs []string `json:"category_ids" validate:"required,min=1,dive,uuid4"`
}

type AddPreferenceCategoriesResponse struct {
	Status string `json:"status"`
}
