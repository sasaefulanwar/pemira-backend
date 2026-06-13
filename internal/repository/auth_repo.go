package repository

import (
	"context"
	"log"
	"pemira-backend/internal/models"

	"github.com/jmoiron/sqlx"
)

// AuthRepository adalah kontrak untuk akses database terkait autentikasi
type AuthRepository interface {
	IsEmailBlacklisted(ctx context.Context, email string) (bool, error)
	GetUserByEmail(email string) (*models.Pemilih, error) // <--- WAJIB TAMBAH INI
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

func (r *authRepositoryImpl) GetUserByEmail(email string) (*models.Pemilih, error) {
	var user models.Pemilih
	// Pastikan kolom ini ada di struct models.Pemilih kamu ya!
	query := "SELECT nim, nama, email_gmail_login, role FROM pemilih WHERE email_gmail_login = $1"

	// Gunakan GetContext biar konsisten dengan method lainnya
	err := r.db.Get(&user, query, email)
	if err != nil {
		log.Printf("❌ Gagal ambil user by email: %v", err)
		return nil, err
	}
	return &user, nil
}
