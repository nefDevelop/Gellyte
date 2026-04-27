package utils

import "strings"

// NormalizeID elimina los guiones de un ID (UUID) para asegurar la compatibilidad
// con los IDs almacenados en la base de datos (MD5 hex sin guiones).
func NormalizeID(id string) string {
	return strings.ReplaceAll(id, "-", "")
}
