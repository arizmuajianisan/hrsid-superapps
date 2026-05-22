package data

import "database/sql"

// Models membungkus semua repository kita
type Models struct {
	Users        UserModel
	Applications ApplicationModel
	Sessions     SessionModel
	AuditLogs    AuditLogModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Users:        UserModel{DB: db},
		Applications: ApplicationModel{DB: db},
		Sessions:     SessionModel{DB: db},
		AuditLogs:    AuditLogModel{DB: db},
	}
}
