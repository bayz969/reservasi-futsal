package handlers

import (
	"net/http"

	"futsalbook/middleware"
	"futsalbook/repository"
)

// GetUserDashboard gets statistics and recent bookings for the logged-in user
func GetUserDashboard(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value(middleware.UserIDKey)
	if userIDVal == nil {
		respondWithError(w, http.StatusUnauthorized, "Sesi tidak ditemukan")
		return
	}
	userID := userIDVal.(int64)

	stats, err := repository.GetUserDashboardStats(userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Gagal mengambil data dashboard user")
		return
	}

	respondWithJSON(w, http.StatusOK, stats)
}

// GetAdminDashboard gets overall business metrics for the admin dashboard
func GetAdminDashboard(w http.ResponseWriter, r *http.Request) {
	stats, err := repository.GetAdminDashboardStats()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Gagal mengambil data dashboard admin")
		return
	}

	respondWithJSON(w, http.StatusOK, stats)
}

// GetBookingReports generates reports of total bookings, revenue, and monthly breakdown (Admin only)
func GetBookingReports(w http.ResponseWriter, r *http.Request) {
	monthly, err := repository.GetMonthlyReports()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Gagal mengambil data laporan bulanan")
		return
	}

	revenue, err := repository.GetTotalRevenue()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Gagal mengambil data laporan pendapatan")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"monthly_reports": monthly,
		"total_revenue":   revenue,
	})
}

// GetPopularFieldsReport generates stats on most popular fields based on booking count (Admin only)
func GetPopularFieldsReport(w http.ResponseWriter, r *http.Request) {
	popular, err := repository.GetPopularFieldsReport()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Gagal mengambil data laporan lapangan terpopuler")
		return
	}

	respondWithJSON(w, http.StatusOK, popular)
}
