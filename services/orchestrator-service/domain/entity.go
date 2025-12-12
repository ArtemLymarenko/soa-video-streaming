package domain

const (
	EventUserSignUp             = "event.user.signup"
	EventBucketCreated          = "event.content.bucket_created"
	EventBucketFailed           = "event.content.bucket_failed"
	EventEmailSent              = "event.notification.email_sent"
	EventEmailFailed            = "event.notification.email_failed"
	EventUserCompensated        = "event.user.compensated"
	EventUserCompensationFailed = "event.user.compensation_failed"
)

const (
	CmdCreateBucket     = "cmd.content.create_bucket"
	CmdCompensateBucket = "cmd.content.compensate_bucket"
	CmdSendEmail        = "cmd.notification.send_email"
	CmdCompensateUser   = "cmd.user.compensate_user"
)

const (
	QueueUserSignUp         = "queue.user.signup"
	QueueSagaErrors         = "queue.saga.errors"
	QueueContentEvents      = "queue.content.events"
	QueueNotificationEvents = "queue.notification.events"
	QueueUserEvents         = "queue.user.events"

	QueueContentCommands      = "queue.content.commands"
	QueueUserCommands         = "queue.user.commands"
	QueueNotificationCommands = "queue.notification.commands"
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
