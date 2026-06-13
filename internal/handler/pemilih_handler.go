package handler

import (
	"net/http"

	"pemira-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type PemilihHandler struct {
	pemilihService service.PemilihService
}

func NewPemilihHandler(ps service.PemilihService) *PemilihHandler {
	return &PemilihHandler{pemilihService: ps}
}

type BindNIMRequest struct {
	NIM string `json:"nim" binding:"required"`
}

func (h *PemilihHandler) BindingNIM(c *gin.Context) {
	// 1. Ambil email user dari context (Hasil tangkapan Satpam JWT)
	email, exists := c.Get("user_email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Sesi tidak valid, silakan login ulang"})
		return
	}

	// 2. Baca input NIM dari Body JSON
	var req BindNIMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format NIM tidak valid"})
		return
	}

	ipAddress := c.ClientIP()

	err := h.pemilihService.BindNIM(c.Request.Context(), req.NIM, email.(string), ipAddress)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "NIM berhasil dikaitkan dengan akun Google Anda!",
	})
}
