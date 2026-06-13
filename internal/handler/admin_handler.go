package handler

import (
	"net/http"

	"pemira-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminService service.AdminService
}

type ResolveRequest struct {
	Action string `json:"action" binding:"required"`
}

type UpdateStatusRequest struct {
	ElectionID int    `json:"election_id" binding:"required"`
	Status     string `json:"status" binding:"required"`
}

func NewAdminHandler(as service.AdminService) *AdminHandler {
	return &AdminHandler{adminService: as}
}

func (h *AdminHandler) GetDisputes(c *gin.Context) {
	disputes, err := h.adminService.GetPendingDisputes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat daftar sengketa"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   disputes,
	})
}

func (h *AdminHandler) ResolveDispute(c *gin.Context) {
	// Ambil ID dari URL (contoh: /admin/disputes/1/resolve)
	var requestUri struct {
		ID int `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&requestUri); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID sengketa tidak valid"})
		return
	}

	// Ambil body JSON ("action": "approve" / "reject")
	var req ResolveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Aksi wajib diisi (approve/reject)"})
		return
	}

	// Ambil data admin yang lagi bertugas dari context JWT
	adminEmail, _ := c.Get("user_email")
	ipAddress := c.ClientIP()

	// Eksekusi Bos Terakhir!
	err := h.adminService.ResolveDispute(c.Request.Context(), requestUri.ID, req.Action, adminEmail.(string), ipAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Sengketa berhasil dieksekusi dengan status: " + req.Action,
	})
}

func (h *AdminHandler) UpdateElectionStatus(c *gin.Context) {
	var req UpdateStatusRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format request tidak valid. Wajib sertakan election_id dan status."})
		return
	}

	adminEmail, _ := c.Get("user_email")
	ipAddress := c.ClientIP()

	err := h.adminService.UpdateElectionStatus(c.Request.Context(), req.ElectionID, req.Status, adminEmail.(string), ipAddress)
	if err != nil {
		// Kalau error dari validasi kita sendiri (huruf kecil), kasih 400 Bad Request
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Status TPS berhasil diubah menjadi: " + req.Status,
	})
}

func (h *AdminHandler) GetDashboardAudit(c *gin.Context) {
	dashboard, err := h.adminService.GetDashboardAudit(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat data audit dashboard"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   dashboard,
	})
}

func (h *AdminHandler) RecalculateVotes(c *gin.Context) {
	adminEmail, _ := c.Get("user_email")
	ipAddress := c.ClientIP()

	results, err := h.adminService.RecalculateVotes(c.Request.Context(), adminEmail.(string), ipAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal melakukan rekapitulasi suara"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Rekapitulasi suara (Quick Count) berhasil dilakukan!",
		"data":    results,
	})
}
