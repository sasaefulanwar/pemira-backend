package handler

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"io"
	"log"
	"net/http"

	"pemira-backend/internal/models"
	"pemira-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(as service.AuthService) *AuthHandler {
	return &AuthHandler{authService: as}
}

// Format JSON yang dikirim React (berisi token dari Google)
type GoogleLoginRequest struct {
	Credential string `json:"credential" validate:"required"`
}

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	var req GoogleLoginRequest

	bodyBytes, _ := io.ReadAll(c.Request.Body)

	// Karena body udah dibaca, kita harus balikin lagi biar bisa dibind
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Error binding: %v", err) // Cek errornya di terminal
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON tidak valid"})
		return
	}

	email, err := h.authService.VerifyGoogleLogin(c.Request.Context(), req.Credential)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.GetUserByEmail(email)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			user = &models.Pemilih{
				EmailGmailLogin: &email,  // Sesuaikan nama field dengan yang ada di struct models.Pemilih lu
				Role:            "voter", // Default role untuk akun baru
				NIM:             "",      // Kosongkan agar frontend me-redirect ke /bind-nim
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data pengguna"})
			return
		}
	}

	tokenString, err := h.authService.GenerateJWT(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat sesi login"})
		return
	}

	bytes := make([]byte, 16)
	rand.Read(bytes)
	csrfToken := hex.EncodeToString(bytes)

	c.SetSameSite(http.SameSiteLaxMode)

	c.SetCookie("jwt_session", tokenString, 3600*24, "/", "localhost", false, true)

	c.SetCookie("csrf_token", csrfToken, 3600*24, "/", "localhost", false, false)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Login berhasil",
		"data": gin.H{
			"email": user.EmailGmailLogin,
			"role":  user.Role,
			"nim":   user.NIM,
		},
	})
}
