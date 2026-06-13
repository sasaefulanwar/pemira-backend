package repository

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"pemira-backend/internal/models"

	"github.com/jmoiron/sqlx"
)

type PemilihRepository interface {
	FindMahasiswaByNIM(ctx context.Context, nim string) (*models.Mahasiswa, error)
	FindPemilihByNIM(ctx context.Context, nim string) (*models.Pemilih, error)
	FindPemilihByEmail(ctx context.Context, email string) (*models.Pemilih, error)
	CreatePemilih(ctx context.Context, nim string, nama string, email string) error
}

type pemilihRepositoryImpl struct {
	db *sqlx.DB
}

func NewPemilihRepository(db *sqlx.DB) PemilihRepository {
	return &pemilihRepositoryImpl{db: db}
}

// FindMahasiswaByNIM mengecek apakah NIM terdaftar di DPT Master (Tabel mahasiswa)
func (r *pemilihRepositoryImpl) FindMahasiswaByNIM(ctx context.Context, nim string) (*models.Mahasiswa, error) {
	var m models.Mahasiswa
	query := `SELECT nim, nama, angkatan, created_at FROM mahasiswa WHERE nim = $1`

	err := r.db.GetContext(ctx, &m, query, nim)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Data tidak ditemukan
		}
		log.Println("❌ Gagal query mahasiswa:", err)
		return nil, err
	}
	return &m, nil
}

// FindPemilihByNIM mengecek apakah NIM sudah pernah melakukan binding (Tabel pemilih)
func (r *pemilihRepositoryImpl) FindPemilihByNIM(ctx context.Context, nim string) (*models.Pemilih, error) {
	var p models.Pemilih
	query := `SELECT nim, nama, email_gmail_login, role, status_memilih, is_suspended, created_at FROM pemilih WHERE nim = $1`

	err := r.db.GetContext(ctx, &p, query, nim)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Belum pernah binding
		}
		log.Println("❌ Gagal query pemilih:", err)
		return nil, err
	}
	return &p, nil
}

// FindPemilihByEmail mengecek apakah email ini sudah pernah mengikat NIM lain
func (r *pemilihRepositoryImpl) FindPemilihByEmail(ctx context.Context, email string) (*models.Pemilih, error) {
	var p models.Pemilih
	query := `SELECT nim, nama, email_gmail_login FROM pemilih WHERE email_gmail_login = $1`

	err := r.db.GetContext(ctx, &p, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Email masih bersih, belum pernah binding
		}
		log.Println("❌ Gagal query pemilih by email:", err)
		return nil, err
	}
	return &p, nil
}

// CreatePemilih mengunci data NIM dengan Email Gmail yang aktif
func (r *pemilihRepositoryImpl) CreatePemilih(ctx context.Context, nim string, nama string, email string) error {
	query := `INSERT INTO pemilih (nim, nama, email_gmail_login, role, status_memilih, is_suspended) 
			  VALUES ($1, $2, $3, 'user', FALSE, FALSE)`

	_, err := r.db.ExecContext(ctx, query, nim, nama, email)
	if err != nil {
		log.Println("❌ Gagal menginsert data pemilih:", err)
		return errors.New("gagal mengaitkan akun, silakan coba lagi")
	}
	return nil
}
