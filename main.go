package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"pemira-backend/internal/config"
	"pemira-backend/internal/handler"
	"pemira-backend/internal/middleware"
	"pemira-backend/internal/repository"
	"pemira-backend/internal/service"
)

func main() {
	// 1. Load .env
	if err := godotenv.Load(".env"); err != nil {
		log.Println("⚠️ Warning: File .env tidak ditemukan")
	}

	// 2. Connect ke PostgreSQL
	db := config.ConnectDB()
	defer db.Close()

	auditRepo := repository.NewAuditRepository(db)

	// 3. Inisialisasi semua Layer (Dependency Injection)
	voteRepo := repository.NewVoteRepository(db)
	voteService := service.NewVoteService(voteRepo)
	voteHandler := handler.NewVoteHandler(voteService)

	// Layer Auth (BARU)
	authRepo := repository.NewAuthRepository(db)
	authService := service.NewAuthService(authRepo)
	authHandler := handler.NewAuthHandler(authService)

	pemilihRepo := repository.NewPemilihRepository(db)
	pemilihService := service.NewPemilihService(pemilihRepo, auditRepo)
	pemilihHandler := handler.NewPemilihHandler(pemilihService)

	sengketaRepo := repository.NewSengketaRepository(db)
	sengketaService := service.NewSengketaService(sengketaRepo, pemilihRepo) // Ambil pemilihRepo juga
	sengketaHandler := handler.NewSengketaHandler(sengketaService)

	adminRepo := repository.NewAdminRepository(db)
	adminService := service.NewAdminService(adminRepo, auditRepo)
	adminHandler := handler.NewAdminHandler(adminService)

	electionRepo := repository.NewElectionRepository(db)
	electionService := service.NewElectionService(electionRepo)
	electionHandler := handler.NewElectionHandler(electionService)

	// Inisialisasi Candidate CRUD
	candidateRepo := repository.NewCandidateRepository(db)
	// Pastikan lu masukin auditRepo sebagai parameter kedua biar log-nya jalan!
	candidateService := service.NewCandidateService(candidateRepo, auditRepo)
	candidateHandler := handler.NewCandidateHandler(candidateService)

	// 4. Setup Gin Router
	router := gin.Default()

	config := cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-CSRF-Token"},
		ExposeHeaders:    []string{"Content-Length", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	router.Use(cors.New(config))

	// 5. Daftarkan Endpoint API
	api := router.Group("/api/v1")
	{
		// Limit 5 per menit
		api.POST("/auth/google", middleware.RateLimiter(5, time.Minute), authHandler.GoogleLogin)

		protected := api.Group("/")
		protected.Use(middleware.RequireAuth())
		{
			protected.POST("/pemilih/bind", middleware.RateLimiter(3, time.Minute), pemilihHandler.BindingNIM)
			protected.POST("/votes/cast", middleware.RateLimiter(3, time.Minute), voteHandler.CastVote)
			protected.GET("/elections/:id/candidates", candidateHandler.GetCandidatesByElection)
			protected.GET("/elections/:id/results", electionHandler.GetPublicResults)
			protected.POST("/pemilih/sengketa", sengketaHandler.SubmitSengketa)
		}

		adminArea := api.Group("/admin")
		adminArea.Use(middleware.RequireAuth())
		adminArea.Use(middleware.RequireAdmin(db))
		{
			// Menampilkan daftar sengketa yang pending
			adminArea.GET("/disputes", adminHandler.GetDisputes)
			adminArea.POST("/disputes/:id/resolve", adminHandler.ResolveDispute)

			// Endpoint khusus untuk melihat foto KTM (Menggantikan fungsi S3 Presigned URL)
			// Ini aman karena dibungkus oleh RequireAuth & RequireAdmin
			adminArea.StaticFS("/files/ktm", http.Dir("./uploads/ktm"))
			adminArea.StaticFS("/files/candidates", http.Dir("./uploads/candidates"))

			adminArea.PUT("/elections/status", adminHandler.UpdateElectionStatus)
			adminArea.POST("/elections/recalculate", adminHandler.RecalculateVotes)

			adminArea.GET("/audit", adminHandler.GetDashboardAudit)

			adminArea.POST("/candidates", candidateHandler.CreateCandidate)
			adminArea.PUT("/candidates/:id", candidateHandler.UpdateCandidate)
			adminArea.DELETE("/candidates/:id", candidateHandler.DeleteCandidate)
		}
	}

	// 6. Jalankan Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 Server PEMIRA berjalan di port %s...", port)
	router.Run(":" + port)
}
