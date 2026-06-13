package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"os"

	"pemira-backend/internal/repository"
)

// VoteService adalah kontrak untuk logika bisnis voting
type VoteService interface {
	ProcessVote(ctx context.Context, nim string, electionID int, idPaslon int, ipAddress string, userAgent string) error
}

type voteServiceImpl struct {
	repo repository.VoteRepository
}

// NewVoteService untuk inisialisasi layer service
func NewVoteService(repo repository.VoteRepository) VoteService {
	return &voteServiceImpl{repo: repo}
}

// GenerateHashedNIM mengubah NIM menjadi kode acak (Hash) menggunakan HMAC-SHA256
func GenerateHashedNIM(nim string) (string, error) {
	// Ambil secret key dari file .env
	secretKey := os.Getenv("VOTE_SECRET_KEY")
	if secretKey == "" {
		return "", errors.New("VOTE_SECRET_KEY tidak ditemukan di environment")
	}

	// Bikin HMAC baru menggunakan algoritma SHA256
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(nim))

	// Ubah hasil hash menjadi format string Hexadecimal
	hashedString := hex.EncodeToString(h.Sum(nil))
	return hashedString, nil
}

// ProcessVote adalah fungsi utama yang dipanggil oleh sistem saat user klik "VOTE"
func (s *voteServiceImpl) ProcessVote(ctx context.Context, nim string, electionID int, idPaslon int, ipAddress string, userAgent string) error {

	// 1. Samarkan NIM mahasiswa (Anonimitas)
	hashedNIM, err := GenerateHashedNIM(nim)
	if err != nil {
		log.Println("❌ Gagal melakukan hashing NIM:", err)
		return errors.New("terjadi kesalahan sistem saat memproses identitas")
	}

	// 2. Kirim NIM asli (buat update status) DAN hashedNIM (buat kertas suara) ke Repository
	// Semua validasi status udah diurus secara aman di dalam layer Repository (Transaction)
	err = s.repo.CastVote(ctx, nim, hashedNIM, electionID, idPaslon, ipAddress, userAgent)
	if err != nil {
		log.Println("❌ Transaksi Voting Gagal:", err)
		return err
	}

	return nil
}
