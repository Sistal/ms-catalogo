package utils

import (
	"fmt"
	"math"
	"ms-catalogo/models"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Paginate aplica paginación a una query GORM
func Paginate(page, limit int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}
		if limit <= 0 {
			limit = 20
		}
		if limit > 100 {
			limit = 100
		}

		offset := (page - 1) * limit
		return db.Offset(offset).Limit(limit)
	}
}

// CalculatePaginationMeta calcula el metadata de paginación
func CalculatePaginationMeta(page, limit int, total int64) models.PaginationMeta {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return models.PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      int(total),
		TotalPages: totalPages,
	}
}

// ValidateTallas valida el formato de tallas (separadas por coma)
func ValidateTallas(tallas string) error {
	if strings.TrimSpace(tallas) == "" {
		return fmt.Errorf("debe especificar al menos una talla")
	}

	// Verificar formato válido (valores separados por coma)
	parts := strings.Split(tallas, ",")
	if len(parts) == 0 {
		return fmt.Errorf("formato de tallas inválido")
	}

	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			return fmt.Errorf("tallas no pueden contener valores vacíos")
		}
	}

	return nil
}

// ValidateDateRange valida que fecha_inicio < fecha_fin
func ValidateDateRange(fechaInicio, fechaFin string) error {
	layout := "2006-01-02"

	inicio, err := time.Parse(layout, fechaInicio)
	if err != nil {
		return fmt.Errorf("fecha_inicio inválida: debe ser formato ISO 8601 (YYYY-MM-DD)")
	}

	fin, err := time.Parse(layout, fechaFin)
	if err != nil {
		return fmt.Errorf("fecha_fin inválida: debe ser formato ISO 8601 (YYYY-MM-DD)")
	}

	if !fin.After(inicio) {
		return fmt.Errorf("fecha_fin debe ser posterior a fecha_inicio")
	}

	return nil
}

// ParseQueryInt convierte query parameter string a int con valor por defecto
func ParseQueryInt(value string, defaultValue int) int {
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}

// ParseQueryBool convierte query parameter string a bool
func ParseQueryBool(value string) *bool {
	if value == "" {
		return nil
	}

	boolValue := value == "true" || value == "1"
	return &boolValue
}

// CalculateDaysRemaining calcula días restantes desde hoy hasta fecha_fin
func CalculateDaysRemaining(fechaFin string) int {
	layout := "2006-01-02"
	fin, err := time.Parse(layout, fechaFin)
	if err != nil {
		return 0
	}

	now := time.Now()
	diff := fin.Sub(now)
	days := int(diff.Hours() / 24)

	if days < 0 {
		return 0
	}

	return days
}

// CalculateDaysElapsed calcula días transcurridos desde fecha_inicio
func CalculateDaysElapsed(fechaInicio string) int {
	layout := "2006-01-02"
	inicio, err := time.Parse(layout, fechaInicio)
	if err != nil {
		return 0
	}

	now := time.Now()
	diff := now.Sub(inicio)
	days := int(diff.Hours() / 24)

	if days < 0 {
		return 0
	}

	return days
}

// CalculatePercentageCompleted calcula porcentaje completado de temporada
func CalculatePercentageCompleted(fechaInicio, fechaFin string) float64 {
	layout := "2006-01-02"
	inicio, err := time.Parse(layout, fechaInicio)
	if err != nil {
		return 0
	}

	fin, err := time.Parse(layout, fechaFin)
	if err != nil {
		return 0
	}

	now := time.Now()

	totalDuration := fin.Sub(inicio).Hours()
	elapsedDuration := now.Sub(inicio).Hours()

	if totalDuration <= 0 {
		return 0
	}

	percentage := (elapsedDuration / totalDuration) * 100

	if percentage < 0 {
		return 0
	}
	if percentage > 100 {
		return 100
	}

	// Redondear a 2 decimales
	return math.Round(percentage*100) / 100
}

// NormalizeTallas normaliza el formato de tallas (trim espacios)
func NormalizeTallas(tallas string) string {
	parts := strings.Split(tallas, ",")
	normalized := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			normalized = append(normalized, trimmed)
		}
	}

	return strings.Join(normalized, ",")
}

// ContainsTalla verifica si una talla específica está en la lista
func ContainsTalla(tallasDisponibles string, tallasBuscadas []string) bool {
	if len(tallasBuscadas) == 0 {
		return true
	}

	tallas := strings.Split(tallasDisponibles, ",")
	tallasMap := make(map[string]bool)

	for _, talla := range tallas {
		tallasMap[strings.TrimSpace(talla)] = true
	}

	// Verificar si al menos una de las tallas buscadas está disponible
	for _, buscada := range tallasBuscadas {
		if tallasMap[strings.TrimSpace(buscada)] {
			return true
		}
	}

	return false
}

// GetTipoEmpresaNombre obtiene el nombre del tipo de empresa
func GetTipoEmpresaNombre(idTipo int) string {
	switch idTipo {
	case 1:
		return "Cliente"
	case 2:
		return "Proveedor"
	case 3:
		return "Ambos"
	default:
		return "Desconocido"
	}
}

// ValidateURL valida formato básico de URL
func ValidateURL(url string) bool {
	if url == "" {
		return true // URL opcional
	}
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

// BuildLikePattern construye patrón LIKE para búsqueda
func BuildLikePattern(search string) string {
	if search == "" {
		return ""
	}
	return "%" + strings.ToLower(search) + "%"
}
