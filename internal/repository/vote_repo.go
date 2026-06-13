package repository

import (
	"context"
	"errors"
	"log"

	"pemira-backend/internal/models"

	"github.com/jmoiron/sqlx"
)

// Tambahkan parameter nim asli ke dalam interface
type VoteRepository interface {
	CastVote(ctx context.Context, nim string, hashedNIM string, electionID int, idPaslon int, ipAddress string, userAgent string) error
}

type voteRepositoryImpl struct {
	db *sqlx.DB
}

func NewVoteRepository(db *sqlx.DB) VoteRepository {
	return &voteRepositoryImpl{db: db}
}

// CastVote sekarang memiliki validasi ketat sesuai FR-04
func (r *voteRepositoryImpl) CastVote(ctx context.Context, nim string, hashedNIM string, electionID int, idPaslon int, ipAddress string, userAgent string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			log.Println("⚠️ Transaksi gagal, melakukan Rollback...")
			tx.Rollback()
		}
	}()

	// 1. Cek Status Pemilu (Harus 'open')
	var electionStatus string
	err = tx.GetContext(ctx, &electionStatus, "SELECT status FROM elections WHERE id = $1", electionID)
	if err != nil {
		return errors.New("data pemilu tidak ditemukan")
	}
	if electionStatus != "open" {
		return errors.New("pemilu sedang tidak aktif atau sudah ditutup")
	}

	// 2. Cek Status Pemilih + ROW LOCKING (FOR UPDATE)
	// Ini bakal ngunci baris data pemilih ini sampai transaksi selesai
	var pemilih models.Pemilih
	err = tx.GetContext(ctx, &pemilih, "SELECT status_memilih, is_suspended FROM pemilih WHERE nim = $1 FOR UPDATE", nim)
	if err != nil {
		return errors.New("data pemilih tidak terdaftar")
	}
	if pemilih.IsSuspended {
		return errors.New("akun anda ditangguhkan karena sedang dalam masa sengketa")
	}
	if pemilih.StatusMemilih {
		return errors.New("anda sudah memberikan suara, tidak bisa memilih lebih dari sekali")
	}

	// 3. Insert ke tabel kertas_suara
	queryInsertSuara := `INSERT INTO kertas_suara (election_id, hashed_nim, id_paslon) VALUES ($1, $2, $3)`
	_, err = tx.ExecContext(ctx, queryInsertSuara, electionID, hashedNIM, idPaslon)
	if err != nil {
		return errors.New("gagal menyimpan suara: " + err.Error())
	}

	// 4. Update status_memilih menjadi TRUE
	queryUpdateStatus := `UPDATE pemilih SET status_memilih = TRUE WHERE nim = $1`
	_, err = tx.ExecContext(ctx, queryUpdateStatus, nim)
	if err != nil {
		return errors.New("gagal memperbarui status pemilih")
	}

	// 5. Catat ke voter_events
	queryAudit := `INSERT INTO voter_events (hashed_nim, event_type, ip_address, user_agent) VALUES ($1, $2, $3, $4)`
	_, err = tx.ExecContext(ctx, queryAudit, hashedNIM, "VOTE_CAST", ipAddress, userAgent)
	if err != nil {
		return errors.New("gagal mencatat audit log")
	}

	// 6. Commit transaksi
	err = tx.Commit()
	if err != nil {
		return errors.New("gagal mengunci transaksi: " + err.Error())
	}

	return nil
}
