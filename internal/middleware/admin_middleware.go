package middleware

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// RequireAdmin adalah satpam khusus yang mengecek role ke database
func RequireAdmin(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Ambil email dari context (Hasil tangkapan satpam pertama: RequireAuth)
		email, exists := c.Get("user_email")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Sesi tidak valid, silakan login ulang"})
			return
		}

		// 2. Cek role langsung ke database (Tabel pemilih)
		var role string
		query := `SELECT role FROM pemilih WHERE email_gmail_login = $1`

		err := db.GetContext(c.Request.Context(), &role, query, email.(string))
		if err != nil {
			if err == sql.ErrNoRows {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Akses ditolak: Akun tidak ditemukan!"})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Gagal memverifikasi otoritas admin"})
			return
		}

		// 3. Sensor Pamungkas: Tendang kalau bukan admin!
		if role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Akses Ditolak: Area VVIP ini khusus Panitia PEMIRA!",
			})
			return
		}

		// 4. Lolos! Silakan masuk ke fitur Admin
		c.Next()
	}
}
