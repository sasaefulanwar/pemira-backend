package handler

import (
	"net/http"

	"pemira-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// VoteHandler adalah struktur untuk handler voting
type VoteHandler struct {
	voteService service.VoteService
}

// NewVoteHandler untuk inisialisasi handler
func NewVoteHandler(vs service.VoteService) *VoteHandler {
	return &VoteHandler{voteService: vs}
}

// CastVoteRequest adalah format JSON yang diharapkan dari frontend React
type CastVoteRequest struct {
	NIM        string `json:"nim" binding:"required"`
	ElectionID int    `json:"election_id" binding:"required"`
	IDPaslon   int    `json:"id_paslon" binding:"required"`
}

// CastVote adalah fungsi yang akan dieksekusi saat ada request POST ke /api/v1/votes/cast
func (h *VoteHandler) CastVote(c *gin.Context) {
	var req CastVoteRequest

	// 1. Validasi input JSON dari frontend
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid atau kurang lengkap"})
		return
	}

	// 2. Ambil IP Address dan User Agent dari request untuk keperluan Audit Log
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// 3. Lempar ke Layer Service untuk diproses (Hashing & Save ke DB)
	err := h.voteService.ProcessVote(c.Request.Context(), req.NIM, req.ElectionID, req.IDPaslon, ipAddress, userAgent)
	if err != nil {
		// Jika gagal, kembalikan status 500 (Internal Server Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 4. Jika sukses, kembalikan status 200 (OK)
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Suara berhasil diamankan secara anonim!",
	})
}
