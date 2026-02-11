package controllers

import (
	"fmt"
	"ms-catalogo/initializers"
	"ms-catalogo/middleware"
	"ms-catalogo/models"
	"ms-catalogo/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ========== LEGACY ENDPOINTS (mantener para compatibilidad) ==========

// Response Structures (legacy)
type PrendaResponse struct {
	ID      string   `json:"id"`
	Nombre  string   `json:"nombre"`
	Tipo    string   `json:"tipo"`
	FotoUrl string   `json:"fotoUrl"`
	Tallas  []string `json:"tallas"`
}

type UniformeResponse struct {
	ID          string           `json:"id"`
	Nombre      string           `json:"nombre"`
	Descripcion string           `json:"descripcion"`
	Prendas     []PrendaResponse `json:"prendas"`
}

// GetUniforms godoc (legacy)
// @Summary List Uniforms (Legacy)
// @Description Get all uniforms with their garments (legacy endpoint)
// @Tags Uniforms (Legacy)
// @Accept json
// @Produce json
// @Success 200 {array} UniformeResponse
// @Router /uniforms [get]
func GetUniforms(c *gin.Context) {
	var uniformes []models.Uniforme

	result := initializers.DB.
		Preload("Prendas.TipoPrenda").
		Preload("Prendas").
		Find(&uniformes)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch uniforms"})
		return
	}

	response := make([]UniformeResponse, 0)

	for _, u := range uniformes {
		pResponses := make([]PrendaResponse, 0)
		for _, p := range u.Prendas {
			tallas := parseTallas(p.TallasDisponibles)

			pResponses = append(pResponses, PrendaResponse{
				ID:      strconv.Itoa(p.IDPrenda),
				Nombre:  p.NombrePrenda,
				Tipo:    p.TipoPrenda.NombreTipoPrenda,
				FotoUrl: p.UrlImagen,
				Tallas:  tallas,
			})
		}

		response = append(response, UniformeResponse{
			ID:          strconv.Itoa(u.IDUniforme),
			Nombre:      u.NombreUniforme,
			Descripcion: u.Descripcion,
			Prendas:     pResponses,
		})
	}

	c.JSON(http.StatusOK, response)
}

// GetGarment godoc (legacy)
// @Summary Get Garment (Legacy)
// @Description Get garment details by ID (legacy endpoint)
// @Tags Garments (Legacy)
// @Accept json
// @Produce json
// @Param id path string true "Garment ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /garments/{id} [get]
func GetGarment(c *gin.Context) {
	idStr := c.Param("id")

	var prenda models.Prenda
	if err := initializers.DB.Preload("TipoPrenda").First(&prenda, "id_prenda = ?", idStr).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Garment not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":             strconv.Itoa(prenda.IDPrenda),
		"nombre":         prenda.NombrePrenda,
		"activo":         true,
		"id_tipo_prenda": prenda.IDTipoPrenda,
	})
}

func parseTallas(csv string) []string {
	if csv == "" {
		return []string{}
	}
	parts := strings.Split(csv, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// ========== NUEVOS ENDPOINTS según contrato /api/v1 ==========

// ListUniformes godoc
// @Summary Listar Uniformes
// @Description Obtiene lista de uniformes con paginación y filtros
// @Tags Uniformes
// @Accept json
// @Produce json
// @Param page query int false "Número de página" default(1)
// @Param limit query int false "Límite por página" default(20)
// @Param id_segmento query int false "Filtrar por segmento"
// @Param search query string false "Buscar en nombre o descripción"
// @Success 200 {object} models.StandardResponse
// @Router /api/v1/uniformes [get]
func ListUniformes(c *gin.Context) {
	page := utils.ParseQueryInt(c.Query("page"), 1)
	limit := utils.ParseQueryInt(c.Query("limit"), 20)
	idSegmento := utils.ParseQueryInt(c.Query("id_segmento"), 0)
	search := c.Query("search")

	query := initializers.DB.Model(&models.Uniforme{})

	// Filtros
	if idSegmento > 0 {
		query = query.Where("id_segmento = ?", idSegmento)
	}
	if search != "" {
		searchPattern := utils.BuildLikePattern(search)
		query = query.Where("LOWER(nombre_uniforme) LIKE ? OR LOWER(descripcion) LIKE ?", searchPattern, searchPattern)
	}

	// Contar total
	var total int64
	query.Count(&total)

	// Obtener uniformes
	var uniformes []models.Uniforme
	query.Preload("Segmento").
		Scopes(utils.Paginate(page, limit)).
		Find(&uniformes)

	// Construir respuesta
	response := make([]models.UniformeListDTO, 0)
	for _, u := range uniformes {
		// Contar prendas y obtener preview
		var uniformePrendas []models.UniformePrenda
		initializers.DB.Where("id_uniforme = ?", u.IDUniforme).Find(&uniformePrendas)

		prendasPreview := make([]string, 0)
		totalPrendas := 0
		for _, up := range uniformePrendas {
			var prenda models.Prenda
			if err := initializers.DB.First(&prenda, "id_prenda = ?", up.IDPrenda).Error; err == nil {
				prendasPreview = append(prendasPreview, fmt.Sprintf("%s (%d)", prenda.NombrePrenda, up.Cantidad))
				totalPrendas += up.Cantidad
			}
		}

		response = append(response, models.UniformeListDTO{
			IDUniforme:     u.IDUniforme,
			NombreUniforme: u.NombreUniforme,
			Descripcion:    u.Descripcion,
			Segmento: models.SegmentoDTO{
				IDSegmento:     u.Segmento.IDSegmento,
				NombreSegmento: u.Segmento.NombreSegmento,
			},
			TotalPrendas:   len(uniformePrendas),
			PrendasPreview: prendasPreview,
		})
	}

	meta := utils.CalculatePaginationMeta(page, limit, total)
	middleware.SuccessResponseWithMeta(c, http.StatusOK, response, meta)
}

// GetUniforme godoc
// @Summary Obtener Uniforme por ID
// @Description Obtiene información detallada de un uniforme con todas sus prendas
// @Tags Uniformes
// @Accept json
// @Produce json
// @Param id_uniforme path int true "ID del uniforme"
// @Success 200 {object} models.StandardResponse
// @Failure 404 {object} models.StandardResponse
// @Router /api/v1/uniformes/{id_uniforme} [get]
func GetUniforme(c *gin.Context) {
	idUniforme := c.Param("id_uniforme")

	var uniforme models.Uniforme
	if err := initializers.DB.Preload("Segmento").First(&uniforme, "id_uniforme = ?", idUniforme).Error; err != nil {
		if middleware.HandleDBError(c, err, "Uniforme") {
			return
		}
	}

	// Obtener prendas con cantidad
	var uniformePrendas []models.UniformePrenda
	initializers.DB.Where("id_uniforme = ?", uniforme.IDUniforme).Find(&uniformePrendas)

	prendas := make([]models.PrendaEnUniformeDTO, 0)
	totalTipos := 0
	totalUnidades := 0

	for _, up := range uniformePrendas {
		var prenda models.Prenda
		if err := initializers.DB.Preload("TipoPrenda").Preload("Genero").
			First(&prenda, "id_prenda = ?", up.IDPrenda).Error; err == nil {

			prendas = append(prendas, models.PrendaEnUniformeDTO{
				IDPrenda:     prenda.IDPrenda,
				NombrePrenda: prenda.NombrePrenda,
				TipoPrenda: models.TipoPrendaDTO{
					IDTipoPrenda:     prenda.TipoPrenda.IDTipoPrenda,
					NombreTipoPrenda: prenda.TipoPrenda.NombreTipoPrenda,
				},
				Genero: models.GeneroDTO{
					IDGenero:     prenda.Genero.IDGenero,
					NombreGenero: prenda.Genero.NombreGenero,
				},
				Cantidad:          up.Cantidad,
				TallasDisponibles: prenda.TallasDisponibles,
				UrlImagen:         prenda.UrlImagen,
			})

			totalTipos++
			totalUnidades += up.Cantidad
		}
	}

	response := models.UniformeResponseDTO{
		IDUniforme:     uniforme.IDUniforme,
		NombreUniforme: uniforme.NombreUniforme,
		Descripcion:    uniforme.Descripcion,
		Segmento: models.SegmentoDTO{
			IDSegmento:     uniforme.Segmento.IDSegmento,
			NombreSegmento: uniforme.Segmento.NombreSegmento,
			Descripcion:    uniforme.Segmento.Descripcion,
		},
		Prendas:              prendas,
		TotalPrendasTipos:    totalTipos,
		TotalPrendasUnidades: totalUnidades,
	}

	middleware.SuccessResponse(c, http.StatusOK, response)
}

// CreateUniforme godoc
// @Summary Crear Uniforme
// @Description Crea un nuevo conjunto de uniforme
// @Tags Uniformes
// @Accept json
// @Produce json
// @Param uniforme body models.CreateUniformeDTO true "Datos del uniforme"
// @Success 201 {object} models.StandardResponse
// @Failure 400 {object} models.StandardResponse
// @Router /api/v1/uniformes [post]
func CreateUniforme(c *gin.Context) {
	var dto models.CreateUniformeDTO
	if !middleware.ValidateRequest(c, &dto) {
		return
	}

	// Validar nombre único
	var existing models.Uniforme
	if err := initializers.DB.Where("nombre_uniforme = ?", dto.NombreUniforme).First(&existing).Error; err == nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "Ya existe un uniforme con ese nombre")
		return
	}

	// Verificar que el segmento existe
	var segmento models.Segmento
	if err := initializers.DB.First(&segmento, "id_segmento = ?", dto.IDSegmento).Error; err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "Segmento no encontrado")
		return
	}

	// Verificar que todas las prendas existen
	for _, p := range dto.Prendas {
		var prenda models.Prenda
		if err := initializers.DB.First(&prenda, "id_prenda = ?", p.IDPrenda).Error; err != nil {
			middleware.ErrorResponse(c, http.StatusBadRequest,
				fmt.Sprintf("Prenda con ID %d no encontrada", p.IDPrenda))
			return
		}
	}

	// Crear uniforme
	uniforme := models.Uniforme{
		NombreUniforme: dto.NombreUniforme,
		Descripcion:    dto.Descripcion,
		IDSegmento:     dto.IDSegmento,
	}

	if err := initializers.DB.Create(&uniforme).Error; err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "Error al crear uniforme")
		return
	}

	// Agregar prendas
	for _, p := range dto.Prendas {
		uniformePrenda := models.UniformePrenda{
			IDUniforme: uniforme.IDUniforme,
			IDPrenda:   p.IDPrenda,
			Cantidad:   p.Cantidad,
		}
		if err := initializers.DB.Create(&uniformePrenda).Error; err != nil {
			middleware.ErrorResponse(c, http.StatusInternalServerError, "Error al agregar prendas")
			return
		}
	}

	middleware.SuccessMessageResponse(c, http.StatusCreated, "Uniforme creado exitosamente", gin.H{
		"id_uniforme":       uniforme.IDUniforme,
		"nombre_uniforme":   uniforme.NombreUniforme,
		"descripcion":       uniforme.Descripcion,
		"id_segmento":       uniforme.IDSegmento,
		"prendas_agregadas": len(dto.Prendas),
	})
}

// UpdateUniforme godoc
// @Summary Actualizar Uniforme
// @Description Actualiza información básica del uniforme
// @Tags Uniformes
// @Accept json
// @Produce json
// @Param id_uniforme path int true "ID del uniforme"
// @Param uniforme body models.UpdateUniformeDTO true "Datos a actualizar"
// @Success 200 {object} models.StandardResponse
// @Failure 404 {object} models.StandardResponse
// @Router /api/v1/uniformes/{id_uniforme} [put]
func UpdateUniforme(c *gin.Context) {
	idUniforme := c.Param("id_uniforme")

	var dto models.UpdateUniformeDTO
	if !middleware.ValidateRequest(c, &dto) {
		return
	}

	var uniforme models.Uniforme
	if err := initializers.DB.First(&uniforme, "id_uniforme = ?", idUniforme).Error; err != nil {
		if middleware.HandleDBError(c, err, "Uniforme") {
			return
		}
	}

	// Actualizar campos
	if dto.NombreUniforme != "" {
		uniforme.NombreUniforme = dto.NombreUniforme
	}
	if dto.Descripcion != "" {
		uniforme.Descripcion = dto.Descripcion
	}
	if dto.IDSegmento > 0 {
		uniforme.IDSegmento = dto.IDSegmento
	}

	if err := initializers.DB.Save(&uniforme).Error; err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "Error al actualizar uniforme")
		return
	}

	middleware.SuccessMessageResponse(c, http.StatusOK, "Uniforme actualizado exitosamente", gin.H{
		"id_uniforme":     uniforme.IDUniforme,
		"nombre_uniforme": uniforme.NombreUniforme,
		"descripcion":     uniforme.Descripcion,
		"id_segmento":     uniforme.IDSegmento,
	})
}

// AgregarPrendaUniforme godoc
// @Summary Agregar Prenda a Uniforme
// @Description Agrega una prenda a un uniforme existente
// @Tags Uniformes
// @Accept json
// @Produce json
// @Param id_uniforme path int true "ID del uniforme"
// @Param prenda body models.AgregarPrendaUniformeDTO true "Prenda a agregar"
// @Success 201 {object} models.StandardResponse
// @Failure 400 {object} models.StandardResponse
// @Router /api/v1/uniformes/{id_uniforme}/prendas [post]
func AgregarPrendaUniforme(c *gin.Context) {
	idUniforme := c.Param("id_uniforme")

	var dto models.AgregarPrendaUniformeDTO
	if !middleware.ValidateRequest(c, &dto) {
		return
	}

	// Verificar uniforme existe
	var uniforme models.Uniforme
	if err := initializers.DB.First(&uniforme, "id_uniforme = ?", idUniforme).Error; err != nil {
		if middleware.HandleDBError(c, err, "Uniforme") {
			return
		}
	}

	// Verificar prenda existe
	var prenda models.Prenda
	if err := initializers.DB.First(&prenda, "id_prenda = ?", dto.IDPrenda).Error; err != nil {
		if middleware.HandleDBError(c, err, "Prenda") {
			return
		}
	}

	// Verificar que no esté ya incluida
	var existing models.UniformePrenda
	if err := initializers.DB.Where("id_uniforme = ? AND id_prenda = ?", idUniforme, dto.IDPrenda).
		First(&existing).Error; err == nil {
		middleware.ErrorResponseWithCode(c, http.StatusBadRequest,
			"La prenda ya está incluida en este uniforme",
			models.ErrorDuplicatePrendaUniforme,
			nil)
		return
	}

	// Agregar prenda
	uniformePrenda := models.UniformePrenda{
		IDUniforme: uniforme.IDUniforme,
		IDPrenda:   dto.IDPrenda,
		Cantidad:   dto.Cantidad,
	}

	if err := initializers.DB.Create(&uniformePrenda).Error; err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "Error al agregar prenda")
		return
	}

	middleware.SuccessMessageResponse(c, http.StatusCreated, "Prenda agregada al uniforme exitosamente", gin.H{
		"id_uniforme":   uniforme.IDUniforme,
		"id_prenda":     prenda.IDPrenda,
		"nombre_prenda": prenda.NombrePrenda,
		"cantidad":      dto.Cantidad,
	})
}

// ActualizarCantidadPrenda godoc
// @Summary Actualizar Cantidad de Prenda en Uniforme
// @Description Actualiza la cantidad de una prenda en el uniforme
// @Tags Uniformes
// @Accept json
// @Produce json
// @Param id_uniforme path int true "ID del uniforme"
// @Param id_prenda path int true "ID de la prenda"
// @Param cantidad body models.ActualizarCantidadDTO true "Nueva cantidad"
// @Success 200 {object} models.StandardResponse
// @Failure 404 {object} models.StandardResponse
// @Router /api/v1/uniformes/{id_uniforme}/prendas/{id_prenda} [put]
func ActualizarCantidadPrenda(c *gin.Context) {
	idUniforme := c.Param("id_uniforme")
	idPrenda := c.Param("id_prenda")

	var dto models.ActualizarCantidadDTO
	if !middleware.ValidateRequest(c, &dto) {
		return
	}

	var uniformePrenda models.UniformePrenda
	if err := initializers.DB.Where("id_uniforme = ? AND id_prenda = ?", idUniforme, idPrenda).
		First(&uniformePrenda).Error; err != nil {
		middleware.NotFoundResponse(c, "Prenda en uniforme")
		return
	}

	uniformePrenda.Cantidad = dto.Cantidad
	if err := initializers.DB.Save(&uniformePrenda).Error; err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "Error al actualizar cantidad")
		return
	}

	// Obtener nombre de prenda
	var prenda models.Prenda
	initializers.DB.First(&prenda, "id_prenda = ?", idPrenda)

	middleware.SuccessMessageResponse(c, http.StatusOK, "Cantidad actualizada exitosamente", gin.H{
		"id_uniforme":   uniformePrenda.IDUniforme,
		"id_prenda":     uniformePrenda.IDPrenda,
		"nombre_prenda": prenda.NombrePrenda,
		"cantidad":      uniformePrenda.Cantidad,
	})
}

// EliminarPrendaUniforme godoc
// @Summary Eliminar Prenda de Uniforme
// @Description Elimina una prenda de un uniforme
// @Tags Uniformes
// @Accept json
// @Produce json
// @Param id_uniforme path int true "ID del uniforme"
// @Param id_prenda path int true "ID de la prenda"
// @Success 200 {object} models.StandardResponse
// @Failure 404 {object} models.StandardResponse
// @Router /api/v1/uniformes/{id_uniforme}/prendas/{id_prenda} [delete]
func EliminarPrendaUniforme(c *gin.Context) {
	idUniforme := c.Param("id_uniforme")
	idPrenda := c.Param("id_prenda")

	var uniformePrenda models.UniformePrenda
	if err := initializers.DB.Where("id_uniforme = ? AND id_prenda = ?", idUniforme, idPrenda).
		First(&uniformePrenda).Error; err != nil {
		middleware.NotFoundResponse(c, "Prenda en uniforme")
		return
	}

	if err := initializers.DB.Delete(&uniformePrenda).Error; err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "Error al eliminar prenda")
		return
	}

	middleware.SuccessMessageResponse(c, http.StatusOK, "Prenda eliminada del uniforme exitosamente", nil)
}
