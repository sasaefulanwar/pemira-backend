package service

import (
	"context"
	"errors"
	"strconv"

	"pemira-backend/internal/repository"
)

// Interface wajib
type CandidateService interface {
	GetCandidatesByElection(ctx context.Context, electionID int) ([]repository.CandidateResponse, error)
	CreateCandidate(ctx context.Context, candidate repository.CandidateResponse, adminEmail, ipAddress string) error
	UpdateCandidate(ctx context.Context, id int, candidate repository.CandidateResponse, adminEmail, ipAddress string) error
	DeleteCandidate(ctx context.Context, id int, adminEmail, ipAddress string) error
}

type candidateServiceImpl struct {
	candidateRepo repository.CandidateRepository
	auditRepo     repository.AuditRepository // Kita butuh ini buat nyatet log aktivitas panitia!
}

func NewCandidateService(cr repository.CandidateRepository, ar repository.AuditRepository) CandidateService {
	return &candidateServiceImpl{
		candidateRepo: cr,
		auditRepo:     ar,
	}
}

// 1. READ (Publik) - Gak perlu dicatat di audit log karena ini buat mahasiswa milih
func (s *candidateServiceImpl) GetCandidatesByElection(ctx context.Context, electionID int) ([]repository.CandidateResponse, error) {
	if electionID <= 0 {
		return nil, errors.New("ID pemilu tidak valid")
	}
	return s.candidateRepo.GetByElectionID(ctx, electionID)
}

// 2. CREATE (Admin)
func (s *candidateServiceImpl) CreateCandidate(ctx context.Context, c repository.CandidateResponse, adminEmail, ipAddress string) error {
	// Validasi dasar
	if c.ElectionID <= 0 || c.CandidateNumber <= 0 || c.ChairmanName == "" {
		return errors.New("data kandidat tidak lengkap (election_id, nomor urut, dan nama ketua wajib diisi)")
	}

	err := s.candidateRepo.Create(ctx, c)
	if err != nil {
		return err
	}

	// 📝 CATAT AUDIT TRAIL
	rincian := "Menambahkan kandidat nomor urut " + strconv.Itoa(c.CandidateNumber) + " untuk pemilu ID " + strconv.Itoa(c.ElectionID)
	s.auditRepo.LogActivity(ctx, adminEmail, "CREATE_CANDIDATE", rincian, ipAddress)

	return nil
}

// 3. UPDATE (Admin)
func (s *candidateServiceImpl) UpdateCandidate(ctx context.Context, id int, c repository.CandidateResponse, adminEmail, ipAddress string) error {
	if id <= 0 {
		return errors.New("ID kandidat tidak valid")
	}

	err := s.candidateRepo.Update(ctx, id, c)
	if err != nil {
		return err
	}

	// 📝 CATAT AUDIT TRAIL
	rincian := "Mengubah data kandidat ID " + strconv.Itoa(id)
	s.auditRepo.LogActivity(ctx, adminEmail, "UPDATE_CANDIDATE", rincian, ipAddress)

	return nil
}

// 4. DELETE (Admin)
func (s *candidateServiceImpl) DeleteCandidate(ctx context.Context, id int, adminEmail, ipAddress string) error {
	if id <= 0 {
		return errors.New("ID kandidat tidak valid")
	}

	err := s.candidateRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// 📝 CATAT AUDIT TRAIL
	rincian := "Menghapus kandidat ID " + strconv.Itoa(id)
	s.auditRepo.LogActivity(ctx, adminEmail, "DELETE_CANDIDATE", rincian, ipAddress)

	return nil
}
