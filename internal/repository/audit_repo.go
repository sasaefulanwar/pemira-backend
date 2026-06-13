package repository

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
)

type AuditRepository interface {
	LogActivity(ctx context.Context, email string, aksi string, rincian string, ipAddress string)
}

type auditRepositoryImpl struct {
	db *sqlx.DB
}

func NewAuditRepository(db *sqlx.DB) AuditRepository {
	return &auditRepositoryImpl{db: db}
}

// LogActivity mencatat jejak digital ke database.
// Kita tidak return error agar jika log gagal, proses utama (seperti voting) tidak ikut batil.
func (r *auditRepositoryImpl) LogActivity(ctx context.Context, email string, aksi string, rincian string, ipAddress string) {
	// Gabungin semuanya jadi satu string panjang karena tabel DB lu cuma punya kolom 'aksi'
	teksAksi := aksi + " | " + rincian + " | IP: " + ipAddress

	// Sesuaikan nama kolom dengan yang ada di skema lu (admin_username dan aksi)
	query := `INSERT INTO audit_logs (admin_username, aksi) VALUES ($1, $2)`

	_, err := r.db.ExecContext(ctx, query, email, teksAksi)
	if err != nil {
		log.Println("❌ Gagal menulis audit log ke DB:", err)
	}
}
