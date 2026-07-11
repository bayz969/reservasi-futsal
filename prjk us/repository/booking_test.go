package repository

import (
	"database/sql"
	"testing"

	"futsalbook/config"
	"futsalbook/models"

	_ "github.com/glebarez/go-sqlite"
)

func TestHasBookingConflict(t *testing.T) {
	// 1. Initialize temporary in-memory database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory db: %v", err)
	}
	defer db.Close()

	config.DB = db
	config.DBType = "sqlite"

	// 2. Run migrations
	migrateQuery := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nama TEXT, email TEXT UNIQUE, password TEXT, nomor_hp TEXT, role TEXT, status TEXT
	);
	CREATE TABLE IF NOT EXISTS fields (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nama_lapangan TEXT, lokasi TEXT, harga REAL, fasilitas TEXT, deskripsi TEXT, foto TEXT, status TEXT
	);
	CREATE TABLE IF NOT EXISTS bookings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER, field_id INTEGER, tanggal TEXT, jam_mulai TEXT, jam_selesai TEXT, total_harga REAL, status TEXT, catatan TEXT
	);`

	_, err = db.Exec(migrateQuery)
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// 3. Seed mock user and field
	_, err = db.Exec(`INSERT INTO users (id, nama, email, password, nomor_hp, role, status) VALUES (1, 'Test User', 'test@user.com', 'pwd', '123', 'user', 'active')`)
	if err != nil {
		t.Fatalf("Failed to seed user: %v", err)
	}

	_, err = db.Exec(`INSERT INTO fields (id, nama_lapangan, lokasi, harga, fasilitas, deskripsi, foto, status) VALUES (1, 'Lapangan Utama', 'Sektor A', 100000, 'Fasilitas', 'Deskripsi', 'Foto', 'tersedia')`)
	if err != nil {
		t.Fatalf("Failed to seed field: %v", err)
	}

	// 4. Seed an APPROVED booking from 10:00 to 12:00
	approvedBooking := models.Booking{
		UserID:     1,
		FieldID:    1,
		Tanggal:    "2026-07-11",
		JamMulai:   "10:00",
		JamSelesai: "12:00",
		TotalHarga: 200000,
		Status:     "approved",
	}
	err = CreateBooking(&approvedBooking)
	if err != nil {
		t.Fatalf("Failed to create approved booking: %v", err)
	}

	// 5. Seed a PENDING booking from 14:00 to 16:00 (should NOT block since it's not approved)
	pendingBooking := models.Booking{
		UserID:     1,
		FieldID:    1,
		Tanggal:    "2026-07-11",
		JamMulai:   "14:00",
		JamSelesai: "16:00",
		TotalHarga: 200000,
		Status:     "pending",
	}
	err = CreateBooking(&pendingBooking)
	if err != nil {
		t.Fatalf("Failed to create pending booking: %v", err)
	}

	// Define test cases
	tests := []struct {
		name       string
		fieldID    int64
		tanggal    string
		jamMulai   string
		jamSelesai string
		excludeID  int64
		want       bool
	}{
		{
			name:       "Overlaps - starts before, ends inside (09:00 - 11:00)",
			fieldID:    1,
			tanggal:    "2026-07-11",
			jamMulai:   "09:00",
			jamSelesai: "11:00",
			excludeID:  0,
			want:       true,
		},
		{
			name:       "Overlaps - starts inside, ends after (11:00 - 13:00)",
			fieldID:    1,
			tanggal:    "2026-07-11",
			jamMulai:   "11:00",
			jamSelesai: "13:00",
			excludeID:  0,
			want:       true,
		},
		{
			name:       "Overlaps - fully inside (10:30 - 11:30)",
			fieldID:    1,
			tanggal:    "2026-07-11",
			jamMulai:   "10:30",
			jamSelesai: "11:30",
			excludeID:  0,
			want:       true,
		},
		{
			name:       "Does not overlap - ends exactly when booking starts (08:00 - 10:00)",
			fieldID:    1,
			tanggal:    "2026-07-11",
			jamMulai:   "08:00",
			jamSelesai: "10:00",
			excludeID:  0,
			want:       false,
		},
		{
			name:       "Does not overlap - starts exactly when booking ends (12:00 - 14:00)",
			fieldID:    1,
			tanggal:    "2026-07-11",
			jamMulai:   "12:00",
			jamSelesai: "14:00",
			excludeID:  0,
			want:       false,
		},
		{
			name:       "Does not overlap - completely different time (17:00 - 18:00)",
			fieldID:    1,
			tanggal:    "2026-07-11",
			jamMulai:   "17:00",
			jamSelesai: "18:00",
			excludeID:  0,
			want:       false,
		},
		{
			name:       "Does not overlap - overlaps with PENDING booking (14:30 - 15:30)",
			fieldID:    1,
			tanggal:    "2026-07-11",
			jamMulai:   "14:30",
			jamSelesai: "15:30",
			excludeID:  0,
			want:       false,
		},
		{
			name:       "Exclude ID - matches approved booking but excludes it (10:30 - 11:30, exclude approved)",
			fieldID:    1,
			tanggal:    "2026-07-11",
			jamMulai:   "10:30",
			jamSelesai: "11:30",
			excludeID:  approvedBooking.ID,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HasBookingConflict(tt.fieldID, tt.tanggal, tt.jamMulai, tt.jamSelesai, tt.excludeID)
			if err != nil {
				t.Fatalf("Unexpected error running check: %v", err)
			}
			if got != tt.want {
				t.Errorf("HasBookingConflict() = %v, want %v", got, tt.want)
			}
		})
	}
}
