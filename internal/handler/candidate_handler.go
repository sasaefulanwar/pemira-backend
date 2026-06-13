package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"pemira-backend/internal/repository"
	"pemira-backend/internal/service"

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
	// Karena ada upload file, kita gak pakai JSON, tapi pakai Form-Data
	electionID, _ := strconv.Atoi(c.PostForm("election_id"))
	candidateNumber, _ := strconv.Atoi(c.PostForm("candidate_number"))
	chairmanName := c.PostForm("chairman_name")
	viceChairmanName := c.PostForm("vice_chairman_name")
	vision := c.PostForm("vision")
	mission := c.PostForm("mission")

	var photoPath string

	// Handle File Upload Foto
	file, err := c.FormFile("photo")
	if err == nil {
		// Buat folder otomatis kalau belum ada
		os.MkdirAll("./uploads/candidates", os.ModePerm)

		// Format nama file: paslon_{electionID}_{nomorUrut}.jpg
		filename := "paslon_" + strconv.Itoa(electionID) + "_" + strconv.Itoa(candidateNumber) + filepath.Ext(file.Filename)
		savePath := "./uploads/candidates/" + filename

		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan foto kandidat"})
			return
		}
		// Path yang bakal disimpan ke database
		photoPath = "/files/candidates/" + filename
	}

	req := repository.CandidateResponse{
		ElectionID:       electionID,
		CandidateNumber:  candidateNumber,
		ChairmanName:     chairmanName,
		ViceChairmanName: viceChairmanName,
		Vision:           vision,
		Mission:          mission,
		PhotoURL:         photoPath,
	}

	adminEmail, _ := c.Get("user_email")
	ipAddress := c.ClientIP()

	err = h.candidateService.CreateCandidate(c.Request.Context(), req, adminEmail.(string), ipAddress)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Kandidat berhasil ditambahkan"})
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
