package service

import (
	"context"
	"encoding/json"
	"errors"
	"soa-video-streaming/pkg/rabbitmq"
	"soa-video-streaming/services/notification-service/pkg/notifications"
	"soa-video-streaming/services/user-service/internal/config"
	"soa-video-streaming/services/user-service/internal/domain/entity"
	"soa-video-streaming/services/user-service/internal/repository/postgres"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	usersRepo    *postgres.UsersRepository
	jwtSecretKey string
	ttl          time.Duration
	publisher    *rabbitmq.Publisher
}

func NewAuthService(usersRepo *postgres.UsersRepository, cfg *config.AppConfig, publisher *rabbitmq.Publisher) *AuthService {
	return &AuthService{
		usersRepo:    usersRepo,
		jwtSecretKey: cfg.Auth.JwtSecretKey,
		ttl:          cfg.Auth.JwtTTL,
		publisher:    publisher,
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

	user.Password = string(hashed)

	err = a.usersRepo.Save(ctx, user)
	if err != nil {
		return AuthResult{}, err
	}

	token, err := a.generateAccessToken(user)
	if err != nil {
		return AuthResult{}, err
	}

	go a.SendUserSignUpEvent(&user)

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

func (a *AuthService) SendUserSignUpEvent(user *entity.User) {
	reqCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	event := notifications.EventSignUp{
		UserID:    user.Id,
		Email:     user.Email,
		Message:   "New user registered successfully",
		CreatedAt: time.Now(),
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		logrus.WithError(err).Error("Failed to marshal signup event")
		return
	}

	if err := a.publisher.PublishJSON(reqCtx, eventJSON); err != nil {
		logrus.WithError(err).Error("Failed to publish signup event")
		return
	}

	logrus.WithFields(logrus.Fields{
		"user_id": user.Id,
		"email":   event.Email,
	}).Info("Signup event published successfully")
}
