package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	credRepo       CredentialRepo
	jwtTokenSecret []byte
}

func NewAuthService(credRepo CredentialRepo, jwtTokenSecret []byte) *AuthService {
	return &AuthService{
		credRepo:       credRepo,
		jwtTokenSecret: jwtTokenSecret,
	}
}

type ValidationResult struct {
	Login string
	ID    int
}

var ErrUserAlreadyExists = fmt.Errorf("user already exists")

func (s *AuthService) ValidateAuthorization(_ context.Context, accessToken string) (ValidationResult, error) {
	parsed, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		return s.jwtTokenSecret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	if err != nil {
		return ValidationResult{}, fmt.Errorf("could not parse token: %w", err)
	}

	mapClaims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return ValidationResult{}, fmt.Errorf("could not parse claims")
	}

	login, ok := mapClaims["login"]
	if !ok {
		return ValidationResult{}, fmt.Errorf("could not get login from claims")
	}

	loginString, ok := login.(string)
	if !ok {
		return ValidationResult{}, fmt.Errorf("could not get login from claims, login is not a string")
	}

	userID, ok := mapClaims["user_id"]
	if !ok {
		return ValidationResult{}, fmt.Errorf("could not get login from claims")
	}

	userIDFloat, ok := userID.(float64)
	if !ok {
		return ValidationResult{}, fmt.Errorf("could not get user_id from claims, user_id is not float64")
	}

	return ValidationResult{Login: loginString, ID: int(userIDFloat)}, nil
}

type AuthResult struct {
	AccessToken string `json:"access_token"`
}

func (s *AuthService) Authorize(ctx context.Context, login, password string) (AuthResult, error) {
	user, err := s.credRepo.GetUserByLogin(ctx, login)
	if err != nil {
		return AuthResult{}, fmt.Errorf("failed to get user by login: %w", err)
	}

	inputPasswordHash := HashPassword(password)

	if user.PasswordHash != inputPasswordHash {
		return AuthResult{}, fmt.Errorf("%w: invalid password. db_hash %s, provided hash %s", ErrUserNotFound, user.PasswordHash, inputPasswordHash)
	}

	jwtEncoder := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login":   login,
		"user_id": user.ID,
		"created": time.Now(),
	})

	accessToken, err := jwtEncoder.SignedString(s.jwtTokenSecret)
	if err != nil {
		return AuthResult{}, fmt.Errorf("failed to sign access token: %w", err)
	}

	return AuthResult{AccessToken: accessToken}, nil

}

type RegisterResult struct {
	PasswordHash string
}

func (s *AuthService) Register(ctx context.Context, login, password string) (RegisterResult, error) {
	_, err := s.credRepo.GetUserByLogin(ctx, login)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		return RegisterResult{}, fmt.Errorf("failed to get user by login: %w", err)
	}

	if !errors.Is(err, ErrUserNotFound) {
		return RegisterResult{}, fmt.Errorf("%w: user with login %s already exists", ErrUserAlreadyExists, login)
	}

	return RegisterResult{PasswordHash: HashPassword(password)}, nil
}

func HashPassword(password string) string {
	sum := md5.Sum([]byte(password))
	return hex.EncodeToString(sum[:])
}
