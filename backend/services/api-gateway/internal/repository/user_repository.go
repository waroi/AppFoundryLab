package repository

import (
	"context"

	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	List(ctx context.Context, limit int) ([]models.User, error)
}

type postgresUserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &postgresUserRepository{pool: pool}
}

func (r *postgresUserRepository) List(ctx context.Context, limit int) ([]models.User, error) {
	query := `
	SELECT id, name, email, created_at
	FROM users
	ORDER BY id
	LIMIT $1`
	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]models.User, 0, limit)
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return users, nil
}
