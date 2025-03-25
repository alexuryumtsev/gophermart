package repository

import (
	"context"
	"errors"

	"github.com/alexuryumtsev/gophermart/internal/model"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

// Обновление структуры UserRepository
type UserRepositoryImpl struct {
	db PgxPool
}

// Переименовываем конструктор
func NewUserRepository(db PgxPool) UserRepository {
	return &UserRepositoryImpl{db: db}
}

func (r *UserRepositoryImpl) Create(ctx context.Context, login, password string) (*model.User, error) {
	// Хешируем пароль
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{}
	query := `INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id, login, password_hash, created_at`

	err = r.db.QueryRow(ctx, query, login, string(passwordHash)).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepositoryImpl) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, login, password_hash, created_at FROM users WHERE login = $1`

	err := r.db.QueryRow(ctx, query, login).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

func (r *UserRepositoryImpl) VerifyPassword(user *model.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return err == nil
}
