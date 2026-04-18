package models

// User representa un usuario en el sistema Gellyte compatible con Jellyfin
type User struct {
	ID       string `gorm:"primaryKey" json:"Id"`
	Username string `gorm:"uniqueIndex" json:"Name"`
	Password string `json:"-"`
	IsAdmin  bool   `json:"HasPassword"`
}

// GORM utiliza una tabla de base de datos llamada 'users' por defecto.
// Al usar string como ID, evitamos los incrementales de SQLite.
