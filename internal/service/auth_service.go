package service

import (
	"context"
	"errors"
	"log"
	"os"
	"pemira-backend/internal/models"
	"pemira-backend/internal/repository"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/api/idtoken"
)

// AuthService adalah kontrak untuk fitur autentikasi
type AuthService interface {
	VerifyGoogleLogin(ctx context.Context, googleToken string) (string, error)
	GenerateJWT(email string) (string, error)
	GetUserByEmail(email string) (*models.Pemilih, error)
}

type authServiceImpl struct {
	repo repository.AuthRepository
}

func NewAuthService(repo repository.AuthRepository) AuthService {
	return &authServiceImpl{repo: repo}
}

// VerifyGoogleLogin memvalidasi token dari frontend ke server Google
func (s *authServiceImpl) VerifyGoogleLogin(ctx context.Context, googleToken string) (string, error) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	if clientID == "" {
		return "", errors.New("GOOGLE_CLIENT_ID belum di-set")
	}

	payload, err := idtoken.Validate(ctx, googleToken, clientID)
	if err != nil {
		log.Println("❌ Token Google tidak valid:", err)
		return "", errors.New("autentikasi google gagal")
	}

	email, ok := payload.Claims["email"].(string)
	if !ok {
		return "", errors.New("tidak dapat membaca email dari akun Google")
	}

	isBlacklisted, err := s.repo.IsEmailBlacklisted(ctx, email)
	if err != nil {
		return "", errors.New("terjadi kesalahan saat memvalidasi keamanan akun")
	}
	if isBlacklisted {
		return "", errors.New("akses ditolak: email ini telah masuk daftar hitam sistem")
	}

	return email, nil
}

func (s *authServiceImpl) GenerateJWT(email string) (string, error) {
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		return "", errors.New("JWT_SECRET belum di-set")
	}

	// Bikin struktur tokennya (isinya email dan waktu kadaluarsa)
	claims := jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(time.Hour * 24).Unix(), // Token berlaku 24 jam
	}

	// Proses pembuatan token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *authServiceImpl) GetUserByEmail(email string) (*models.Pemilih, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}
	return user, nil
}
