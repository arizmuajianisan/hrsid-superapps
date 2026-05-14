package data

import "database/sql"

// Models membungkus semua repository kita
type Models struct {
	Users UserModel
	// Kedepannya tambahkan: Applications ApplicationModel
	Applications ApplicationModel // Tambahkan ini
}

func NewModels(db *sql.DB) Models {
	return Models{
		Users:        UserModel{DB: db},
		Applications: ApplicationModel{DB: db}, // Dan ini
	}
}
