package notifications

import (
	"time"
)

const SignUpEventQueueName = "user.signup"

type EventSignUp struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}
