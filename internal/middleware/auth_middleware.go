package middleware

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// RequireAuth adalah satpam yang akan mencegat request tanpa token JWT valid
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Coba ambil tiket JWT dari dalam Cookie
		tokenString, err := c.Cookie("jwt_session")
		if err != nil {
			// Kalau nggak bawa cookie, langsung tendang!
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Akses ditolak: Anda belum login"})
			return
		}

		csrfHeader := c.GetHeader("X-CSRF-Token")
		csrfCookie, err := c.Cookie("csrf_token")

		if err != nil || csrfHeader == "" || csrfHeader != csrfCookie {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Akses ditolak: CSRF Token tidak valid!"})
			return
		}

		// 2. Ambil kunci rahasia dari .env
		secretKey := os.Getenv("JWT_SECRET")

		// 3. Verifikasi keaslian tiket JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Pastikan algoritma enkripsinya sesuai
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("metode token tidak valid: %v", token.Header["alg"])
			}
			return []byte(secretKey), nil
		})

		// 4. Jika token rusak, kedaluwarsa, atau palsu
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Sesi Anda telah berakhir atau tidak valid"})
			return
		}

		// 5. Ekstrak data email dari dalam token (claims)
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// Simpan email ke dalam context biar bisa dipakai oleh Handler selanjutnya
			c.Set("user_email", claims["email"])
		}

		// 6. Kalau tiketnya sah, silakan masuk ke proses selanjutnya!
		c.Next()
	}
}
