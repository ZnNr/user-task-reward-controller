package errors

import "fmt"

// ErrorType - тип для обозначения категории ошибки.
type ErrorType string

// Определение различных типов ошибок.
const (
	NotFound      ErrorType = "NOT_FOUND"
	BadRequest    ErrorType = "BAD_REQUEST"
	Internal      ErrorType = "INTERNAL"
	Validation    ErrorType = "VALIDATION"
	AlreadyExists ErrorType = "ALREADY_EXISTS"
	InvalidToken  ErrorType = "INVALID_TOKEN" // Новая ошибка для недействительного токена

	ErrMsgInvalidInput = "invalid input parameters"
	ErrMsgInternal     = "internal server error"
	ErrMsgNotFound     = "resource not found"
	ErrMsgInvalidToken = "invalid token" // Сообщение для недействительного токена
)

// StatusCode - мапа с кодами статуса для каждого типа ошибки.
var StatusCode = map[ErrorType]int{
	NotFound:      404,
	BadRequest:    400,
	Internal:      500,
	Validation:    422,
	AlreadyExists: 409,
	InvalidToken:  401, // Код состояния для недействительного токена
}

// Error - структура, представляющая ошибку с дополнительной информацией.
type Error struct {
	Type    ErrorType // Тип ошибки
	Message string    // Сообщение об ошибке
	Err     error     // Вложенная ошибка, если есть
}

// Error - метод для реализации интерфейса error.
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Status - метод возвращает код статуса для ошибки.
func (e *Error) Status() int {
	if code, exists := StatusCode[e.Type]; exists {
		return code
	}
	return 500 // Если тип ошибки не определен, возвращаем код по умолчанию.
}

// NewError создает новую ошибку с заданным типом и сообщением.
func NewError(errorType ErrorType, message string, err error) *Error {
	return &Error{
		Type:    errorType,
		Message: message,
		Err:     err,
	}
}

// Удобные функции для создания ошибок конкретных типов.

func NewNotFound(message string, err error) *Error {
	return NewError(NotFound, message, err)
}

func NewBadRequest(message string, err error) *Error {
	return NewError(BadRequest, message, err)
}

func NewInternal(message string, err error) *Error {
	return NewError(Internal, message, err)
}

func NewValidation(message string, err error) *Error {
	return NewError(Validation, message, err)
}

func NewAlreadyExists(message string, err error) *Error {
	return NewError(AlreadyExists, message, err)
}

func NewInvalidArgument(message string, err error) *Error {
	return NewError(BadRequest, message, err)
}

func NewInvalidToken(message string, err error) *Error {
	return NewError(InvalidToken, message, err) // Новая функция для недействительного токена
}

// Проверки типов ошибок.
func IsErrorType(err error, errorType ErrorType) bool {
	if e, ok := err.(*Error); ok {
		return e.Type == errorType
	}
	return false
}

// Упрощенные проверки для конкретных типов ошибок.
func IsNotFound(err error) bool {
	return IsErrorType(err, NotFound)
}

func IsBadRequest(err error) bool {
	return IsErrorType(err, BadRequest)
}

func IsInvalidToken(err error) bool { // Функция для проверки недействительного токена
	return IsErrorType(err, InvalidToken)
}

// Unwrap для поддержки errors.Is и errors.As
func (e *Error) Unwrap() error {
	return e.Err
}
