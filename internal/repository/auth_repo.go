package repository

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
)

// AuthRepository adalah kontrak untuk akses database terkait autentikasi
type AuthRepository interface {
	IsEmailBlacklisted(ctx context.Context, email string) (bool, error)
}

type authRepositoryImpl struct {
	db *sqlx.DB
}

func NewAuthRepository(db *sqlx.DB) AuthRepository {
	return &authRepositoryImpl{db: db}
}

// IsEmailBlacklisted mengecek apakah email penyusup ada di daftar hitam
func (r *authRepositoryImpl) IsEmailBlacklisted(ctx context.Context, email string) (bool, error) {
	var count int
	query := `SELECT count(*) FROM email_blacklist WHERE email = $1`

	err := r.db.GetContext(ctx, &count, query, email)
	if err != nil {
		log.Println("❌ Gagal mengecek blacklist:", err)
		return false, err
	}

	return count > 0, nil
}
