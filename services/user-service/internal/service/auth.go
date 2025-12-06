package service

import (
	"context"
	"encoding/json"
	"errors"
	"soa-video-streaming/pkg/rabbitmq"
	"time"

	"soa-video-streaming/services/notification-service/pkg/notifications"
	"soa-video-streaming/services/user-service/internal/config"
	"soa-video-streaming/services/user-service/internal/domain/entity"
	"soa-video-streaming/services/user-service/internal/repository/postgres"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/oagudo/outbox"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	usersRepo    *postgres.UsersRepository
	userInfoRepo *postgres.UserInfoRepository
	outboxRepo   *postgres.OutboxRepository
	tm           *postgres.TransactionManager
	jwtSecretKey string
	ttl          time.Duration
	client       *rabbitmq.Client
}

func NewAuthService(
	usersRepo *postgres.UsersRepository,
	userInfoRepo *postgres.UserInfoRepository,
	outboxRepo *postgres.OutboxRepository,
	tm *postgres.TransactionManager,
	cfg *config.AppConfig,
	client *rabbitmq.Client,
) *AuthService {
	return &AuthService{
		usersRepo:    usersRepo,
		userInfoRepo: userInfoRepo,
		outboxRepo:   outboxRepo,
		tm:           tm,
		jwtSecretKey: cfg.Auth.JwtSecretKey,
		ttl:          cfg.Auth.JwtTTL,
		client:       client,
	}
}

type AuthResult struct {
	AccessToken string
}

func (a *AuthService) SignUp(ctx context.Context, user entity.User) (AuthResult, error) {
	actualUser, err := a.usersRepo.FindByEmail(ctx, user.Email)
	if err != nil {
		return AuthResult{}, err
	}

	if actualUser.Email != "" {
		return AuthResult{}, errors.New("user already exists")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return AuthResult{}, err
	}

	user.Id = uuid.NewString()
	user.Password = string(hashed)

	err = a.tm.RunInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if txErr := a.usersRepo.WithTx(tx).Save(ctx, user); txErr != nil {
			return txErr
		}

		if txErr := a.userInfoRepo.WithTx(tx).Save(ctx, user.Id, user.UserInfo); txErr != nil {
			return txErr
		}

		event := notifications.EventSignUp{
			UserID:    user.Id,
			Email:     user.Email,
			Message:   "New user registered successfully",
			CreatedAt: time.Now(),
		}

		eventJSON, err := json.Marshal(event)
		if err != nil {
			return err
		}

		msg := outbox.NewMessage(eventJSON,
			outbox.WithID(uuid.New()),
			outbox.WithCreatedAt(time.Now()),
			outbox.WithMetadata([]byte(notifications.QueueSignUpEvent)),
		)

		return a.outboxRepo.WithTx(tx).Save(ctx, msg)
	})

	if err != nil {
		return AuthResult{}, err
	}

	token, err := a.generateAccessToken(user)
	if err != nil {
		return AuthResult{}, err
	}

	return AuthResult{
		AccessToken: token,
	}, nil
}

func (a *AuthService) SignIn(ctx context.Context, email, password string) (AuthResult, error) {
	user, err := a.usersRepo.FindByEmail(ctx, email)
	if err != nil {
		return AuthResult{}, err
	}

	if user.Email == "" {
		return AuthResult{}, errors.New("invalid credentials")
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return AuthResult{}, errors.New("invalid credentials")
	}

	token, err := a.generateAccessToken(user)
	if err != nil {
		return AuthResult{}, err
	}

	return AuthResult{
		AccessToken: token,
	}, nil
}

func (a *AuthService) generateAccessToken(user entity.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.Id,
		"email":   user.Email,
		"exp":     time.Now().Add(a.ttl).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.jwtSecretKey))
}
