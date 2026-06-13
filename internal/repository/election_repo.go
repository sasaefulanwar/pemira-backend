package repository

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
)

type ElectionResult struct {
	IDPaslon         int    `json:"id_paslon" db:"id_paslon"`
	ChairmanName     string `json:"chairman_name" db:"chairman_name"`
	ViceChairmanName string `json:"vice_chairman_name" db:"vice_chairman_name"`
	TotalSuara       int    `json:"total_suara" db:"total_suara"`
}

type ElectionRepository interface {
	GetElectionStatus(ctx context.Context, electionID int) (string, error)
	GetFinalResults(ctx context.Context, electionID int) ([]ElectionResult, error)
}

type electionRepositoryImpl struct {
	db *sqlx.DB
}

func NewElectionRepository(db *sqlx.DB) ElectionRepository {
	return &electionRepositoryImpl{db: db}
}

func (r *electionRepositoryImpl) GetElectionStatus(ctx context.Context, electionID int) (string, error) {
	var status string
	err := r.db.GetContext(ctx, &status, `SELECT status FROM elections WHERE id = $1`, electionID)
	if err != nil {
		return "", errors.New("pemilu tidak ditemukan")
	}
	return status, nil
}

func (r *electionRepositoryImpl) GetFinalResults(ctx context.Context, electionID int) ([]ElectionResult, error) {
	var results []ElectionResult

	// Query sama kayak admin, tapi kita filter spesifik pakai election_id
	query := `
		SELECT 
			k.id AS id_paslon, 
			k.chairman_name, 
			k.vice_chairman_name, 
			COUNT(ks.id) AS total_suara
		FROM kandidat k
		LEFT JOIN kertas_suara ks ON k.id = ks.id_paslon
		WHERE k.election_id = $1
		GROUP BY k.id, k.chairman_name, k.vice_chairman_name
		ORDER BY total_suara DESC
	`

	err := r.db.SelectContext(ctx, &results, query, electionID)
	if err != nil {
		return nil, err
	}
	if results == nil {
		return []ElectionResult{}, nil
	}
	return results, nil
}
