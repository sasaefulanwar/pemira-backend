package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"pemira-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type SengketaHandler struct {
	sengketaService service.SengketaService
}

func NewSengketaHandler(ss service.SengketaService) *SengketaHandler {
	return &SengketaHandler{sengketaService: ss}
}

func (h *SengketaHandler) SubmitSengketa(c *gin.Context) {
	// 1. Ambil email pelapor dari session JWT
	emailPelapor, exists := c.Get("user_email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Sesi tidak valid, silakan login ulang"})
		return
	}

	// 2. Ambil data NIM Sengketa dari Form Text
	nimSengketa := c.PostForm("nim_sengketa")
	if nimSengketa == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "NIM sengketa wajib diisi"})
		return
	}

	// 3. Ambil file foto KTM dari Form File
	file, err := c.FormFile("foto_ktm")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File foto KTM wajib diunggah"})
		return
	}

	// 🔒 VALIDASI FILE: Pastikan file berupa gambar (jpg/jpeg/png)
	ext := filepath.Ext(file.Filename)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format file harus JPG, JPEG, atau PNG"})
		return
	}

	// 🔒 VALIDASI UKURAN: Maksimal 2MB (2 * 1024 * 1024 bytes) (BARU)
	if file.Size > 2*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ukuran file terlalu besar! Maksimal 2MB cuy."})
		return
	}

	// 4. Siapkan folder penyimpanan lokal di server
	uploadDir := "./uploads/ktm"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyiapkan folder server"})
		return
	}

	// Bikin nama file unik agar tidak bentrok (Format: NIM_TIMESTAMP.ext)
	uniqueFilename := fmt.Sprintf("%s_%d%s", nimSengketa, time.Now().Unix(), ext)
	finalPath := filepath.Join(uploadDir, uniqueFilename)

	// 5. Simpan file fisik ke folder server
	if err := c.SaveUploadedFile(file, finalPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan file di server"})
		return
	}

	// 6. Jalankan Logika Bisnis di Layer Service
	err = h.sengketaService.AjukanSengketa(c.Request.Context(), nimSengketa, emailPelapor.(string), finalPath)
	if err != nil {
		// Jika gagal di database, hapus file gambar yang terlanjur terupload biar gak menumpuk sampah
		os.Remove(finalPath)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Laporan sengketa berhasil dikirim! Panitia akan segera memeriksa KTM Anda.",
	})
}
