// handler/user.go
package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/alexuryumtsev/gophermart/internal/auth"
	"github.com/alexuryumtsev/gophermart/internal/repository"
)

type UserHandler struct {
	userRepo *repository.UserRepository
}

type UserCredentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func NewUserHandler(userRepo *repository.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
	}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var creds UserCredentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Проверяем, существует ли пользователь
	existingUser, err := h.userRepo.GetByLogin(ctx, creds.Login)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if existingUser != nil {
		http.Error(w, "Login already taken", http.StatusConflict)
		return
	}

	// Создаем пользователя
	user, err := h.userRepo.Create(ctx, creds.Login, creds.Password)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Генерируем JWT токен
	token, err := auth.GenerateToken(user.ID)
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

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var creds UserCredentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Получаем пользователя по логину
	user, err := h.userRepo.GetByLogin(ctx, creds.Login)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if user == nil || !h.userRepo.VerifyPassword(user, creds.Password) {
		http.Error(w, "Invalid login/password pair", http.StatusUnauthorized)
		return
	}

	// Генерируем JWT токен
	token, err := auth.GenerateToken(user.ID)
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
