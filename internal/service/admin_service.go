package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"pemira-backend/internal/repository"
)

type AdminService interface {
	GetPendingDisputes(ctx context.Context) ([]repository.SengketaResponse, error)
	ResolveDispute(ctx context.Context, disputeID int, action string, adminEmail string, ipAddress string) error
	UpdateElectionStatus(ctx context.Context, electionID int, status string, adminEmail string, ipAddress string) error
	GetDashboardAudit(ctx context.Context) (repository.DashboardAuditResponse, error)
	RecalculateVotes(ctx context.Context, adminEmail string, ipAddress string) ([]repository.RecalculateResult, error)
}

type adminServiceImpl struct {
	adminRepo repository.AdminRepository
	auditRepo repository.AuditRepository
}

func NewAdminService(ar repository.AdminRepository, auditRepo repository.AuditRepository) AdminService {
	return &adminServiceImpl{adminRepo: ar, auditRepo: auditRepo}
}

func (s *adminServiceImpl) GetPendingDisputes(ctx context.Context) ([]repository.SengketaResponse, error) {
	return s.adminRepo.GetAllPendingDisputes(ctx)
}

func (s *adminServiceImpl) ResolveDispute(ctx context.Context, disputeID int, action string, adminEmail string, ipAddress string) error {
	if action != "approve" && action != "reject" {
		return errors.New("aksi hanya bisa 'approve' atau 'reject'")
	}

	var nimHash string

	// Kalau Approve, kita butuh HASH dari NIM yang disengketakan buat ngehapus suara penyusup!
	if action == "approve" {
		// 1. Ambil detail sengketa dari repo (Kita pakai trik manggil GetAll dan filter manual aja biar cepat)
		disputes, err := s.adminRepo.GetAllPendingDisputes(ctx)
		if err != nil {
			return err
		}

		var targetNIM string
		for _, d := range disputes {
			if d.ID == disputeID {
				targetNIM = d.NIMSengketa
				break
			}
		}

		if targetNIM == "" {
			return errors.New("sengketa tidak ditemukan atau sudah diproses")
		}

		// 2. Generate Hash rahasianya (Mirip kayak pas nyoblos)
		secretKey := os.Getenv("HMAC_SECRET")
		if secretKey == "" {
			secretKey = "rahasia-pemira-super-aman"
		} // Samakan dengan yang di vote_service!

		h := hmac.New(sha256.New, []byte(secretKey))
		h.Write([]byte(targetNIM))
		nimHash = hex.EncodeToString(h.Sum(nil))
	}

	// Eksekusi transaksinya
	err := s.adminRepo.ResolveDispute(ctx, disputeID, action, nimHash)
	if err != nil {
		return err
	}

	// Catat log
	rincian := "Menolak sengketa ID "
	if action == "approve" {
		rincian = "Menyetujui sengketa dan membanned penyusup untuk ID "
	}
	s.auditRepo.LogActivity(ctx, adminEmail, "RESOLVE_DISPUTE", rincian, ipAddress)

	return nil
}

func (s *adminServiceImpl) UpdateElectionStatus(ctx context.Context, electionID int, status string, adminEmail string, ipAddress string) error {
	// Sensor Status: Cuma boleh 4 kata sakti ini
	validStatuses := map[string]bool{
		"draft":    true,
		"open":     true,
		"closed":   true,
		"archived": true,
	}

	if !validStatuses[status] {
		return errors.New("status tidak valid! Gunakan: draft, open, closed, atau archived")
	}

	// Eksekusi ke database
	err := s.adminRepo.UpdateElectionStatus(ctx, electionID, status)
	if err != nil {
		return err
	}

	// 📝 CATAT AUDIT TRAIL!
	// (Menggunakan format 5 parameter karena penggabungannya udah lu handle di repo kemaren)
	rincian := "Mengubah status pemilu ID " + string(rune(electionID+'0')) + " menjadi " + status
	s.auditRepo.LogActivity(ctx, adminEmail, "UPDATE_ELECTION_STATUS", rincian, ipAddress)

	return nil
}

func (s *adminServiceImpl) GetDashboardAudit(ctx context.Context) (repository.DashboardAuditResponse, error) {
	return s.adminRepo.GetDashboardAudit(ctx)
}

func (s *adminServiceImpl) RecalculateVotes(ctx context.Context, adminEmail string, ipAddress string) ([]repository.RecalculateResult, error) {
	results, err := s.adminRepo.RecalculateVotes(ctx)
	if err != nil {
		return nil, err
	}

	// 📝 CATAT AUDIT TRAIL!
	s.auditRepo.LogActivity(ctx, adminEmail, "RECALCULATE_VOTES", "Memulai proses rekapitulasi suara (Quick Count)", ipAddress)

	return results, nil
}
