package models

import "time"

// User represents the users table in database
type User struct {
	ID        int64     `json:"id"`
	Nama      string    `json:"nama"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Hidden in JSON response
	NomorHP   string    `json:"nomor_hp"`
	Role      string    `json:"role"`   // 'user' or 'admin'
	Status    string    `json:"status"` // 'active' or 'inactive'
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Field represents the fields table (lapangan)
type Field struct {
	ID           int64     `json:"id"`
	NamaLapangan string    `json:"nama_lapangan"`
	Lokasi       string    `json:"lokasi"`
	Harga        float64   `json:"harga"`
	Fasilitas    string    `json:"fasilitas"`
	Deskripsi    string    `json:"deskripsi"`
	Foto         string    `json:"foto"`
	Status       string    `json:"status"` // 'tersedia' or 'tidak_tersedia'
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Booking represents the bookings table
type Booking struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	FieldID    int64     `json:"field_id"`
	Tanggal    string    `json:"tanggal"` // YYYY-MM-DD
	JamMulai   string    `json:"jam_mulai"` // HH:MM
	JamSelesai string    `json:"jam_selesai"` // HH:MM
	TotalHarga float64   `json:"total_harga"`
	Status     string    `json:"status"` // 'pending', 'approved', 'rejected', 'finished'
	Catatan    string    `json:"catatan"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Relational fields for join queries
	UserNama     string  `json:"user_nama,omitempty"`
	UserEmail    string  `json:"user_email,omitempty"`
	UserNomorHP  string  `json:"user_nomor_hp,omitempty"`
	FieldNama    string  `json:"field_nama,omitempty"`
	FieldLokasi  string  `json:"field_lokasi,omitempty"`
	FieldFoto    string  `json:"field_foto,omitempty"`
	FieldHarga   float64 `json:"field_harga,omitempty"`
}

// API Request/Response Payloads

type RegisterRequest struct {
	Nama     string `json:"nama"`
	Email    string `json:"email"`
	NomorHP  string `json:"nomor_hp"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

type UpdateProfileRequest struct {
	Nama     string `json:"nama"`
	NomorHP  string `json:"nomor_hp"`
	Password string `json:"password,omitempty"` // Optional password update
}

type FieldRequest struct {
	NamaLapangan string  `json:"nama_lapangan"`
	Lokasi       string  `json:"lokasi"`
	Harga        float64 `json:"harga"`
	Fasilitas    string  `json:"fasilitas"`
	Deskripsi    string  `json:"deskripsi"`
	Foto         string  `json:"foto"`
	Status       string  `json:"status"`
}

type BookingRequest struct {
	FieldID  int64  `json:"field_id"`
	Tanggal  string `json:"tanggal"` // YYYY-MM-DD
	JamMulai string `json:"jam_mulai"` // HH:MM
	Durasi   int    `json:"durasi"` // In hours
	Catatan  string `json:"catatan"`
}

type UpdateBookingStatusRequest struct {
	Status string `json:"status"` // 'approved', 'rejected', 'finished'
}

type UserStatusRequest struct {
	Status string `json:"status"` // 'active', 'inactive'
}

type DashboardUserStats struct {
	ActiveBookingsCount int64     `json:"active_bookings_count"`
	NextBookingDate     string    `json:"next_booking_date"`
	NextBookingTime     string    `json:"next_booking_time"`
	NextBookingField    string    `json:"next_booking_field"`
	RecentBookings      []Booking `json:"recent_bookings"`
}

type DashboardAdminStats struct {
	TotalUsers      int64 `json:"total_users"`
	TotalFields     int64 `json:"total_fields"`
	TotalBookings   int64 `json:"total_bookings"`
	PendingBookings int64 `json:"pending_bookings"`
	BookingsToday   int64 `json:"bookings_today"`
}

type MonthlyReport struct {
	Month string `json:"month"` // e.g. "January" or "2026-07"
	Count int64  `json:"count"`
}

type PopularFieldReport struct {
	FieldName string `json:"field_name"`
	Count     int64  `json:"count"`
}

type AdminReportResponse struct {
	MonthlyReports []MonthlyReport       `json:"monthly_reports"`
	PopularFields  []PopularFieldReport  `json:"popular_fields"`
	TotalRevenue   float64               `json:"total_revenue"`
}
