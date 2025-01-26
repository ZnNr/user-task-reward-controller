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

// AuthService структура для работы с аутентификацией и регистрацией пользователей
type AuthService struct {
	repo     repository.AuthRepository
	logger   *zap.Logger
	SignKey  string
	TokenTTL time.Duration
}

// AuthDependencies зависимости для создания AuthService
type AuthDependencies struct {
	authRepo repository.AuthRepository
	logger   *zap.Logger
	signKey  string
	tokenTTL time.Duration
}

// NewAuthService создает новый экземпляр AuthService
func NewAuthService(deps AuthDependencies) *AuthService {
	return &AuthService{
		repo:     deps.authRepo,
		logger:   deps.logger,
		SignKey:  deps.signKey,
		TokenTTL: deps.tokenTTL,
	}
}

// Login выполняет вход пользователя и возвращает токен
func (s *AuthService) Login(ctx context.Context, login *models.SignIn) (string, error) {
	const op = "service.Auth.Login"
	logger := s.logger.With(zap.String("op", op))
	if login.Username == "" {
		logger.Error("username is required")
		return "", errors.NewBadRequest(errors.ErrorMessage[errors.BadRequest], nil)
	}
	if login.Password == "" {
		logger.Error("password is required")
		return "", errors.NewBadRequest(errors.ErrorMessage[errors.BadRequest], nil)
	}
	user, err := s.repo.GetUserByUsername(ctx, login.Username)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Error("user not found", zap.String("Username", login.Username))
			return "", errors.NewUnauthorized(errors.ErrorMessage[errors.Unauthorized], nil)
		}
		logger.Error("cannot get user", zap.String("Username", login.Username), zap.Error(err))
		return "", errors.NewInternal(errors.ErrorMessage[errors.Internal], err)
	}
	// Сравниваем хэш пароля
	if !CheckPasswordHash(login.Password, user.Password) {
		logger.Error("invalid password", zap.String("Username", login.Username))
		return "", errors.NewUnauthorized(errors.ErrorMessage[errors.Unauthorized], nil)
	}
	// Генерируем токен
	token, err := s.generateToken(*user)
	if err != nil {
		logger.Error("cannot generate token", zap.String("Username", login.Username), zap.Error(err))
		return "", errors.NewInternal(errors.ErrorMessage[errors.Internal], err)
	}

	logger.Info("User logged in successfully", zap.String("Username", login.Username))
	return token, nil
}

// Register регистрирует нового пользователя
func (s *AuthService) Register(ctx context.Context, signUp *models.CreateUser) (int64, error) {
	const op = "service.Auth.Register"
	logger := s.logger.With(zap.String("op", op))
	if signUp.Username == "" {
		logger.Error("username is required")
		return 0, errors.NewBadRequest(errors.ErrorMessage[errors.BadRequest], nil)
	}
	if signUp.Password == "" {
		logger.Error("password is required")
		return 0, errors.NewBadRequest(errors.ErrorMessage[errors.BadRequest], nil)
	}
	signUp.Password, _ = s.generatePasswordHash(signUp.Password)
	userId, err := s.repo.CreateUser(ctx, signUp)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			logger.Error("user already exists", zap.String("username", signUp.Username))
			return 0, errors.NewAlreadyExists(errors.ErrorMessage[errors.AlreadyExists], err)
		}
		logger.Error("cannot create user", zap.String("username", signUp.Username))
		return 0, errors.NewInternal(errors.ErrorMessage[errors.Internal], err)
	}

	logger.Info("User registered successfully", zap.Int64("user_id", userId))
	return userId, nil
}

// GetUser получает пользователя по имени и паролю
func (s *AuthService) GetUser(ctx context.Context, up *models.SignIn) (*models.User, error) {
	const op = "service.Auth.GetUser"
	logger := s.logger.With(zap.String("op", op))
	if up.Username == "" {
		logger.Error("username is required")
		return nil, errors.NewBadRequest(errors.ErrorMessage[errors.BadRequest], nil)
	}
	if up.Password == "" {
		logger.Error("password is required")
		return nil, errors.NewBadRequest(errors.ErrorMessage[errors.BadRequest], nil)
	}
	user, err := s.repo.GetUser(ctx, up)
	if err != nil {
		logger.Error("cannot get user", zap.String("Username", up.Username))
		if errors.IsNotFound(err) {
			return nil, errors.NewNotFound(errors.ErrorMessage[errors.NotFound], err)
		}
		return nil, errors.NewInternal(errors.ErrorMessage[errors.Internal], err)
	}

	logger.Info("User fetched successfully", zap.Int64("user_id", user.ID))
	return user, nil
}

// generateToken генерирует JWT токен для пользователя
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
		return "", errors.NewInternal(errors.ErrorMessage[errors.Internal], err)
	}
	logger.Info("Token generated successfully", zap.String("username", user.Username))
	return tokenString, nil
}

// ParseToken разбирает JWT токен и возвращает ID пользователя
func (s *AuthService) ParseToken(accessToken string) (int64, error) {
	const op = "service.Auth.ParseToken"
	logger := s.logger.With(zap.String("op", op))
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			logger.Error(fmt.Sprintf("unexpected signing method: %v", token.Header["alg"]))
			return nil, errors.NewInvalidToken(errors.ErrorMessage[errors.InvalidToken], nil)
		}
		return []byte(s.SignKey), nil
	})
	if err != nil {
		logger.Error("cannot parse token", zap.String("accessToken", accessToken))
		return 0, errors.NewInvalidToken(errors.ErrorMessage[errors.InvalidToken], err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		logger.Error("cannot parse token claims", zap.String("accessToken", accessToken))
		return 0, errors.NewInvalidToken(errors.ErrorMessage[errors.InvalidToken], nil)
	}
	userId, ok := claims["user_id"].(float64)
	if !ok {
		logger.Error("cannot get user_id from token claims", zap.String("accessToken", accessToken))
		return 0, errors.NewInvalidToken(errors.ErrorMessage[errors.InvalidToken], nil)
	}

	logger.Info("Token parsed successfully", zap.Float64("user_id", userId))
	return int64(userId), nil
}

// generatePasswordHash генерирует хэш пароля
func (s *AuthService) generatePasswordHash(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// Проверка пароля
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
