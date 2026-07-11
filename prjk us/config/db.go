package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/glebarez/go-sqlite"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

var DB *sql.DB
var DBType string

// InitDB initializes the database connection and runs migrations
func InitDB() *sql.DB {
	dbType := os.Getenv("DB_TYPE")
	if dbType == "" {
		dbType = "sqlite" // Fallback to SQLite
	}
	dbType = strings.ToLower(dbType)
	DBType = dbType

	var db *sql.DB
	var err error

	if dbType == "mysql" {
		dbUser := os.Getenv("DB_USER")
		dbPass := os.Getenv("DB_PASS")
		dbHost := os.Getenv("DB_HOST")
		dbPort := os.Getenv("DB_PORT")
		dbName := os.Getenv("DB_NAME")

		if dbUser == "" || dbHost == "" || dbName == "" {
			log.Println("MySQL environment variables missing. Falling back to SQLite...")
			dbType = "sqlite"
			DBType = "sqlite"
		} else {
			dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPass, dbHost, dbPort, dbName)
			db, err = sql.Open("mysql", dsn)
			if err != nil {
				log.Fatalf("Error opening MySQL connection: %v", err)
			}
			err = db.Ping()
			if err != nil {
				log.Printf("MySQL ping failed: %v. Falling back to SQLite...", err)
				dbType = "sqlite"
				DBType = "sqlite"
			} else {
				log.Println("Connected to MySQL database successfully!")
			}
		}
	}

	if dbType == "sqlite" {
		dbFile := os.Getenv("DB_FILE")
		if dbFile == "" {
			dbFile = "futsalbook.db"
		}
		db, err = sql.Open("sqlite", dbFile)
		if err != nil {
			log.Fatalf("Error opening SQLite connection: %v", err)
		}
		log.Printf("Connected to SQLite database (%s) successfully!", dbFile)
	}

	DB = db
	migrate()
	seed()
	return DB
}

func migrate() {
	var userTableQuery, fieldTableQuery, bookingTableQuery string

	if DBType == "mysql" {
		userTableQuery = `
		CREATE TABLE IF NOT EXISTS users (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			nama VARCHAR(100) NOT NULL,
			email VARCHAR(150) NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL,
			nomor_hp VARCHAR(20) NOT NULL,
			role VARCHAR(20) NOT NULL DEFAULT 'user',
			status VARCHAR(20) NOT NULL DEFAULT 'active',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		);`

		fieldTableQuery = `
		CREATE TABLE IF NOT EXISTS fields (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			nama_lapangan VARCHAR(150) NOT NULL,
			lokasi VARCHAR(255) NOT NULL,
			harga DECIMAL(10,2) NOT NULL,
			fasilitas TEXT NOT NULL,
			deskripsi TEXT NOT NULL,
			foto VARCHAR(255) NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'tersedia',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		);`

		bookingTableQuery = `
		CREATE TABLE IF NOT EXISTS bookings (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT NOT NULL,
			field_id BIGINT NOT NULL,
			tanggal DATE NOT NULL,
			jam_mulai TIME NOT NULL,
			jam_selesai TIME NOT NULL,
			total_harga DECIMAL(10,2) NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			catatan TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (field_id) REFERENCES fields(id) ON DELETE CASCADE
		);`
	} else {
		// SQLite
		userTableQuery = `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			nama TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			nomor_hp TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'user',
			status TEXT NOT NULL DEFAULT 'active',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`

		fieldTableQuery = `
		CREATE TABLE IF NOT EXISTS fields (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			nama_lapangan TEXT NOT NULL,
			lokasi TEXT NOT NULL,
			harga REAL NOT NULL,
			fasilitas TEXT NOT NULL,
			deskripsi TEXT NOT NULL,
			foto TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'tersedia',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`

		bookingTableQuery = `
		CREATE TABLE IF NOT EXISTS bookings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			field_id INTEGER NOT NULL,
			tanggal TEXT NOT NULL,
			jam_mulai TEXT NOT NULL,
			jam_selesai TEXT NOT NULL,
			total_harga REAL NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			catatan TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (field_id) REFERENCES fields(id) ON DELETE CASCADE
		);`
	}

	_, err := DB.Exec(userTableQuery)
	if err != nil {
		log.Fatalf("Error creating users table: %v", err)
	}

	_, err = DB.Exec(fieldTableQuery)
	if err != nil {
		log.Fatalf("Error creating fields table: %v", err)
	}

	_, err = DB.Exec(bookingTableQuery)
	if err != nil {
		log.Fatalf("Error creating bookings table: %v", err)
	}

	log.Println("Database migrations applied successfully.")
}

func seed() {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		log.Printf("Error checking user count for seeding: %v", err)
		return
	}

	// Seed if empty
	if count == 0 {
		log.Println("Seeding initial data...")

		// Helper to hash password
		hashPassword := func(pw string) string {
			h, _ := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
			return string(h)
		}

		adminPW := hashPassword("admin123")
		userPW := hashPassword("user123")

		// Create Admin
		_, err = DB.Exec(`INSERT INTO users (nama, email, password, nomor_hp, role, status) VALUES (?, ?, ?, ?, ?, ?)`,
			"Admin FutsalBook", "admin@futsalbook.com", adminPW, "08123456789", "admin", "active")
		if err != nil {
			log.Printf("Error seeding admin: %v", err)
		} else {
			log.Println("Seeding Admin: admin@futsalbook.com / admin123")
		}

		// Create User
		_, err = DB.Exec(`INSERT INTO users (nama, email, password, nomor_hp, role, status) VALUES (?, ?, ?, ?, ?, ?)`,
			"Andi Wijaya", "andi@gmail.com", userPW, "08987654321", "user", "active")
		if err != nil {
			log.Printf("Error seeding user: %v", err)
		} else {
			log.Println("Seeding User: andi@gmail.com / user123")
		}

		// Seed some fields (lapangan)
		fields := []struct {
			Nama      string
			Lokasi    string
			Harga     float64
			Fasilitas string
			Deskripsi string
			Foto      string
		}{
			{
				Nama:      "Lapangan A (Sintetis Premium)",
				Lokasi:    "Gedung Olahraga Sudirman, Jakarta",
				Harga:     150000,
				Fasilitas: "Rumput Sintetis Premium, Gawang Standar FIFA, Ruang Ganti, Shower Panas, Kantin, WiFi Gratis",
				Deskripsi: "Lapangan futsal indoor dengan rumput sintetis berkualitas tinggi tipe monofilament, lembut dan meminimalisir risiko cedera lutut. Dilengkapi penerangan LED 400 Lux untuk kenyamanan bermain malam hari.",
				Foto:      "https://images.unsplash.com/photo-1575361204480-aadea25e6e68?w=800&auto=format&fit=crop&q=60",
			},
			{
				Nama:      "Lapangan B (Vinyl Interlock)",
				Lokasi:    "Gedung Olahraga Sudirman, Jakarta",
				Harga:     175000,
				Fasilitas: "Lantai Vinyl Interlock, Papan Skor Digital, Bench Pemain Cadangan, AC (Tribun), Parkir Luas",
				Deskripsi: "Lapangan futsal dengan lantai interlock berstandar kompetisi nasional. Memiliki daya cengkram sepatu yang sangat baik untuk kecepatan permainan tinggi. Sangat direkomendasikan untuk turnamen.",
				Foto:      "https://images.unsplash.com/photo-1517649763962-0c623066013b?w=800&auto=format&fit=crop&q=60",
			},
			{
				Nama:      "Lapangan C (Semi Outdoor Parquet)",
				Lokasi:    "Kuningan Sport Center, Jakarta",
				Harga:     120000,
				Fasilitas: "Lantai Parquet Kayu, Semi Outdoor, Kipas Angin Raksasa, Mushola, Kamar Mandi Bersih",
				Deskripsi: "Lapangan futsal semi outdoor dengan lantai kayu parket yang klasik. Udara segar mengalir bebas, menciptakan atmosfer bermain yang segar dan menyenangkan di sore hari.",
				Foto:      "https://images.unsplash.com/photo-1529900748604-07564a03e7a6?w=800&auto=format&fit=crop&q=60",
			},
		}

		for _, f := range fields {
			_, err = DB.Exec(`INSERT INTO fields (nama_lapangan, lokasi, harga, fasilitas, deskripsi, foto, status) VALUES (?, ?, ?, ?, ?, ?, ?)`,
				f.Nama, f.Lokasi, f.Harga, f.Fasilitas, f.Deskripsi, f.Foto, "tersedia")
			if err != nil {
				log.Printf("Error seeding field %s: %v", f.Nama, err)
			}
		}
		log.Println("Seeding sample fields completed.")
	}
}
