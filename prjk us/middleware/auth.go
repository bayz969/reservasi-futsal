package middleware

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"futsalbook/models"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("futsalbook_secret_key")

func init() {
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		jwtKey = []byte(secret)
	}
}

// Claims defines JWT claims structure
type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateJWT creates a new JWT token for a user
func GenerateJWT(user models.User) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// VerifyJWT parses and validates the JWT token
func VerifyJWT(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

type contextKey string

const (
	UserIDKey contextKey = "user_id"
	RoleKey   contextKey = "role"
	EmailKey  contextKey = "email"
)

// AuthMiddleware authenticates requests using JWT in Auth header or cookie
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenStr string

		// 1. Check Authorization Header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenStr = parts[1]
			}
		}

		// 2. Check Cookie if header is empty
		if tokenStr == "" {
			cookie, err := r.Cookie("token")
			if err == nil {
				tokenStr = cookie.Value
			}
		}

		if tokenStr == "" {
			http.Error(w, `{"error":"Autentikasi diperlukan. Silakan login terlebih dahulu."}`, http.StatusUnauthorized)
			return
		}

		claims, err := VerifyJWT(tokenStr)
		if err != nil {
			http.Error(w, `{"error":"Token tidak valid atau kedaluwarsa. Silakan login kembali."}`, http.StatusUnauthorized)
			return
		}

		// Inject user info into context
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, RoleKey, claims.Role)
		ctx = context.WithValue(ctx, EmailKey, claims.Email)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AdminMiddleware restricts access to admins only
func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := r.Context().Value(RoleKey)
		if role == nil || role.(string) != "admin" {
			http.Error(w, `{"error":"Akses ditolak. Halaman ini hanya untuk Admin."}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
