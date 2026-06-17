package http

import (
	"errors"
	"fmt"
	"strings"

	"github.com/SergeiGD/testify-profile/internal/domain"
	"github.com/go-playground/validator/v10"
)

func mapRegisterError(err error) string {
	switch {
	case errors.Is(err, domain.ErrEmailAlreadyConfirmed):
		return "Данный email уже зарегистрирован"
	case errors.Is(err, domain.ErrResendTooFrequent):
		return "Письмо с подтверждением уже было отправлено, повторная отправка возможна раз в 1 минуту"
	default:
		return "Произошла ошибка при регистрации"
	}
}

func mapValidationError(err error) string {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return "Некорректные данные запроса"
	}

	msgs := make([]string, 0, len(ve))
	for _, fe := range ve {
		msgs = append(msgs, fieldValidationMessage(fe))
	}
	return strings.Join(msgs, "; ")
}

func fieldValidationMessage(fe validator.FieldError) string {
	switch fe.Field() {
	case "Email":
		return "Некорректный email"
	case "Password":
		if fe.Tag() == "min" {
			return fmt.Sprintf("Пароль должен содержать не менее %s символов", fe.Param())
		}
	case "Username":
		switch fe.Tag() {
		case "min":
			return fmt.Sprintf("Имя пользователя должно содержать не менее %s символов", fe.Param())
		case "max":
			return fmt.Sprintf("Имя пользователя должно содержать не более %s символов", fe.Param())
		}
	case "BirthDate":
		return "Дата рождения обязательна"
	}
	return fmt.Sprintf("Поле %s не прошло валидацию", fe.Field())
}

func mapConfirmError(err error) string {
	switch {
	case errors.Is(err, domain.ErrInvalidOrExpiredToken):
		return "Недействительная или истёкшая ссылка подтверждения"
	default:
		return "Произошла ошибка при подтверждении регистрации"
	}
}

// mapAuthError returns a user-facing message and an HTTP status code (400 or 401).
func mapAuthError(err error) (string, int) {
	switch {
	case errors.Is(err, domain.ErrInvalidCredentials):
		return "Неверный email или пароль", 401
	case errors.Is(err, domain.ErrAccountNotConfirmed):
		return "Аккаунт не подтверждён", 401
	case errors.Is(err, domain.ErrInvalidRefreshToken):
		return "Недействительный или истёкший refresh токен", 401
	default:
		return "Произошла ошибка при авторизации", 400
	}
}
