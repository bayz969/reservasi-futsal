package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"futsalbook/middleware"
	"futsalbook/models"
	"futsalbook/repository"

	"golang.org/x/crypto/bcrypt"
)

// In-memory store for reset password tokens (for simulation)
// In production, this could be in a database table or Redis
var resetTokens = make(map[string]string) // token -> email
var resetTokensMu sync.Mutex

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

// Register registers a new user
func Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Format request tidak valid")
		return
	}

	if req.Nama == "" || req.Email == "" || req.NomorHP == "" || req.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Semua field harus diisi")
		return
	}

	if len(req.Password) < 6 {
		respondWithError(w, http.StatusBadRequest, "Password minimal harus 6 karakter")
		return
	}

	// Check if user already exists
	_, err := repository.GetUserByEmail(req.Email)
	if err == nil {
		respondWithError(w, http.StatusConflict, "Email sudah terdaftar")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Gagal memproses password")
		return
	}

	user := models.User{
		Nama:     req.Nama,
		Email:    req.Email,
		Password: string(hashedPassword),
		NomorHP:  req.NomorHP,
		Role:     "user", // default role
		Status:   "active",
	}

	if err := repository.CreateUser(&user); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Gagal membuat user")
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Registrasi berhasil! Silakan login.",
		"user":    user,
	})
}

// Login authenticates users and admins
func Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Format request tidak valid")
		return
	}

	user, err := repository.GetUserByEmail(req.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Email atau password salah")
		return
	}

	if user.Status != "active" {
		respondWithError(w, http.StatusForbidden, "Akun Anda dinonaktifkan. Silakan hubungi admin.")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Email atau password salah")
		return
	}

	token, err := middleware.GenerateJWT(user)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Gagal membuat sesi login")
		return
	}

	// Set cookie for browser convenience
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		Path:     "/",
		HttpOnly: false, // Accessible by frontend JS for routing, or secure as needed
	})

	respondWithJSON(w, http.StatusOK, models.LoginResponse{
		Token: token,
		User:  user,
	})
}

// ForgotPassword requests a reset password token and prints the link to server logs
func ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req models.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Format request tidak valid")
		return
	}

	user, err := repository.GetUserByEmail(req.Email)
	if err != nil {
		// To prevent user enumeration, we return success even if user not found,
		// but for development testing let's return a nice simulation status
		respondWithError(w, http.StatusNotFound, "Email tidak terdaftar")
		return
	}

	// Generate safe random token
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	token := hex.EncodeToString(b)

	resetTokensMu.Lock()
	resetTokens[token] = user.Email
	resetTokensMu.Unlock()

	// Simulate sending email by printing reset link to logs
	resetLink := "http://localhost:8080/#/reset-password?token=" + token
	log.Printf("[SIMULASI EMAIL] Reset Password untuk %s: %s", user.Email, resetLink)

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Kode/Tautan reset password telah dikirim ke email. (Periksa log konsol server untuk tautannya!)",
		"token_simulation": token, // Return for easier frontend integration/testing
	})
}

// ResetPassword resets the user's password using the token
func ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req models.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Format request tidak valid")
		return
	}

	if req.Token == "" || req.NewPassword == "" {
		respondWithError(w, http.StatusBadRequest, "Token dan password baru harus diisi")
		return
	}

	resetTokensMu.Lock()
	email, exists := resetTokens[req.Token]
	if exists {
		delete(resetTokens, req.Token) // consume token
	}
	resetTokensMu.Unlock()

	if !exists {
		respondWithError(w, http.StatusBadRequest, "Token reset password tidak valid atau telah kedaluwarsa")
		return
	}

	user, err := repository.GetUserByEmail(email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "User tidak ditemukan")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Gagal memproses password")
		return
	}

	user.Password = string(hashedPassword)
	if err := repository.UpdateUser(&user); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Gagal memperbarui password")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Password berhasil diperbarui! Silakan login dengan password baru Anda.",
	})
}

// Logout invalidates the local session cookie
func Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		Path:     "/",
		HttpOnly: false,
	})

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Logout berhasil",
	})
}

// GetMe gets the profile of currently logged-in user
func GetMe(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value(middleware.UserIDKey)
	if userIDVal == nil {
		respondWithError(w, http.StatusUnauthorized, "Sesi tidak ditemukan")
		return
	}

	userID := userIDVal.(int64)
	user, err := repository.GetUserByID(userID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "User tidak ditemukan")
		return
	}

	respondWithJSON(w, http.StatusOK, user)
}

// UpdateMe updates profile of currently logged-in user
func UpdateMe(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value(middleware.UserIDKey)
	if userIDVal == nil {
		respondWithError(w, http.StatusUnauthorized, "Sesi tidak ditemukan")
		return
	}

	userID := userIDVal.(int64)
	user, err := repository.GetUserByID(userID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "User tidak ditemukan")
		return
	}

	var req models.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Format request tidak valid")
		return
	}

	if req.Nama == "" || req.NomorHP == "" {
		respondWithError(w, http.StatusBadRequest, "Nama dan Nomor HP tidak boleh kosong")
		return
	}

	user.Nama = req.Nama
	user.NomorHP = req.NomorHP

	if req.Password != "" {
		if len(req.Password) < 6 {
			respondWithError(w, http.StatusBadRequest, "Password minimal 6 karakter")
			return
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Gagal memproses password")
			return
		}
		user.Password = string(hashedPassword)
	} else {
		user.Password = "" // Set to empty so repository.UpdateUser knows not to update password
	}

	if err := repository.UpdateUser(&user); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Gagal memperbarui profil")
		return
	}

	// Fetch updated user to return
	updatedUser, _ := repository.GetUserByID(userID)
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Profil berhasil diperbarui",
		"user":    updatedUser,
	})
}

// ListUsers lists all registered users (Admin only)
func ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := repository.ListUsers()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Gagal mengambil data user")
		return
	}
	respondWithJSON(w, http.StatusOK, users)
}

// ToggleUserStatus activates or deactivates a user (Admin only)
func ToggleUserStatus(w http.ResponseWriter, r *http.Request, userIDStr string) {
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID user tidak valid")
		return
	}

	var req models.UserStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Format request tidak valid")
		return
	}

	if req.Status != "active" && req.Status != "inactive" {
		respondWithError(w, http.StatusBadRequest, "Status harus active atau inactive")
		return
	}

	// Check if trying to disable self
	myIDVal := r.Context().Value(middleware.UserIDKey)
	if myIDVal != nil && myIDVal.(int64) == userID {
		respondWithError(w, http.StatusBadRequest, "Anda tidak dapat menonaktifkan akun Anda sendiri")
		return
	}

	if err := repository.UpdateUserStatus(userID, req.Status); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Gagal mengubah status user")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Status user berhasil diperbarui menjadi " + req.Status,
	})
}
