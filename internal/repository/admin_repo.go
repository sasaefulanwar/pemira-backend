package repository

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

// DTO (Data Transfer Object) untuk output sengketa
type SengketaResponse struct {
	ID           int       `json:"id" db:"id"`
	NIMSengketa  string    `json:"nim_sengketa" db:"nim_sengketa"`
	EmailPelapor string    `json:"email_pelapor" db:"email_pelapor"`
	PathFotoKTM  string    `json:"path_foto_ktm" db:"path_foto_ktm"`
	StatusProses string    `json:"status_proses" db:"status_proses"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type AuditLogResponse struct {
	ID            int       `json:"id" db:"id"`
	AdminUsername string    `json:"admin_username" db:"admin_username"`
	Aksi          string    `json:"aksi" db:"aksi"`
	Timestamp     time.Time `json:"timestamp" db:"timestamp"`
}

type DashboardAuditResponse struct {
	TotalPemilih int                `json:"total_pemilih"`
	TotalSuara   int                `json:"total_suara"`
	RecentLogs   []AuditLogResponse `json:"recent_logs"`
}

type RecalculateResult struct {
	IDPaslon         int    `json:"id_paslon" db:"id_paslon"`
	ChairmanName     string `json:"chairman_name" db:"chairman_name"`
	ViceChairmanName string `json:"vice_chairman_name" db:"vice_chairman_name"`
	TotalSuara       int    `json:"total_suara" db:"total_suara"`
}

type AdminRepository interface {
	GetAllPendingDisputes(ctx context.Context) ([]SengketaResponse, error)
	UpdateElectionStatus(ctx context.Context, electionID int, status string) error
	ResolveDispute(ctx context.Context, disputeID int, action string, nimHash string) error
	GetDashboardAudit(ctx context.Context) (DashboardAuditResponse, error)
	RecalculateVotes(ctx context.Context) ([]RecalculateResult, error)
}

type adminRepositoryImpl struct {
	db *sqlx.DB
}

func NewAdminRepository(db *sqlx.DB) AdminRepository {
	return &adminRepositoryImpl{db: db}
}

func (r *adminRepositoryImpl) GetAllPendingDisputes(ctx context.Context) ([]SengketaResponse, error) {
	var disputes []SengketaResponse

	// Ambil semua sengketa yang statusnya masih 'pending'
	query := `
		SELECT id, nim_sengketa, email_pelapor, path_foto_ktm, status_proses, created_at 
		FROM sengketa_nim 
		WHERE status_proses = 'pending' 
		ORDER BY created_at ASC
	`

	err := r.db.SelectContext(ctx, &disputes, query)
	if err != nil {
		log.Println("❌ Gagal mengambil daftar sengketa:", err)
		return nil, err
	}

	// Jika kosong, pastikan return array kosong (bukan null) biar React gak error
	if len(disputes) == 0 {
		return []SengketaResponse{}, nil
	}

	return disputes, nil
}

func (r *adminRepositoryImpl) ResolveDispute(ctx context.Context, disputeID int, action string, nimHash string) error {
	// Mulai Transaksi Pembantaian
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // Kalau ada yang gagal, undo semua!

	// 1. Ambil data sengketa (Cari NIM yang disengketakan)
	var nimSengketa string
	err = tx.GetContext(ctx, &nimSengketa, `SELECT nim_sengketa FROM sengketa_nim WHERE id = $1 AND status_proses = 'pending'`, disputeID)
	if err != nil {
		return errors.New("sengketa tidak ditemukan atau sudah diproses sebelumnya")
	}

	// 🛑 SKENARIO 1: SENGKETA DITOLAK (REJECT)
	if action == "reject" {
		// Tutup kasus jadi rejected dan cabut suspend dari tabel pemilih
		_, err = tx.ExecContext(ctx, `UPDATE sengketa_nim SET status_proses = 'rejected' WHERE id = $1`, disputeID)
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, `UPDATE pemilih SET is_suspended = FALSE WHERE nim = $1`, nimSengketa)
		if err != nil {
			return err
		}

		return tx.Commit() // Selesai.
	}

	// 🗡️ SKENARIO 2: SENGKETA DITERIMA (APPROVE) - MODE PEMBANTAIAN PENYUSUP
	if action == "approve" {
		// A. Cari tau email si penyusup di tabel pemilih
		var emailPenyusup *string
		err = tx.GetContext(ctx, &emailPenyusup, `SELECT email_gmail_login FROM pemilih WHERE nim = $1`, nimSengketa)
		if err != nil {
			return err
		}

		// B. Blacklist email penyusup selamanya (Jika ada)
		if emailPenyusup != nil {
			_, err = tx.ExecContext(ctx, `INSERT INTO email_blacklist (email) VALUES ($1) ON CONFLICT (email) DO NOTHING`, *emailPenyusup)
			if err != nil {
				return err
			}
		}

		// C. Cari dan Hapus kertas suara si penyusup pakai HASH NIM
		_, err = tx.ExecContext(ctx, `DELETE FROM kertas_suara WHERE hashed_nim = $1`, nimHash)
		if err != nil {
			return err
		}

		// D. Pemutihan tabel pemilih (Kosongkan email penyusup, cabut suspend)
		_, err = tx.ExecContext(ctx, `UPDATE pemilih SET email_gmail_login = NULL, is_suspended = FALSE WHERE nim = $1`, nimSengketa)
		if err != nil {
			return err
		}

		// E. Tutup kasus sengketa menjadi approved
		_, err = tx.ExecContext(ctx, `UPDATE sengketa_nim SET status_proses = 'approved' WHERE id = $1`, disputeID)
		if err != nil {
			return err
		}

		return tx.Commit() // Eksekusi semua perubahan!
	}

	return errors.New("aksi tidak valid")
}

func (r *adminRepositoryImpl) UpdateElectionStatus(ctx context.Context, electionID int, status string) error {
	query := `UPDATE elections SET status = $1 WHERE id = $2`

	res, err := r.db.ExecContext(ctx, query, status, electionID)
	if err != nil {
		return err
	}

	// Cek apakah ID pemilunya beneran ada
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("pemilu dengan ID tersebut tidak ditemukan")
	}

	return nil
}

func (r *adminRepositoryImpl) GetDashboardAudit(ctx context.Context) (DashboardAuditResponse, error) {
	var dashboard DashboardAuditResponse

	// A. Hitung total mahasiswa yang udah berhasil Bind NIM (Terdaftar)
	err := r.db.GetContext(ctx, &dashboard.TotalPemilih, `SELECT COUNT(*) FROM pemilih`)
	if err != nil {
		return dashboard, err
	}

	// B. Hitung total suara yang udah masuk ke kotak suara
	err = r.db.GetContext(ctx, &dashboard.TotalSuara, `SELECT COUNT(*) FROM kertas_suara`)
	if err != nil {
		return dashboard, err
	}

	// C. Tarik 10 aktivitas terbaru panitia
	queryLogs := `SELECT id, admin_username, aksi, timestamp FROM audit_logs ORDER BY timestamp DESC LIMIT 10`
	err = r.db.SelectContext(ctx, &dashboard.RecentLogs, queryLogs)
	if err != nil {
		return dashboard, err
	}

	// Jaga-jaga kalau log masih kosong biar gak jadi null di React
	if dashboard.RecentLogs == nil {
		dashboard.RecentLogs = []AuditLogResponse{}
	}

	return dashboard, nil
}

func (r *adminRepositoryImpl) RecalculateVotes(ctx context.Context) ([]RecalculateResult, error) {
	var results []RecalculateResult

	// Query Sakti: Gabungin tabel kandidat dan kertas_suara, lalu hitung jumlahnya
	query := `
		SELECT 
			k.id AS id_paslon, 
			k.chairman_name, 
			k.vice_chairman_name, 
			COUNT(ks.id) AS total_suara
		FROM kandidat k
		LEFT JOIN kertas_suara ks ON k.id = ks.id_paslon
		GROUP BY k.id, k.chairman_name, k.vice_chairman_name
		ORDER BY total_suara DESC
	`

	err := r.db.SelectContext(ctx, &results, query)
	if err != nil {
		return nil, err
	}

	// Jaga-jaga kalau belum ada paslon sama sekali
	if results == nil {
		return []RecalculateResult{}, nil
	}

	return results, nil
}
