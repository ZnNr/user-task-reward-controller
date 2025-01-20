package service

import (
	"context"
	"fmt"
	"github.com/ZnNr/user-task-reward-controller/internal/errors"
	"github.com/ZnNr/user-task-reward-controller/internal/models"
	"github.com/ZnNr/user-task-reward-controller/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type AuthService struct {
	repo     repository.AuthRepository
	logger   *zap.Logger
	SignKey  string
	TokenTTL time.Duration
	Salt     string
}

type AuthDependencies struct {
	authRepo repository.AuthRepository
	logger   *zap.Logger
	signKey  string
	tokenTTL time.Duration
	salt     string
}

func NewAuthService(deps AuthDependencies) *AuthService {
	fmt.Println(deps)
	return &AuthService{
		repo:     deps.authRepo,
		logger:   deps.logger,
		SignKey:  deps.signKey,
		TokenTTL: deps.tokenTTL,
		Salt:     deps.salt,
	}
}

func (s *AuthService) Login(ctx context.Context, login *models.SignIn) (string, error) {
	const op = "service.Auth.Login"
	logger := s.logger.With(zap.String("op", op))

	if login.Username == "" {
		logger.Error("username is required")
		return "", errors.NewBadRequest(errors.ErrMsgInvalidInput, nil)
	}
	if login.Password == "" {
		logger.Error("password is required")
		return "", errors.NewBadRequest(errors.ErrMsgInvalidInput, nil)
	}
	login.Password, _ = s.generatePasswordHash(login.Password)
	user, err := s.repo.GetUser(ctx, login)
	if err != nil {
		logger.Error("cannot get user", zap.String("Username", login.Username))
		return "", errors.NewNotFound("user not found", err)
	}

	token, err := s.generateToken(user)
	if err != nil {
		logger.Error("cannot generate token", zap.String("Username", login.Username))
		return "", errors.NewInternal("failed to generate token", err)
	}
	return token, nil
}

func (s *AuthService) Register(ctx context.Context, signUp *models.CreateUser) error {
	const op = "service.Auth.Register"
	logger := s.logger.With(zap.String("op", op))

	if signUp.Username == "" {
		logger.Error("username is required")
		return errors.NewBadRequest(errors.ErrMsgInvalidInput, nil)
	}
	if signUp.Password == "" {
		logger.Error("password is required")
		return errors.NewBadRequest(errors.ErrMsgInvalidInput, nil)
	}
	signUp.Password, _ = s.generatePasswordHash(signUp.Password)
	_, err := s.repo.CreateUser(ctx, signUp)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			logger.Error("user already exists", zap.String("username", signUp.Username))
			return errors.NewAlreadyExists("user already exists", err)
		}
		logger.Error("cannot create user", zap.String("username", signUp.Username))
		return errors.NewInternal("cannot create user", err)
	}
	return nil
}

func (s *AuthService) GetUser(ctx context.Context, up *models.SignIn) (int64, error) {
	const op = "service.Auth.GetUser"
	logger := s.logger.With(zap.String("op", op))

	if up.Username == "" {
		logger.Error("username is required")
		return 0, errors.NewBadRequest(errors.ErrMsgInvalidInput, nil)
	}
	if up.Password == "" {
		logger.Error("password is required")
		return 0, errors.NewBadRequest(errors.ErrMsgInvalidInput, nil)
	}
	up.Password, _ = s.generatePasswordHash(up.Password)
	user, err := s.repo.GetUser(ctx, up)
	if err != nil {
		logger.Error("cannot get user", zap.String(" Username", up.Username))
		return 0, errors.NewNotFound("user not found", err)
	}
	return user.ID, nil
}

func (s *AuthService) generateToken(user models.User) (string, error) {
	const op = "service.Auth.generateToken"
	logger := s.logger.With(zap.String("op", op))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"expires_at": time.Now().Add(s.TokenTTL).Unix(),
		"issued_at":  time.Now().Unix(),
		"user_id":    user.ID,
	})
	tokenString, err := token.SignedString([]byte(s.SignKey))
	if err != nil {
		logger.Error("cannot sign token", zap.String("username", user.Username))
		return "", errors.NewInternal("cannot sign token", err)
	}
	return tokenString, nil
}

func (s *AuthService) ParseToken(accessToken string) (int64, error) {
	const op = "service.Auth.ParseToken"
	logger := s.logger.With(zap.String("op", op))

	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			logger.Error(fmt.Sprintf("unexpected signing method: %v", token.Header["alg"]))
			return nil, errors.NewInvalidToken("unexpected signing method", nil)
		}
		return []byte(s.SignKey), nil
	})
	if err != nil {
		logger.Error("cannot parse token", zap.String("accessToken", accessToken))
		return 0, errors.NewInvalidToken("cannot parse token", err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		logger.Error("cannot parse token claims", zap.String("accessToken", accessToken))
		return 0, errors.NewInvalidToken("cannot parse token claims", nil)
	}
	userId, ok := claims["user_id"].(float64)
	if !ok {
		logger.Error("cannot get user_id from token claims", zap.String("accessToken", accessToken))
		return 0, errors.NewInvalidToken("cannot get user_id from token claims", nil)
	}
	return int64(userId), nil
}

func (s *AuthService) generatePasswordHash(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}
