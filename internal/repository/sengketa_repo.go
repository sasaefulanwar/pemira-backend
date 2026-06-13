package repository

import (
	"context"
	"errors"
	"log"

	"github.com/jmoiron/sqlx"
)

type SengketaRepository interface {
	CreateSengketa(ctx context.Context, nimSengketa string, emailPelapor string, pathFotoKTM string) error
}

type sengketaRepositoryImpl struct {
	db *sqlx.DB
}

func NewSengketaRepository(db *sqlx.DB) SengketaRepository {
	return &sengketaRepositoryImpl{db: db}
}

// CreateSengketa memasukkan laporan pengaduan baru ke database
func (r *sengketaRepositoryImpl) CreateSengketa(ctx context.Context, nimSengketa string, emailPelapor string, pathFotoKTM string) error {
	// Mulai transaksi DB
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	// Pastikan rollback kalau ada panic atau error di tengah jalan
	defer tx.Rollback()

	// 1. Masukkan data ke tabel sengketa
	queryInsert := `INSERT INTO sengketa_nim (nim_sengketa, email_pelapor, path_foto_ktm, status_proses) VALUES ($1, $2, $3, 'pending')`
	_, err = tx.ExecContext(ctx, queryInsert, nimSengketa, emailPelapor, pathFotoKTM)
	if err != nil {
		log.Println("❌ Gagal insert sengketa:", err)
		return errors.New("gagal memproses laporan sengketa")
	}

	// 2. Suspend akun pemilih yang disengketakan! (BARU)
	queryUpdate := `UPDATE pemilih SET is_suspended = TRUE WHERE nim = $1`
	_, err = tx.ExecContext(ctx, queryUpdate, nimSengketa)
	if err != nil {
		log.Println("❌ Gagal suspend akun pemilih:", err)
		return errors.New("gagal membekukan akun yang disengketakan")
	}

	// Kalau dua-duanya mulus, commit!
	return tx.Commit()
}
