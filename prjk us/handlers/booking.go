package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"futsalbook/middleware"
	"futsalbook/models"
	"futsalbook/repository"
)

// CreateBooking handles reservation requests
func CreateBooking(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value(middleware.UserIDKey)
	if userIDVal == nil {
		respondWithError(w, http.StatusUnauthorized, "Sesi tidak ditemukan")
		return
	}
	userID := userIDVal.(int64)

	var req models.BookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Format request tidak valid")
		return
	}

	if req.FieldID <= 0 || req.Tanggal == "" || req.JamMulai == "" || req.Durasi <= 0 {
		respondWithError(w, http.StatusBadRequest, "FieldID, Tanggal, JamMulai, dan Durasi wajib diisi")
		return
	}

	// 1. Verify field exists and is available
	field, err := repository.GetFieldByID(req.FieldID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Lapangan tidak ditemukan")
		return
	}

	if field.Status != "tersedia" {
		respondWithError(w, http.StatusBadRequest, "Lapangan sedang tidak tersedia untuk disewa")
		return
	}

	// 2. Validate tanggal format (YYYY-MM-DD)
	parsedDate, err := time.Parse("2006-01-02", req.Tanggal)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Format tanggal tidak valid (harus YYYY-MM-DD)")
		return
	}

	// Ensure booking date is not in the past
	todayStr := time.Now().Format("2006-01-02")
	if req.Tanggal < todayStr {
		respondWithError(w, http.StatusBadRequest, "Tidak dapat membuat booking untuk tanggal di masa lalu")
		return
	}

	// 3. Parse and calculate end time
	timeLayout := "15:04"
	startTime, err := time.Parse(timeLayout, req.JamMulai)
	if err != nil {
		// Try format without leading zero
		startTime, err = time.Parse("15:4", req.JamMulai)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Format jam mulai tidak valid (harus HH:MM)")
			return
		}
	}
	// Standardize to HH:MM format
	req.JamMulai = startTime.Format(timeLayout)

	endTime := startTime.Add(time.Duration(req.Durasi) * time.Hour)
	jamSelesai := endTime.Format(timeLayout)

	// Ensure booking doesn't cross midnight or exceed opening hour constraints if any (we assume 24h for simplicity, but cross-midnight is restricted)
	if jamSelesai <= req.JamMulai && req.Durasi > 0 {
		respondWithError(w, http.StatusBadRequest, "Durasi booking melebihi batas hari (maksimal sebelum jam 24:00)")
		return
	}

	// 4. Check for schedule conflicts (Only with Approved bookings)
	conflict, err := repository.HasBookingConflict(req.FieldID, req.Tanggal, req.JamMulai, jamSelesai, 0)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Gagal memverifikasi jadwal bentrok")
		return
	}

	if conflict {
		respondWithError(w, http.StatusConflict, "Jadwal bentrok dengan booking yang sudah disetujui (Approved)!")
		return
	}

	// 5. Calculate pricing
	totalHarga := field.Harga * float64(req.Durasi)

	// 6. Create booking entry (status starts at 'pending')
	booking := models.Booking{
		UserID:     userID,
		FieldID:    req.FieldID,
		Tanggal:    parsedDate.Format("2006-01-02"),
		JamMulai:   req.JamMulai,
		JamSelesai: jamSelesai,
		TotalHarga: totalHarga,
		Status:     "pending",
		Catatan:    req.Catatan,
	}

	if err := repository.CreateBooking(&booking); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Gagal menyimpan data booking")
		return
	}

	// Load detail with joined fields for return
	newlyCreated, err := repository.GetBookingByID(booking.ID)
	if err != nil {
		respondWithJSON(w, http.StatusCreated, booking)
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Booking berhasil diajukan! Menunggu persetujuan admin.",
		"booking": newlyCreated,
	})
}

// UpdateBookingStatus updates status of a booking (Admin only)
func UpdateBookingStatus(w http.ResponseWriter, r *http.Request, bookingIDStr string) {
	bookingID, err := strconv.ParseInt(bookingIDStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID booking tidak valid")
		return
	}

	// Verify booking exists
	booking, err := repository.GetBookingByID(bookingID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Booking tidak ditemukan")
		return
	}

	var req models.UpdateBookingStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Format request tidak valid")
		return
	}

	status := req.Status
	if status != "approved" && status != "rejected" && status != "finished" && status != "pending" {
		respondWithError(w, http.StatusBadRequest, "Status tidak valid. Harus pending, approved, rejected, atau finished")
		return
	}

	// Business Rule: Booking with status Rejected cannot be changed back to Approved (must create a new booking)
	if booking.Status == "rejected" && status == "approved" {
		respondWithError(w, http.StatusBadRequest, "Booking yang sudah Ditolak (Rejected) tidak dapat disetujui kembali. Silakan buat booking baru.")
		return
	}

	// If transitioning to Approved, double check for schedule conflicts
	if status == "approved" {
		conflict, err := repository.HasBookingConflict(booking.FieldID, booking.Tanggal, booking.JamMulai, booking.JamSelesai, booking.ID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Gagal memvalidasi bentrok jadwal")
			return
		}
		if conflict {
			respondWithError(w, http.StatusConflict, "Jadwal bentrok dengan booking Approved lainnya! Tolak booking ini atau sesuaikan jadwal.")
			return
		}
	}

	if err := repository.UpdateBookingStatus(bookingID, status); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Gagal memperbarui status booking")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Status booking berhasil diperbarui menjadi " + status,
	})
}

// ListBookings returns bookings depending on user role
func ListBookings(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value(middleware.UserIDKey)
	roleVal := r.Context().Value(middleware.RoleKey)

	if userIDVal == nil || roleVal == nil {
		respondWithError(w, http.StatusUnauthorized, "Sesi tidak ditemukan")
		return
	}

	userID := userIDVal.(int64)
	role := roleVal.(string)

	bookings, err := repository.ListBookings(userID, role)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Gagal mengambil daftar booking")
		return
	}

	respondWithJSON(w, http.StatusOK, bookings)
}

// GetBooking detail for a specific booking
func GetBooking(w http.ResponseWriter, r *http.Request, bookingIDStr string) {
	bookingID, err := strconv.ParseInt(bookingIDStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID booking tidak valid")
		return
	}

	booking, err := repository.GetBookingByID(bookingID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Booking tidak ditemukan")
		return
	}

	// Auth check: Regular user can only view their own bookings
	userIDVal := r.Context().Value(middleware.UserIDKey)
	roleVal := r.Context().Value(middleware.RoleKey)

	if userIDVal == nil || roleVal == nil {
		respondWithError(w, http.StatusUnauthorized, "Sesi tidak ditemukan")
		return
	}

	userID := userIDVal.(int64)
	role := roleVal.(string)

	if role == "user" && booking.UserID != userID {
		respondWithError(w, http.StatusForbidden, "Anda tidak memiliki hak akses untuk melihat booking ini")
		return
	}

	respondWithJSON(w, http.StatusOK, booking)
}

// CancelBooking deletes a pending booking (User only can delete pending)
func CancelBooking(w http.ResponseWriter, r *http.Request, bookingIDStr string) {
	bookingID, err := strconv.ParseInt(bookingIDStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID booking tidak valid")
		return
	}

	booking, err := repository.GetBookingByID(bookingID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Booking tidak ditemukan")
		return
	}

	// Auth check: Check ownership
	userIDVal := r.Context().Value(middleware.UserIDKey)
	roleVal := r.Context().Value(middleware.RoleKey)

	if userIDVal == nil || roleVal == nil {
		respondWithError(w, http.StatusUnauthorized, "Sesi tidak ditemukan")
		return
	}

	userID := userIDVal.(int64)
	role := roleVal.(string)

	if role == "user" && booking.UserID != userID {
		respondWithError(w, http.StatusForbidden, "Anda tidak memiliki hak untuk membatalkan booking ini")
		return
	}

	// Business Rule: Users can only cancel if booking status is 'pending'
	if role == "user" && booking.Status != "pending" {
		respondWithError(w, http.StatusBadRequest, "Hanya booking berstatus Pending yang dapat dibatalkan")
		return
	}

	// Admin can cancel anything (delete booking from record)
	err = repository.DeleteBooking(bookingID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Gagal membatalkan booking")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Booking berhasil dibatalkan dan dihapus",
	})
}
