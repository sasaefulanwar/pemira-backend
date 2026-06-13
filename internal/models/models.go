package models

import "time"

type Mahasiswa struct {
	NIM       string    `db:"nim" json:"nim"`
	Nama      string    `db:"nama" json:"nama"`
	Angkatan  int       `db:"angkatan" json:"angkatan"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// Pemilih merepresentasikan tabel pemilih (User System)
type Pemilih struct {
	NIM             string    `db:"nim" json:"nim"`
	Nama            string    `db:"nama" json:"nama"`
	EmailGmailLogin *string   `db:"email_gmail_login" json:"email_gmail_login"`
	Role            string    `db:"role" json:"role"` // Tambahan kolom role
	StatusMemilih   bool      `db:"status_memilih" json:"status_memilih"`
	IsSuspended     bool      `db:"is_suspended" json:"is_suspended"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
}

// Election merepresentasikan tabel elections
type Election struct {
	ID         int        `db:"id" json:"id"`
	NamaPemilu string     `db:"nama_pemilu" json:"nama_pemilu"`
	StartAt    *time.Time `db:"start_at" json:"start_at"` // Pointer karena bisa null
	EndAt      *time.Time `db:"end_at" json:"end_at"`     // Pointer karena bisa null
	Status     string     `db:"status" json:"status"`
}

// Kandidat merepresentasikan tabel kandidat yang baru
type Kandidat struct {
	ID               int     `db:"id" json:"id"`
	ElectionID       int     `db:"election_id" json:"election_id"`
	CandidateNumber  *int    `db:"candidate_number" json:"candidate_number"`
	ChairmanName     *string `db:"chairman_name" json:"chairman_name"`
	ViceChairmanName *string `db:"vice_chairman_name" json:"vice_chairman_name"`
	Vision           *string `db:"vision" json:"vision"`
	Mission          *string `db:"mission" json:"mission"`
	PhotoURL         *string `db:"photo_url" json:"photo_url"`
}

// KertasSuara merepresentasikan tabel kertas_suara
type KertasSuara struct {
	ID         int       `db:"id" json:"id"`
	ElectionID int       `db:"election_id" json:"election_id"`
	HashedNIM  string    `db:"hashed_nim" json:"hashed_nim"`
	IDPaslon   int       `db:"id_paslon" json:"id_paslon"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

type SengketaNIM struct {
	ID           int       `db:"id" json:"id"`
	NIMSengketa  string    `db:"nim_sengketa" json:"nim_sengketa"`
	EmailPelapor string    `db:"email_pelapor" json:"email_pelapor"`
	PathFotoKTM  string    `db:"path_foto_ktm" json:"path_foto_ktm"`
	StatusProses string    `db:"status_proses" json:"status_proses"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

type EmailBlacklist struct {
	ID        int       `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	Alasan    string    `db:"alasan" json:"alasan"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type AuditLog struct {
	ID            int       `db:"id" json:"id"`
	AdminUsername string    `db:"admin_username" json:"admin_username"`
	Aksi          string    `db:"aksi" json:"aksi"`
	Timestamp     time.Time `db:"timestamp" json:"timestamp"`
}

type VoterEvent struct {
	ID        int       `db:"id" json:"id"`
	HashedNIM string    `db:"hashed_nim" json:"hashed_nim"`
	EventType string    `db:"event_type" json:"event_type"`
	IPAddress string    `db:"ip_address" json:"ip_address"`
	UserAgent string    `db:"user_agent" json:"user_agent"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
