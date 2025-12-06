package domain

//TODO: remove unused constants

// WE NEED
// EventUserSignUp -> EventBucketCreated -> EventEmailSent

// So we only need CompensateSignUp and CompensateBucket if notification failed
// Or CompensateSignUp if CompesateBucket failed

// THE END!

const (
	EventUserSignUp             = "saga.user.signup"
	EventBucketCreated          = "saga.bucket.created"
	EventBucketFailed           = "saga.bucket.failed"
	EventEmailSent              = "saga.email.sent"
	EventEmailFailed            = "saga.email.failed"
	EventUserCompensated        = "saga.user.compensated"
	EventUserCompensationFailed = "saga.user.compensation_failed"
)

const (
	CmdCreateBucket     = "saga.cmd.create_bucket"
	CmdCompensateBucket = "saga.cmd.compensate_bucket"
	CmdSendEmail        = "saga.cmd.send_email"
	CmdCompensateUser   = "saga.cmd.compensate_user"
)

const (
	QueueUserSignUp           = "saga.user.signup"
	QueueBucketEvents         = "saga.bucket.events"
	QueueEmailEvents          = "saga.email.events"
	QueueContentCommands      = "saga.content.commands"
	QueueUserCommands         = "saga.user.commands"
	QueueNotificationCommands = "saga.notification.commands"
	QueueUserEvents           = "saga.user.events"
	QueueUserCompensated      = "saga.user.compensated"
)

type UserSignUpPayload struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type BucketPayload struct {
	UserID     string `json:"user_id"`
	BucketName string `json:"bucket_name,omitempty"`
	Error      string `json:"error,omitempty"`
}

type EmailPayload struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Error     string `json:"error,omitempty"`
}

type UserPayload struct {
	UserID string `json:"user_id"`
}
