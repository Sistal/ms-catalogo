package controllers

import (
	"ms-catalogo/initializers"
	"ms-catalogo/middleware"
	"ms-catalogo/models"
	"ms-catalogo/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ListTemporadas godoc
// @Summary Listar Temporadas
// @Description Obtiene lista de temporadas/campañas con paginación
// @Tags Temporadas
// @Accept json
// @Produce json
// @Param page query int false "Número de página" default(1)
// @Param limit query int false "Límite por página" default(20)
// @Param id_estado_temporada query int false "Filtrar por estado (40=Activa, 41=Inactiva, 42=Finalizada)"
// @Param activa_solo query bool false "Solo temporada activa"
// @Success 200 {object} models.StandardResponse
// @Router /api/v1/temporadas [get]
func ListTemporadas(c *gin.Context) {
	page := utils.ParseQueryInt(c.Query("page"), 1)
	limit := utils.ParseQueryInt(c.Query("limit"), 20)
	idEstado := utils.ParseQueryInt(c.Query("id_estado_temporada"), 0)
	activaSolo := utils.ParseQueryBool(c.Query("activa_solo"))

	query := initializers.DB.Model(&models.Temporada{})

	// Filtros
	if idEstado > 0 {
		query = query.Where("id_estado_temporada = ?", idEstado)
	}
	if activaSolo != nil && *activaSolo {
		query = query.Where("id_estado_temporada = ?", 40) // 40 = Activa
	}

	// Contar total
	var total int64
	query.Count(&total)

	// Obtener temporadas
	var temporadas []models.Temporada
	query.Preload("Estado").
		Scopes(utils.Paginate(page, limit)).
		Order("fecha_inicio DESC").
		Find(&temporadas)

	// Construir respuesta
	response := make([]models.TemporadaResponseDTO, 0)
	for _, t := range temporadas {
		// Contar prendas de esta temporada
		var totalPrendas int64
		initializers.DB.Model(&models.Prenda{}).Where("id_temporada = ?", t.IDTemporada).Count(&totalPrendas)

		diasRestantes := utils.CalculateDaysRemaining(t.FechaFin)

		response = append(response, models.TemporadaResponseDTO{
			IDTemporada:     t.IDTemporada,
			NombreTemporada: t.NombreTemporada,
			FechaInicio:     t.FechaInicio,
			FechaFin:        t.FechaFin,
			Estado: models.EstadoDTO{
				IDEstado:     t.Estado.IDEstado,
				NombreEstado: t.Estado.NombreEstado,
				TablaEstado:  t.Estado.TablaEstado,
			},
			TotalPrendas:  int(totalPrendas),
			DiasRestantes: diasRestantes,
		})
	}

	meta := utils.CalculatePaginationMeta(page, limit, total)
	middleware.SuccessResponseWithMeta(c, http.StatusOK, response, meta)
}

// GetTemporadaActiva godoc
// @Summary Obtener Temporada Activa
// @Description Obtiene la temporada actualmente activa
// @Tags Temporadas
// @Accept json
// @Produce json
// @Success 200 {object} models.StandardResponse
// @Failure 404 {object} models.StandardResponse
// @Router /api/v1/temporadas/activa [get]
func GetTemporadaActiva(c *gin.Context) {
	var temporada models.Temporada
	if err := initializers.DB.Preload("Estado").
		Where("id_estado_temporada = ?", 40).
		First(&temporada).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.ErrorResponse(c, http.StatusNotFound, "No hay temporada activa actualmente")
			return
		}
		middleware.ErrorResponse(c, http.StatusInternalServerError, "Error al obtener temporada")
		return
	}

	// Calcular métricas
	diasTranscurridos := utils.CalculateDaysElapsed(temporada.FechaInicio)
	diasRestantes := utils.CalculateDaysRemaining(temporada.FechaFin)
	porcentajeCompletado := utils.CalculatePercentageCompleted(temporada.FechaInicio, temporada.FechaFin)

	// Contar prendas
	var prendasDisponibles int64
	initializers.DB.Model(&models.Prenda{}).Where("id_temporada = ?", temporada.IDTemporada).Count(&prendasDisponibles)

	response := models.TemporadaResponseDTO{
		IDTemporada:     temporada.IDTemporada,
		NombreTemporada: temporada.NombreTemporada,
		FechaInicio:     temporada.FechaInicio,
		FechaFin:        temporada.FechaFin,
		Estado: models.EstadoDTO{
			IDEstado:     temporada.Estado.IDEstado,
			NombreEstado: temporada.Estado.NombreEstado,
		},
		DiasTranscurridos:    diasTranscurridos,
		DiasRestantes:        diasRestantes,
		PorcentajeCompletado: porcentajeCompletado,
		PrendasDisponibles:   int(prendasDisponibles),
	}

	middleware.SuccessResponse(c, http.StatusOK, response)
}

// CreateTemporada godoc
// @Summary Crear Temporada
// @Description Crea una nueva temporada
// @Tags Temporadas
// @Accept json
// @Produce json
// @Param temporada body models.CreateTemporadaDTO true "Datos de la temporada"
// @Success 201 {object} models.StandardResponse
// @Failure 400 {object} models.StandardResponse
// @Router /api/v1/temporadas [post]
func CreateTemporada(c *gin.Context) {
	var dto models.CreateTemporadaDTO
	if !middleware.ValidateRequest(c, &dto) {
		return
	}

	// Validar rango de fechas
	if err := utils.ValidateDateRange(dto.FechaInicio, dto.FechaFin); err != nil {
		middleware.ErrorResponseWithCode(c, http.StatusBadRequest, err.Error(), models.ErrorInvalidDateRange, nil)
		return
	}

	// Verificar nombre único
	var existing models.Temporada
	if err := initializers.DB.Where("nombre_temporada = ?", dto.NombreTemporada).First(&existing).Error; err == nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "Ya existe una temporada con ese nombre")
		return
	}

	// Estado por defecto: 41 (Inactiva)
	estadoDefault := 41
	if dto.IDEstadoTemporada > 0 {
		estadoDefault = dto.IDEstadoTemporada
	}

	// Si se intenta crear como activa (40), verificar que no haya otra activa
	if estadoDefault == 40 {
		var activa models.Temporada
		if err := initializers.DB.Where("id_estado_temporada = ?", 40).First(&activa).Error; err == nil {
			middleware.ErrorResponseWithCode(c, http.StatusBadRequest,
				"Ya existe una temporada activa",
				models.ErrorMultipleActiveSeasons,
				nil)
			return
		}
	}

	// Crear temporada
	temporada := models.Temporada{
		NombreTemporada:   dto.NombreTemporada,
		FechaInicio:       dto.FechaInicio,
		FechaFin:          dto.FechaFin,
		IDEstadoTemporada: estadoDefault,
	}

	if err := initializers.DB.Create(&temporada).Error; err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "Error al crear temporada")
		return
	}

	middleware.SuccessMessageResponse(c, http.StatusCreated, "Temporada creada exitosamente", gin.H{
		"id_temporada":        temporada.IDTemporada,
		"nombre_temporada":    temporada.NombreTemporada,
		"fecha_inicio":        temporada.FechaInicio,
		"fecha_fin":           temporada.FechaFin,
		"id_estado_temporada": temporada.IDEstadoTemporada,
	})
}

// ActivarTemporada godoc
// @Summary Activar Temporada
// @Description Activa una temporada (desactiva automáticamente la temporada actual)
// @Tags Temporadas
// @Accept json
// @Produce json
// @Param id_temporada path int true "ID de la temporada"
// @Success 200 {object} models.StandardResponse
// @Failure 404 {object} models.StandardResponse
// @Router /api/v1/temporadas/{id_temporada}/activar [patch]
func ActivarTemporada(c *gin.Context) {
	idTemporada := c.Param("id_temporada")

	// Verificar que la temporada a activar existe
	var temporadaNueva models.Temporada
	if err := initializers.DB.First(&temporadaNueva, "id_temporada = ?", idTemporada).Error; err != nil {
		if middleware.HandleDBError(c, err, "Temporada") {
			return
		}
	}

	// Verificar que no esté ya activa
	if temporadaNueva.IDEstadoTemporada == 40 {
		middleware.ErrorResponse(c, http.StatusBadRequest, "La temporada ya está activa")
		return
	}

	// Buscar temporada actualmente activa
	var temporadaActual models.Temporada
	err := initializers.DB.Where("id_estado_temporada = ?", 40).First(&temporadaActual).Error

	// Iniciar transacción
	tx := initializers.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var temporadaDesactivada *gin.H

	// Desactivar temporada actual si existe
	if err == nil {
		temporadaActual.IDEstadoTemporada = 41 // 41 = Inactiva
		if err := tx.Save(&temporadaActual).Error; err != nil {
			tx.Rollback()
			middleware.ErrorResponse(c, http.StatusInternalServerError, "Error al desactivar temporada actual")
			return
		}
		temporadaDesactivada = &gin.H{
			"id_temporada":        temporadaActual.IDTemporada,
			"nombre_temporada":    temporadaActual.NombreTemporada,
			"id_estado_temporada": temporadaActual.IDEstadoTemporada,
		}
	}

	// Activar nueva temporada
	temporadaNueva.IDEstadoTemporada = 40 // 40 = Activa
	if err := tx.Save(&temporadaNueva).Error; err != nil {
		tx.Rollback()
		middleware.ErrorResponse(c, http.StatusInternalServerError, "Error al activar temporada")
		return
	}

	// Commit transacción
	if err := tx.Commit().Error; err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "Error al confirmar cambios")
		return
	}

	responseData := gin.H{
		"temporada_activada": gin.H{
			"id_temporada":        temporadaNueva.IDTemporada,
			"nombre_temporada":    temporadaNueva.NombreTemporada,
			"id_estado_temporada": temporadaNueva.IDEstadoTemporada,
		},
	}

	if temporadaDesactivada != nil {
		responseData["temporada_desactivada"] = *temporadaDesactivada
	}

	middleware.SuccessMessageResponse(c, http.StatusOK, "Temporada activada exitosamente", responseData)
}
