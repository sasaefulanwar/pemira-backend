package service

import (
	"context"
	"errors"

	"pemira-backend/internal/repository"
)

type PemilihService interface {
	BindNIM(ctx context.Context, nim string, email string, ip string) error
}

type pemilihServiceImpl struct {
	repo      repository.PemilihRepository
	auditRepo repository.AuditRepository
}

func NewPemilihService(repo repository.PemilihRepository, auditRepo repository.AuditRepository) PemilihService {
	return &pemilihServiceImpl{repo: repo, auditRepo: auditRepo}
}

func (s *pemilihServiceImpl) BindNIM(ctx context.Context, nim string, email string, ip string) error {
	// SENSOR 1: Apakah email lu udah pernah nge-bind NIM lain sebelumnya? (BARU)
	existingEmail, err := s.repo.FindPemilihByEmail(ctx, email)
	if err != nil {
		return err
	}
	if existingEmail != nil {
		return errors.New("akun Google Anda sudah terikat dengan NIM lain. Satu akun Google hanya untuk satu NIM!")
	}

	// SENSOR 2: Apakah NIM yang diinput terdaftar di DPT Master?
	mhs, err := s.repo.FindMahasiswaByNIM(ctx, nim)
	if err != nil {
		return err
	}
	if mhs == nil {
		return errors.New("NIM Anda tidak terdaftar dalam DPT Master. Hubungi panitia KPUM!")
	}

	// SENSOR 3: Apakah NIM ini sudah pernah di-bind oleh email orang lain?
	pemilih, err := s.repo.FindPemilihByNIM(ctx, nim)
	if err != nil {
		return err
	}
	if pemilih != nil {
		return errors.New("NIM sudah terikat dengan akun Google lain. Silakan ajukan Sengketa Akun!")
	}

	// Jika lolos semua sensor, baru kawinkan data
	err = s.repo.CreatePemilih(ctx, mhs.NIM, mhs.Nama, email)
	if err != nil {
		return err
	}

	s.auditRepo.LogActivity(ctx, email, "BINDING_NIM", "Berhasil mengikat NIM "+nim, ip)

	return nil
}
