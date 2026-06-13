package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

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
	Credential string `json:"credential" binding:"required"`
}

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	var req GoogleLoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token Google tidak ditemukan"})
		return
	}

	// 1. Verifikasi keaslian token ke Google & cek Blacklist
	email, err := h.authService.VerifyGoogleLogin(c.Request.Context(), req.Credential)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// 2. Buat tiket JWT untuk sesi aplikasi kita
	tokenString, err := h.authService.GenerateJWT(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat sesi login"})
		return
	}

	bytes := make([]byte, 16)
	rand.Read(bytes)
	csrfToken := hex.EncodeToString(bytes)

	// Set SameSite=Lax sebelum pasang Cookie
	c.SetSameSite(http.SameSiteLaxMode)

	// Cookie 1: JWT Session (HttpOnly = true -> Gak bisa dibaca React)
	c.SetCookie("jwt_session", tokenString, 3600*24, "/", "localhost", false, true)

	// Cookie 2: CSRF Token (HttpOnly = false -> BISA dibaca React)
	c.SetCookie("csrf_token", csrfToken, 3600*24, "/", "localhost", false, false)

	// 4. Kirim respons sukses ke React
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Login berhasil",
		"data": gin.H{
			"email": email,
		},
	})
}
