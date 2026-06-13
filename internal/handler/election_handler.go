package handler

import (
	"net/http"
	"strconv"

	"pemira-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type ElectionHandler struct {
	electionService service.ElectionService
}

func NewElectionHandler(es service.ElectionService) *ElectionHandler {
	return &ElectionHandler{electionService: es}
}

func (h *ElectionHandler) GetPublicResults(c *gin.Context) {
	// Ambil ID Pemilu dari URL (contoh: /elections/1/results)
	electionIDStr := c.Param("id")
	electionID, err := strconv.Atoi(electionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID pemilu tidak valid"})
		return
	}

	results, err := h.electionService.GetPublicResults(c.Request.Context(), electionID)
	if err != nil {
		// Tolak dengan 403 Forbidden kalau TPS masih buka
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "TPS telah ditutup. Berikut adalah hasil akhir PEMIRA!",
		"data":    results,
	})
}
