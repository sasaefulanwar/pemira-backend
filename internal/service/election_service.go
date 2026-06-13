package service

import (
	"context"
	"errors"

	"pemira-backend/internal/repository"
)

type ElectionService interface {
	GetPublicResults(ctx context.Context, electionID int) ([]repository.ElectionResult, error)
}

type electionServiceImpl struct {
	electionRepo repository.ElectionRepository
}

func NewElectionService(er repository.ElectionRepository) ElectionService {
	return &electionServiceImpl{electionRepo: er}
}

func (s *electionServiceImpl) GetPublicResults(ctx context.Context, electionID int) ([]repository.ElectionResult, error) {
	// 1. Cek status TPS ke database
	status, err := s.electionRepo.GetElectionStatus(ctx, electionID)
	if err != nil {
		return nil, err
	}

	// 2. Sensor Ketat: Kalau belum closed, tolak mentah-mentah!
	if status != "closed" {
		return nil, errors.New("Sabar cuy! TPS masih buka. Hasil akhir baru bisa dilihat setelah pemungutan suara resmi ditutup oleh panitia.")
	}

	// 3. Kalau udah closed, hitung dan keluarin datanya
	return s.electionRepo.GetFinalResults(ctx, electionID)
}
