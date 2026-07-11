package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"futsalbook/models"
	"futsalbook/repository"
)

// ListFields handles fetching list of fields with filters
// Query params: search, lokasi, harga_min, harga_max, status
func ListFields(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	lokasi := r.URL.Query().Get("lokasi")
	status := r.URL.Query().Get("status")

	var minHarga, maxHarga float64
	var err error

	if minStr := r.URL.Query().Get("harga_min"); minStr != "" {
		minHarga, err = strconv.ParseFloat(minStr, 64)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "harga_min tidak valid")
			return
		}
	}

	if maxStr := r.URL.Query().Get("harga_max"); maxStr != "" {
		maxHarga, err = strconv.ParseFloat(maxStr, 64)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "harga_max tidak valid")
			return
		}
	}

	fields, err := repository.ListFields(search, lokasi, minHarga, maxHarga, status)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Gagal mengambil daftar lapangan")
		return
	}

	respondWithJSON(w, http.StatusOK, fields)
}

// GetField handles fetching field detail by ID
func GetField(w http.ResponseWriter, r *http.Request, fieldIDStr string) {
	fieldID, err := strconv.ParseInt(fieldIDStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID lapangan tidak valid")
		return
	}

	field, err := repository.GetFieldByID(fieldID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Lapangan tidak ditemukan")
		return
	}

	respondWithJSON(w, http.StatusOK, field)
}

// CreateField handles adding new field (Admin only)
func CreateField(w http.ResponseWriter, r *http.Request) {
	var req models.FieldRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Format request tidak valid")
		return
	}

	if req.NamaLapangan == "" || req.Lokasi == "" || req.Harga <= 0 || req.Fasilitas == "" || req.Deskripsi == "" {
		respondWithError(w, http.StatusBadRequest, "Semua field wajib diisi dengan nilai yang valid")
		return
	}

	if req.Status == "" {
		req.Status = "tersedia"
	}

	if req.Foto == "" {
		// Fallback sample image
		req.Foto = "https://images.unsplash.com/photo-1575361204480-aadea25e6e68?w=800&auto=format&fit=crop&q=60"
	}

	field := models.Field{
		NamaLapangan: req.NamaLapangan,
		Lokasi:       req.Lokasi,
		Harga:        req.Harga,
		Fasilitas:    req.Fasilitas,
		Deskripsi:    req.Deskripsi,
		Foto:         req.Foto,
		Status:       req.Status,
	}

	if err := repository.CreateField(&field); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Gagal menambahkan lapangan baru")
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Lapangan berhasil ditambahkan",
		"field":   field,
	})
}

// UpdateField handles updating existing field (Admin only)
func UpdateField(w http.ResponseWriter, r *http.Request, fieldIDStr string) {
	fieldID, err := strconv.ParseInt(fieldIDStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID lapangan tidak valid")
		return
	}

	// Verify field exists
	field, err := repository.GetFieldByID(fieldID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Lapangan tidak ditemukan")
		return
	}

	var req models.FieldRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Format request tidak valid")
		return
	}

	if req.NamaLapangan == "" || req.Lokasi == "" || req.Harga <= 0 || req.Fasilitas == "" || req.Deskripsi == "" {
		respondWithError(w, http.StatusBadRequest, "Semua field wajib diisi dengan nilai yang valid")
		return
	}

	field.NamaLapangan = req.NamaLapangan
	field.Lokasi = req.Lokasi
	field.Harga = req.Harga
	field.Fasilitas = req.Fasilitas
	field.Deskripsi = req.Deskripsi
	if req.Foto != "" {
		field.Foto = req.Foto
	}
	if req.Status != "" {
		field.Status = req.Status
	}

	if err := repository.UpdateField(&field); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Gagal memperbarui data lapangan")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Lapangan berhasil diperbarui",
		"field":   field,
	})
}

// DeleteField handles removing a field (Admin only)
func DeleteField(w http.ResponseWriter, r *http.Request, fieldIDStr string) {
	fieldID, err := strconv.ParseInt(fieldIDStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "ID lapangan tidak valid")
		return
	}

	// Verify field exists
	_, err = repository.GetFieldByID(fieldID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Lapangan tidak ditemukan")
		return
	}

	// Business Rule: Admin cannot delete a field that has active bookings (Pending / Approved)
	hasActive, err := repository.HasActiveBookings(fieldID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Gagal memeriksa jadwal booking aktif")
		return
	}

	if hasActive {
		respondWithError(w, http.StatusConflict, "Tidak dapat menghapus lapangan karena masih memiliki jadwal booking aktif (Pending/Approved)")
		return
	}

	if err := repository.DeleteField(fieldID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Gagal menghapus data lapangan")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Lapangan berhasil dihapus",
	})
}
