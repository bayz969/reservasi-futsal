package repository

import (
	"database/sql"
	"errors"
	"time"

	"futsalbook/config"
	"futsalbook/models"
)

// === USER REPOSITORY ===

func CreateUser(user *models.User) error {
	query := `INSERT INTO users (nama, email, password, nomor_hp, role, status) VALUES (?, ?, ?, ?, ?, ?)`
	res, err := config.DB.Exec(query, user.Nama, user.Email, user.Password, user.NomorHP, user.Role, user.Status)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		user.ID = id
	}
	return nil
}

func GetUserByEmail(email string) (models.User, error) {
	var user models.User
	query := `SELECT id, nama, email, password, nomor_hp, role, status, created_at, updated_at FROM users WHERE email = ?`
	err := config.DB.QueryRow(query, email).Scan(
		&user.ID, &user.Nama, &user.Email, &user.Password, &user.NomorHP, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return user, err
	}
	return user, nil
}

func GetUserByID(id int64) (models.User, error) {
	var user models.User
	query := `SELECT id, nama, email, password, nomor_hp, role, status, created_at, updated_at FROM users WHERE id = ?`
	err := config.DB.QueryRow(query, id).Scan(
		&user.ID, &user.Nama, &user.Email, &user.Password, &user.NomorHP, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return user, err
	}
	return user, nil
}

func UpdateUser(user *models.User) error {
	var query string
	var err error
	if user.Password != "" {
		query = `UPDATE users SET nama = ?, nomor_hp = ?, password = ?, updated_at = ? WHERE id = ?`
		_, err = config.DB.Exec(query, user.Nama, user.NomorHP, user.Password, time.Now(), user.ID)
	} else {
		query = `UPDATE users SET nama = ?, nomor_hp = ?, updated_at = ? WHERE id = ?`
		_, err = config.DB.Exec(query, user.Nama, user.NomorHP, time.Now(), user.ID)
	}
	return err
}

func ListUsers() ([]models.User, error) {
	users := []models.User{}
	query := `SELECT id, nama, email, nomor_hp, role, status, created_at, updated_at FROM users ORDER BY created_at DESC`
	rows, err := config.DB.Query(query)
	if err != nil {
		return users, err
	}
	defer rows.Close()

	for rows.Next() {
		var u models.User
		err := rows.Scan(&u.ID, &u.Nama, &u.Email, &u.NomorHP, &u.Role, &u.Status, &u.CreatedAt, &u.UpdatedAt)
		if err != nil {
			return users, err
		}
		users = append(users, u)
	}
	return users, nil
}

func UpdateUserStatus(id int64, status string) error {
	query := `UPDATE users SET status = ?, updated_at = ? WHERE id = ?`
	_, err := config.DB.Exec(query, status, time.Now(), id)
	return err
}

// === FIELD REPOSITORY ===

func CreateField(field *models.Field) error {
	query := `INSERT INTO fields (nama_lapangan, lokasi, harga, fasilitas, deskripsi, foto, status) VALUES (?, ?, ?, ?, ?, ?, ?)`
	res, err := config.DB.Exec(query, field.NamaLapangan, field.Lokasi, field.Harga, field.Fasilitas, field.Deskripsi, field.Foto, field.Status)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		field.ID = id
	}
	return nil
}

func UpdateField(field *models.Field) error {
	query := `UPDATE fields SET nama_lapangan = ?, lokasi = ?, harga = ?, fasilitas = ?, deskripsi = ?, foto = ?, status = ?, updated_at = ? WHERE id = ?`
	_, err := config.DB.Exec(query, field.NamaLapangan, field.Lokasi, field.Harga, field.Fasilitas, field.Deskripsi, field.Foto, field.Status, time.Now(), field.ID)
	return err
}

func DeleteField(id int64) error {
	query := `DELETE FROM fields WHERE id = ?`
	_, err := config.DB.Exec(query, id)
	return err
}

func ListFields(search, lokasi string, minHarga, maxHarga float64, status string) ([]models.Field, error) {
	fields := []models.Field{}
	query := `SELECT id, nama_lapangan, lokasi, harga, fasilitas, deskripsi, foto, status, created_at, updated_at FROM fields WHERE 1=1`
	args := []interface{}{}

	if search != "" {
		query += " AND (nama_lapangan LIKE ? OR deskripsi LIKE ?)"
		pattern := "%" + search + "%"
		args = append(args, pattern, pattern)
	}

	if lokasi != "" {
		query += " AND lokasi LIKE ?"
		args = append(args, "%"+lokasi+"%")
	}

	if minHarga > 0 {
		query += " AND harga >= ?"
		args = append(args, minHarga)
	}

	if maxHarga > 0 {
		query += " AND harga <= ?"
		args = append(args, maxHarga)
	}

	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}

	query += " ORDER BY id DESC"

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		return fields, err
	}
	defer rows.Close()

	for rows.Next() {
		var f models.Field
		err := rows.Scan(&f.ID, &f.NamaLapangan, &f.Lokasi, &f.Harga, &f.Fasilitas, &f.Deskripsi, &f.Foto, &f.Status, &f.CreatedAt, &f.UpdatedAt)
		if err != nil {
			return fields, err
		}
		fields = append(fields, f)
	}
	return fields, nil
}

func GetFieldByID(id int64) (models.Field, error) {
	var f models.Field
	query := `SELECT id, nama_lapangan, lokasi, harga, fasilitas, deskripsi, foto, status, created_at, updated_at FROM fields WHERE id = ?`
	err := config.DB.QueryRow(query, id).Scan(
		&f.ID, &f.NamaLapangan, &f.Lokasi, &f.Harga, &f.Fasilitas, &f.Deskripsi, &f.Foto, &f.Status, &f.CreatedAt, &f.UpdatedAt,
	)
	if err != nil {
		return f, err
	}
	return f, nil
}

func HasActiveBookings(fieldID int64) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM bookings WHERE field_id = ? AND status IN ('pending', 'approved')`
	err := config.DB.QueryRow(query, fieldID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// === BOOKING REPOSITORY ===

func CreateBooking(booking *models.Booking) error {
	query := `INSERT INTO bookings (user_id, field_id, tanggal, jam_mulai, jam_selesai, total_harga, status, catatan) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := config.DB.Exec(query, booking.UserID, booking.FieldID, booking.Tanggal, booking.JamMulai, booking.JamSelesai, booking.TotalHarga, booking.Status, booking.Catatan)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		booking.ID = id
	}
	return nil
}

func UpdateBookingStatus(id int64, status string) error {
	query := `UPDATE bookings SET status = ?, updated_at = ? WHERE id = ?`
	_, err := config.DB.Exec(query, status, time.Now(), id)
	return err
}

func GetBookingByID(id int64) (models.Booking, error) {
	var b models.Booking
	query := `
		SELECT b.id, b.user_id, b.field_id, b.tanggal, b.jam_mulai, b.jam_selesai, b.total_harga, b.status, b.catatan, b.created_at, b.updated_at,
		       u.nama as user_nama, u.email as user_email, u.nomor_hp as user_nomor_hp,
		       f.nama_lapangan as field_nama, f.lokasi as field_lokasi, f.foto as field_foto, f.harga as field_harga
		FROM bookings b
		JOIN users u ON b.user_id = u.id
		JOIN fields f ON b.field_id = f.id
		WHERE b.id = ?`

	err := config.DB.QueryRow(query, id).Scan(
		&b.ID, &b.UserID, &b.FieldID, &b.Tanggal, &b.JamMulai, &b.JamSelesai, &b.TotalHarga, &b.Status, &b.Catatan, &b.CreatedAt, &b.UpdatedAt,
		&b.UserNama, &b.UserEmail, &b.UserNomorHP,
		&b.FieldNama, &b.FieldLokasi, &b.FieldFoto, &b.FieldHarga,
	)
	return b, err
}

func ListBookings(userID int64, role string) ([]models.Booking, error) {
	bookings := []models.Booking{}
	var query string
	var args []interface{}

	query = `
		SELECT b.id, b.user_id, b.field_id, b.tanggal, b.jam_mulai, b.jam_selesai, b.total_harga, b.status, b.catatan, b.created_at, b.updated_at,
		       u.nama as user_nama, u.email as user_email, u.nomor_hp as user_nomor_hp,
		       f.nama_lapangan as field_nama, f.lokasi as field_lokasi, f.foto as field_foto, f.harga as field_harga
		FROM bookings b
		JOIN users u ON b.user_id = u.id
		JOIN fields f ON b.field_id = f.id`

	if role == "user" {
		query += " WHERE b.user_id = ? ORDER BY b.tanggal DESC, b.jam_mulai DESC"
		args = append(args, userID)
	} else {
		query += " ORDER BY b.created_at DESC"
	}

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		return bookings, err
	}
	defer rows.Close()

	for rows.Next() {
		var b models.Booking
		err := rows.Scan(
			&b.ID, &b.UserID, &b.FieldID, &b.Tanggal, &b.JamMulai, &b.JamSelesai, &b.TotalHarga, &b.Status, &b.Catatan, &b.CreatedAt, &b.UpdatedAt,
			&b.UserNama, &b.UserEmail, &b.UserNomorHP,
			&b.FieldNama, &b.FieldLokasi, &b.FieldFoto, &b.FieldHarga,
		)
		if err != nil {
			return bookings, err
		}
		bookings = append(bookings, b)
	}
	return bookings, nil
}

// HasBookingConflict checks if there's any approved booking overlapping the requested slot.
// Time range overlaps if: start1 < end2 AND end1 > start2.
func HasBookingConflict(fieldID int64, tanggal, jamMulai, jamSelesai string, excludeBookingID int64) (bool, error) {
	var count int
	var query string
	var err error

	if excludeBookingID > 0 {
		query = `
			SELECT COUNT(*) FROM bookings
			WHERE field_id = ? AND tanggal = ? AND status = 'approved'
			  AND jam_mulai < ? AND jam_selesai > ?
			  AND id != ?`
		err = config.DB.QueryRow(query, fieldID, tanggal, jamSelesai, jamMulai, excludeBookingID).Scan(&count)
	} else {
		query = `
			SELECT COUNT(*) FROM bookings
			WHERE field_id = ? AND tanggal = ? AND status = 'approved'
			  AND jam_mulai < ? AND jam_selesai > ?`
		err = config.DB.QueryRow(query, fieldID, tanggal, jamSelesai, jamMulai).Scan(&count)
	}

	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// === REPORT & DASHBOARD REPOSITORY ===

func GetUserDashboardStats(userID int64) (models.DashboardUserStats, error) {
	var stats models.DashboardUserStats
	stats.RecentBookings = []models.Booking{}

	// 1. Get Active Bookings count (pending or approved)
	queryCount := `SELECT COUNT(*) FROM bookings WHERE user_id = ? AND status IN ('pending', 'approved')`
	err := config.DB.QueryRow(queryCount, userID).Scan(&stats.ActiveBookingsCount)
	if err != nil {
		return stats, err
	}

	// 2. Get Next Booking details (approved, date >= today, ordered closest to now)
	todayStr := time.Now().Format("2006-01-02")
	queryNext := `
		SELECT b.tanggal, b.jam_mulai, f.nama_lapangan
		FROM bookings b
		JOIN fields f ON b.field_id = f.id
		WHERE b.user_id = ? AND b.status = 'approved' AND b.tanggal >= ?
		ORDER BY b.tanggal ASC, b.jam_mulai ASC
		LIMIT 1`

	err = config.DB.QueryRow(queryNext, userID, todayStr).Scan(&stats.NextBookingDate, &stats.NextBookingTime, &stats.NextBookingField)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			stats.NextBookingDate = "-"
			stats.NextBookingTime = "-"
			stats.NextBookingField = "Tidak ada jadwal terdekat"
		} else {
			return stats, err
		}
	}

	// 3. Get Recent Bookings (top 5)
	stats.RecentBookings, err = ListBookings(userID, "user")
	if err != nil {
		return stats, err
	}
	if len(stats.RecentBookings) > 5 {
		stats.RecentBookings = stats.RecentBookings[:5]
	}

	return stats, nil
}

func GetAdminDashboardStats() (models.DashboardAdminStats, error) {
	var stats models.DashboardAdminStats
	todayStr := time.Now().Format("2006-01-02")

	// Total Users
	err := config.DB.QueryRow(`SELECT COUNT(*) FROM users WHERE role = 'user'`).Scan(&stats.TotalUsers)
	if err != nil {
		return stats, err
	}

	// Total Fields
	err = config.DB.QueryRow(`SELECT COUNT(*) FROM fields`).Scan(&stats.TotalFields)
	if err != nil {
		return stats, err
	}

	// Total Bookings
	err = config.DB.QueryRow(`SELECT COUNT(*) FROM bookings`).Scan(&stats.TotalBookings)
	if err != nil {
		return stats, err
	}

	// Pending Bookings
	err = config.DB.QueryRow(`SELECT COUNT(*) FROM bookings WHERE status = 'pending'`).Scan(&stats.PendingBookings)
	if err != nil {
		return stats, err
	}

	// Bookings Today
	err = config.DB.QueryRow(`SELECT COUNT(*) FROM bookings WHERE tanggal = ?`, todayStr).Scan(&stats.BookingsToday)
	if err != nil {
		return stats, err
	}

	return stats, nil
}

func GetMonthlyReports() ([]models.MonthlyReport, error) {
	reports := []models.MonthlyReport{}
	var query string

	if config.DBType == "mysql" {
		query = `
			SELECT DATE_FORMAT(tanggal, '%Y-%m') as month, COUNT(*) as count
			FROM bookings
			WHERE status IN ('approved', 'finished')
			GROUP BY month
			ORDER BY month ASC`
	} else {
		query = `
			SELECT strftime('%Y-%m', tanggal) as month, COUNT(*) as count
			FROM bookings
			WHERE status IN ('approved', 'finished')
			GROUP BY month
			ORDER BY month ASC`
	}

	rows, err := config.DB.Query(query)
	if err != nil {
		return reports, err
	}
	defer rows.Close()

	for rows.Next() {
		var r models.MonthlyReport
		err := rows.Scan(&r.Month, &r.Count)
		if err != nil {
			return reports, err
		}
		reports = append(reports, r)
	}

	return reports, nil
}

func GetPopularFieldsReport() ([]models.PopularFieldReport, error) {
	reports := []models.PopularFieldReport{}
	query := `
		SELECT f.nama_lapangan, COUNT(b.id) as count
		FROM bookings b
		JOIN fields f ON b.field_id = f.id
		WHERE b.status IN ('approved', 'finished')
		GROUP BY f.id, f.nama_lapangan
		ORDER BY count DESC
		LIMIT 5`

	rows, err := config.DB.Query(query)
	if err != nil {
		return reports, err
	}
	defer rows.Close()

	for rows.Next() {
		var r models.PopularFieldReport
		err := rows.Scan(&r.FieldName, &r.Count)
		if err != nil {
			return reports, err
		}
		reports = append(reports, r)
	}

	return reports, nil
}

func GetTotalRevenue() (float64, error) {
	var total float64
	query := `SELECT COALESCE(SUM(total_harga), 0) FROM bookings WHERE status IN ('approved', 'finished')`
	err := config.DB.QueryRow(query).Scan(&total)
	return total, err
}

func DeleteBooking(id int64) error {
	query := `DELETE FROM bookings WHERE id = ?`
	_, err := config.DB.Exec(query, id)
	return err
}

