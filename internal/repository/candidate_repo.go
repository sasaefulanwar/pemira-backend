package repository

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
)

// DTO untuk Response Publik & Admin
type CandidateResponse struct {
	ID               int    `json:"id" db:"id"`
	ElectionID       int    `json:"election_id" db:"election_id"`
	CandidateNumber  int    `json:"candidate_number" db:"candidate_number"`
	ChairmanName     string `json:"chairman_name" db:"chairman_name"`
	ViceChairmanName string `json:"vice_chairman_name" db:"vice_chairman_name"`
	Vision           string `json:"vision" db:"vision"`
	Mission          string `json:"mission" db:"mission"`
	PhotoURL         string `json:"photo_url" db:"photo_url"`
}

// Interface wajib CRUD
type CandidateRepository interface {
	GetByElectionID(ctx context.Context, electionID int) ([]CandidateResponse, error)
	Create(ctx context.Context, candidate CandidateResponse) error
	Update(ctx context.Context, id int, candidate CandidateResponse) error
	Delete(ctx context.Context, id int) error
}

type candidateRepositoryImpl struct {
	db *sqlx.DB
}

func NewCandidateRepository(db *sqlx.DB) CandidateRepository {
	return &candidateRepositoryImpl{db: db}
}

// 1. READ: Nampilin daftar kandidat berdasarkan ID Pemilu
func (r *candidateRepositoryImpl) GetByElectionID(ctx context.Context, electionID int) ([]CandidateResponse, error) {
	var candidates []CandidateResponse
	query := `SELECT * FROM kandidat WHERE election_id = $1 ORDER BY candidate_number ASC`

	err := r.db.SelectContext(ctx, &candidates, query, electionID)
	if err != nil {
		return nil, err
	}
	if candidates == nil {
		return []CandidateResponse{}, nil
	}
	return candidates, nil
}

// 2. CREATE: Panitia nambahin kandidat baru
func (r *candidateRepositoryImpl) Create(ctx context.Context, c CandidateResponse) error {
	query := `
		INSERT INTO kandidat (election_id, candidate_number, chairman_name, vice_chairman_name, vision, mission, photo_url) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query, c.ElectionID, c.CandidateNumber, c.ChairmanName, c.ViceChairmanName, c.Vision, c.Mission, c.PhotoURL)
	return err
}

// 3. UPDATE: Panitia ngedit teks kalau ada typo (tanpa ubah foto dulu biar simpel)
func (r *candidateRepositoryImpl) Update(ctx context.Context, id int, c CandidateResponse) error {
	query := `
		UPDATE kandidat 
		SET candidate_number = $1, chairman_name = $2, vice_chairman_name = $3, vision = $4, mission = $5
		WHERE id = $6
	`
	res, err := r.db.ExecContext(ctx, query, c.CandidateNumber, c.ChairmanName, c.ViceChairmanName, c.Vision, c.Mission, id)
	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("kandidat tidak ditemukan")
	}
	return nil
}

// 4. DELETE: Panitia ngehapus kandidat yang didiskualifikasi
func (r *candidateRepositoryImpl) Delete(ctx context.Context, id int) error {
	// PENTING: Cek dulu apakah dia udah dapet suara? Kalau udah, gabisa dihapus sembarangan!
	// Tapi untuk level crud dasar, kita hajar aja langsung:
	query := `DELETE FROM kandidat WHERE id = $1`

	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		// Biasanya error karena melanggar Foreign Key di tabel kertas_suara
		return errors.New("gagal menghapus! Kandidat ini mungkin sudah menerima suara")
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("kandidat tidak ditemukan")
	}
	return nil
}
