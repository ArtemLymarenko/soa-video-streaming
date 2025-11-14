package service

import (
	"context"
	"errors"
	"soa-video-streaming/services/user-service/internal/config"
	"soa-video-streaming/services/user-service/internal/domain/entity"
	"soa-video-streaming/services/user-service/internal/repository/postgres"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	usersRepo    *postgres.UsersRepository
	jwtSecretKey string
	ttl          time.Duration
}

func NewAuthService(usersRepo *postgres.UsersRepository, cfg *config.AppConfig) *AuthService {
	return &AuthService{
		usersRepo:    usersRepo,
		jwtSecretKey: cfg.Auth.JwtSecretKey,
		ttl:          cfg.Auth.JwtTTL,
	}
}

type AuthResult struct {
	AccessToken string
}

func (a *AuthService) SignUp(ctx context.Context, user entity.User) (AuthResult, error) {
	if err := user.Validate(); err != nil {
		return AuthResult{}, err
	}

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
