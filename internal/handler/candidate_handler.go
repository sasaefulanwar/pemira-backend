package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"pemira-backend/internal/repository"
	"pemira-backend/internal/service"
	"pemira-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type CandidateHandler struct {
	candidateService service.CandidateService
}

func NewCandidateHandler(cs service.CandidateService) *CandidateHandler {
	return &CandidateHandler{candidateService: cs}
}

// 1. GET (Publik): Nampilin paslon buat mahasiswa
func (h *CandidateHandler) GetCandidatesByElection(c *gin.Context) {
	electionIDStr := c.Param("id")
	electionID, err := strconv.Atoi(electionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID pemilu tidak valid"})
		return
	}

	candidates, err := h.candidateService.GetCandidatesByElection(c.Request.Context(), electionID)
	if err != nil {
		fmt.Println("ERROR DATABASE:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data kandidat"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   candidates,
	})
}

// 2. POST (Admin): Nambah Paslon + Upload Foto
func (h *CandidateHandler) CreateCandidate(c *gin.Context) {
	// 1. Ambil data teks dari form
	electionID, _ := strconv.Atoi(c.PostForm("election_id"))
	candidateNumber, _ := strconv.Atoi(c.PostForm("candidate_number"))
	chairmanName := c.PostForm("chairman_name")
	viceChairmanName := c.PostForm("vice_chairman_name")
	vision := c.PostForm("vision")
	mission := c.PostForm("mission")

	var photoURL string

	// 2. Handle File Upload ke Cloudinary
	file, err := c.FormFile("photo")
	if err == nil {
		// Buka file untuk dibaca
		openedFile, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membaca file foto"})
			return
		}
		defer openedFile.Close()

		// Upload ke Cloudinary menggunakan utils yang tadi kita buat
		cloudURL, err := utils.UploadToCloudinary(openedFile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal upload foto ke cloud: " + err.Error()})
			return
		}
		photoURL = cloudURL
	}

	// 3. Siapkan request ke service
	req := repository.CandidateResponse{
		ElectionID:       electionID,
		CandidateNumber:  candidateNumber,
		ChairmanName:     chairmanName,
		ViceChairmanName: viceChairmanName,
		Vision:           vision,
		Mission:          mission,
		PhotoURL:         photoURL, // Ini sekarang berisi URL Cloudinary
	}

	// 4. Panggil service (logika audit sudah aman di service)
	adminEmail, _ := c.Get("user_email")
	ipAddress := c.ClientIP()

	err = h.candidateService.CreateCandidate(c.Request.Context(), req, adminEmail.(string), ipAddress)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Pasangan Calon berhasil didaftarkan ke sistem.",
	})
}

// 3. PUT (Admin): Update Teks Visi Misi / Nama
func (h *CandidateHandler) UpdateCandidate(c *gin.Context) {
	candidateID, _ := strconv.Atoi(c.Param("id"))
	var req repository.CandidateResponse

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format request tidak valid"})
		return
	}

	adminEmail, _ := c.Get("user_email")
	ipAddress := c.ClientIP()

	err := h.candidateService.UpdateCandidate(c.Request.Context(), candidateID, req, adminEmail.(string), ipAddress)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Data kandidat berhasil diperbarui"})
}

// 4. DELETE (Admin): Hapus Paslon
func (h *CandidateHandler) DeleteCandidate(c *gin.Context) {
	candidateID, _ := strconv.Atoi(c.Param("id"))

	adminEmail, _ := c.Get("user_email")
	ipAddress := c.ClientIP()

	err := h.candidateService.DeleteCandidate(c.Request.Context(), candidateID, adminEmail.(string), ipAddress)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Kandidat berhasil dihapus"})
}
