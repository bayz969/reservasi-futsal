package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strings"

	"futsalbook/config"
	"futsalbook/handlers"
	"futsalbook/middleware"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func main() {
	// 1. Load environment variables from .env
	loadEnv()

	// 2. Initialize Database (MySQL or SQLite fallback)
	db := config.InitDB()
	defer db.Close()

	// 3. Setup Router
	r := chi.NewRouter()

	// Standard Middlewares
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	// API Endpoints Group
	r.Route("/api/v1", func(r chi.Router) {
		// Public Autentikasi
		r.Post("/auth/register", handlers.Register)
		r.Post("/auth/login", handlers.Login)
		r.Post("/auth/forgot-password", handlers.ForgotPassword)
		r.Post("/auth/reset-password", handlers.ResetPassword)
		r.Post("/auth/logout", handlers.Logout)

		// Protected Routes (User & Admin)
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware)

			// User & Profil
			r.Get("/users/me", handlers.GetMe)
			r.Put("/users/me", handlers.UpdateMe)

			// Lapangan (Fields) - viewable by logged-in users
			r.Get("/fields", handlers.ListFields)
			r.Get("/fields/{id}", func(w http.ResponseWriter, req *http.Request) {
				handlers.GetField(w, req, chi.URLParam(req, "id"))
			})

			// Booking Lapangan
			r.Post("/bookings", handlers.CreateBooking)
			r.Get("/bookings", handlers.ListBookings)
			r.Get("/bookings/{id}", func(w http.ResponseWriter, req *http.Request) {
				handlers.GetBooking(w, req, chi.URLParam(req, "id"))
			})
			r.Delete("/bookings/{id}", func(w http.ResponseWriter, req *http.Request) {
				handlers.CancelBooking(w, req, chi.URLParam(req, "id"))
			})

			// Dashboard User
			r.Get("/dashboard/user", handlers.GetUserDashboard)

			// Admin-Only Routes Group
			r.Group(func(r chi.Router) {
				r.Use(middleware.AdminMiddleware)

				// Kelola User (Admin)
				r.Get("/users", handlers.ListUsers)
				r.Put("/users/{id}/status", func(w http.ResponseWriter, req *http.Request) {
					handlers.ToggleUserStatus(w, req, chi.URLParam(req, "id"))
				})

				// Kelola Lapangan (Admin CRUD)
				r.Post("/fields", handlers.CreateField)
				r.Put("/fields/{id}", func(w http.ResponseWriter, req *http.Request) {
					handlers.UpdateField(w, req, chi.URLParam(req, "id"))
				})
				r.Delete("/fields/{id}", func(w http.ResponseWriter, req *http.Request) {
					handlers.DeleteField(w, req, chi.URLParam(req, "id"))
				})

				// Kelola Status Booking (Admin Approval)
				r.Put("/bookings/{id}/status", func(w http.ResponseWriter, req *http.Request) {
					handlers.UpdateBookingStatus(w, req, chi.URLParam(req, "id"))
				})

				// Dashboard Admin & Laporan
				r.Get("/dashboard/admin", handlers.GetAdminDashboard)
				r.Get("/reports/bookings", handlers.GetBookingReports)
				r.Get("/reports/popular-fields", handlers.GetPopularFieldsReport)
			})
		})
	})

	// 4. Serve Static Frontend Files
	// Check if public folder exists, if not create a basic placeholder
	if _, err := os.Stat("./public"); os.IsNotExist(err) {
		_ = os.Mkdir("./public", 0755)
	}

	fs := http.FileServer(http.Dir("./public"))
	r.Handle("/*", fs)

	// 5. Run Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server FutsalBook berjalan di http://localhost:%s", port)
	err := http.ListenAndServe(":"+port, r)
	if err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}

// loadEnv is a simple custom .env file parser to load environment variables
func loadEnv() {
	file, err := os.Open(".env")
	if err != nil {
		return // Ignore if .env doesn't exist
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])

			// Remove surrounding quotes if present
			if (strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"")) ||
				(strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'")) {
				val = val[1 : len(val)-1]
			}

			_ = os.Setenv(key, val)
		}
	}
}
