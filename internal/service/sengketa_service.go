package service

import (
	"context"
	"errors"

	"pemira-backend/internal/repository"
)

type SengketaService interface {
	AjukanSengketa(ctx context.Context, nimSengketa string, emailPelapor string, pathFotoKTM string) error
}

type sengketaServiceImpl struct {
	sengketaRepo repository.SengketaRepository
	pemilihRepo  repository.PemilihRepository // Kita pinjam repo pemilih buat ngecek data
}

func NewSengketaService(sr repository.SengketaRepository, pr repository.PemilihRepository) SengketaService {
	return &sengketaServiceImpl{sengketaRepo: sr, pemilihRepo: pr}
}

func (s *sengketaServiceImpl) AjukanSengketa(ctx context.Context, nimSengketa string, emailPelapor string, pathFotoKTM string) error {
	// 1. Validasi: Pastikan NIM yang digugat emang beneran udah terikat di tabel pemilih
	pemilih, err := s.pemilihRepo.FindPemilihByNIM(ctx, nimSengketa)
	if err != nil {
		return err
	}
	if pemilih == nil {
		return errors.New("NIM ini belum diklaim oleh siapapun. Silakan lakukan Binding NIM secara normal saja")
	}

	// 2. Validasi: Pelapor tidak boleh menggugat NIM-nya sendiri jika emailnya sama
	if pemilih.EmailGmailLogin != nil && *pemilih.EmailGmailLogin == emailPelapor {
		return errors.New("Anda tidak bisa menggugat akun Anda sendiri")
	}

	// 3. Simpan laporan sengketa ke database
	return s.sengketaRepo.CreateSengketa(ctx, nimSengketa, emailPelapor, pathFotoKTM)
}
