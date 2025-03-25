package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/alexuryumtsev/gophermart/internal/service"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

type UserCredentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Register обрабатывает запрос на регистрацию пользователя
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var creds UserCredentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Используем сервис для регистрации пользователя
	user, err := h.userService.Register(ctx, creds.Login, creds.Password)
	if err != nil {
		if err.Error() == "login already taken" {
			http.Error(w, "Login already taken", http.StatusConflict)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Генерируем JWT токен
	token, err := h.userService.GenerateToken(user.ID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Устанавливаем cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	})

	w.WriteHeader(http.StatusOK)
}

// Login обрабатывает запрос на аутентификацию пользователя
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var creds UserCredentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Используем сервис для аутентификации пользователя
	user, err := h.userService.Authenticate(ctx, creds.Login, creds.Password)
	if err != nil {
		http.Error(w, "Invalid login/password pair", http.StatusUnauthorized)
		return
	}

	// Генерируем JWT токен
	token, err := h.userService.GenerateToken(user.ID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Устанавливаем cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	})

	w.WriteHeader(http.StatusOK)
}
