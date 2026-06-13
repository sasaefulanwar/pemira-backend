package handler

import (
	"net/http"
	"path/filepath"

	"pemira-backend/internal/service"
	"pemira-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type SengketaHandler struct {
	sengketaService service.SengketaService
}

func NewSengketaHandler(ss service.SengketaService) *SengketaHandler {
	return &SengketaHandler{sengketaService: ss}
}

func (h *SengketaHandler) SubmitSengketa(c *gin.Context) {
	emailPelapor, exists := c.Get("user_email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Sesi tidak valid"})
		return
	}

	nimSengketa := c.PostForm("nim_sengketa")
	file, err := c.FormFile("foto_ktm")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File foto KTM wajib diunggah"})
		return
	}

	// Validasi file (tetap sama, ini sudah bagus)
	ext := filepath.Ext(file.Filename)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format file harus JPG, JPEG, atau PNG"})
		return
	}
	if file.Size > 2*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ukuran file terlalu besar! Maksimal 2MB."})
		return
	}

	// --- UBAH BAGIAN INI: GANTI LOCAL SAVE KE CLOUDINARY ---

	// 1. Buka file buat dibaca
	openedFile, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membaca file"})
		return
	}
	defer openedFile.Close()

	// 2. Panggil fungsi Cloudinary yang kita buat tadi
	// (Lu tinggal import package utils lu)
	cloudURL, err := utils.UploadToCloudinary(openedFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal upload ke cloud: " + err.Error()})
		return
	}

	// 3. Simpan cloudURL ke database melalui Service
	err = h.sengketaService.AjukanSengketa(c.Request.Context(), nimSengketa, emailPelapor.(string), cloudURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Laporan sengketa berhasil diproses oleh KPU.",
	})
}
