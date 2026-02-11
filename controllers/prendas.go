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

// ListPrendas godoc
// @Summary Listar Prendas
// @Description Obtiene lista paginada de prendas con filtros
// @Tags Prendas
// @Accept json
// @Produce json
// @Param page query int false "Número de página" default(1)
// @Param limit query int false "Límite por página" default(20)
// @Param id_tipo_prenda query int false "Filtrar por tipo de prenda"
// @Param id_genero query int false "Filtrar por género"
// @Param id_temporada query int false "Filtrar por temporada"
// @Param maternal query bool false "Solo prendas maternales"
// @Param search query string false "Buscar en nombre o descripción"
// @Param sort_by query string false "Campo para ordenar" default("nombre_prenda")
// @Param order query string false "Orden (asc|desc)" default("asc")
// @Success 200 {object} models.StandardResponse
// @Router /api/v1/prendas [get]
func ListPrendas(c *gin.Context) {
	page := utils.ParseQueryInt(c.Query("page"), 1)
	limit := utils.ParseQueryInt(c.Query("limit"), 20)
	idTipoPrenda := utils.ParseQueryInt(c.Query("id_tipo_prenda"), 0)
	idGenero := utils.ParseQueryInt(c.Query("id_genero"), 0)
	idTemporada := utils.ParseQueryInt(c.Query("id_temporada"), 0)
	maternal := utils.ParseQueryBool(c.Query("maternal"))
	search := c.Query("search")
	sortBy := c.DefaultQuery("sort_by", "nombre_prenda")
	order := c.DefaultQuery("order", "asc")

	// Base query
	query := initializers.DB.Model(&models.Prenda{})

	// Aplicar filtros
	if idTipoPrenda > 0 {
		query = query.Where("id_tipo_prenda = ?", idTipoPrenda)
	}
	if idGenero > 0 {
		query = query.Where("id_genero = ?", idGenero)
	}
	if idTemporada > 0 {
		query = query.Where("id_temporada = ?", idTemporada)
	}
	if maternal != nil {
		query = query.Where("maternal = ?", *maternal)
	}
	if search != "" {
		searchPattern := utils.BuildLikePattern(search)
		query = query.Where("LOWER(nombre_prenda) LIKE ? OR LOWER(descripcion) LIKE ?", searchPattern, searchPattern)
	}

	// Contar total
	var total int64
	query.Count(&total)

	// Ordenamiento
	if order != "asc" && order != "desc" {
		order = "asc"
	}
	orderClause := sortBy + " " + order

	// Obtener prendas paginadas
	var prendas []models.Prenda
	query.Preload("TipoPrenda").
		Preload("Genero").
		Preload("Temporada").
		Scopes(utils.Paginate(page, limit)).
		Order(orderClause).
		Find(&prendas)

	// Construir respuesta
	response := make([]models.PrendaResponseDTO, 0)
	for _, p := range prendas {
		// Obtener proveedores
		var proveedorPrendas []models.ProveedorPrenda
		initializers.DB.Preload("Empresa").Where("id_prenda = ?", p.IDPrenda).Find(&proveedorPrendas)

		proveedores := make([]models.EmpresaBasicDTO, 0)
		for _, pp := range proveedorPrendas {
			proveedores = append(proveedores, models.EmpresaBasicDTO{
				IDEmpresa:     pp.Empresa.IDEmpresa,
				NombreEmpresa: pp.Empresa.NombreEmpresa,
			})
		}

		response = append(response, models.PrendaResponseDTO{
			IDPrenda:          p.IDPrenda,
			NombrePrenda:      p.NombrePrenda,
			Descripcion:       p.Descripcion,
			Maternal:          p.Maternal,
			UrlImagen:         p.UrlImagen,
			TallasDisponibles: p.TallasDisponibles,
			TipoPrenda: models.TipoPrendaDTO{
				IDTipoPrenda:     p.TipoPrenda.IDTipoPrenda,
				NombreTipoPrenda: p.TipoPrenda.NombreTipoPrenda,
			},
			Genero: models.GeneroDTO{
				IDGenero:     p.Genero.IDGenero,
				NombreGenero: p.Genero.NombreGenero,
			},
			Temporada: models.TemporadaBasicDTO{
				IDTemporada:     p.Temporada.IDTemporada,
				NombreTemporada: p.Temporada.NombreTemporada,
			},
			Proveedores: proveedores,
		})
	}

	meta := utils.CalculatePaginationMeta(page, limit, total)
	middleware.SuccessResponseWithMeta(c, http.StatusOK, response, meta)
}

// GetPrenda godoc
// @Summary Obtener Prenda por ID
// @Description Obtiene información detallada de una prenda
// @Tags Prendas
// @Accept json
// @Produce json
// @Param id_prenda path int true "ID de la prenda"
// @Success 200 {object} models.StandardResponse
// @Failure 404 {object} models.StandardResponse
// @Router /api/v1/prendas/{id_prenda} [get]
func GetPrenda(c *gin.Context) {
	idPrenda := c.Param("id_prenda")

	var prenda models.Prenda
	if err := initializers.DB.Preload("TipoPrenda").
		Preload("Genero").
		Preload("Temporada.Estado").
		First(&prenda, "id_prenda = ?", idPrenda).Error; err != nil {
		if middleware.HandleDBError(c, err, "Prenda") {
			return
		}
	}

	// Obtener proveedores
	var proveedorPrendas []models.ProveedorPrenda
	initializers.DB.Preload("Empresa").Where("id_prenda = ?", prenda.IDPrenda).Find(&proveedorPrendas)

	proveedores := make([]models.EmpresaBasicDTO, 0)
	for _, pp := range proveedorPrendas {
		proveedores = append(proveedores, models.EmpresaBasicDTO{
			IDEmpresa:     pp.Empresa.IDEmpresa,
			NombreEmpresa: pp.Empresa.NombreEmpresa,
			RutEmpresa:    pp.Empresa.RutEmpresa,
			Email:         pp.Empresa.Email,
		})
	}

	// Obtener uniformes que incluyen esta prenda
	var uniformePrendas []models.UniformePrenda
	initializers.DB.Where("id_prenda = ?", prenda.IDPrenda).Find(&uniformePrendas)

	uniformes := make([]models.UniformeEnPrendaDTO, 0)
	for _, up := range uniformePrendas {
		var uniforme models.Uniforme
		if err := initializers.DB.First(&uniforme, "id_uniforme = ?", up.IDUniforme).Error; err == nil {
			uniformes = append(uniformes, models.UniformeEnPrendaDTO{
				IDUniforme:     uniforme.IDUniforme,
				NombreUniforme: uniforme.NombreUniforme,
				Cantidad:       up.Cantidad,
			})
		}
	}

	response := models.PrendaResponseDTO{
		IDPrenda:          prenda.IDPrenda,
		NombrePrenda:      prenda.NombrePrenda,
		Descripcion:       prenda.Descripcion,
		Maternal:          prenda.Maternal,
		UrlImagen:         prenda.UrlImagen,
		TallasDisponibles: prenda.TallasDisponibles,
		TipoPrenda: models.TipoPrendaDTO{
			IDTipoPrenda:     prenda.TipoPrenda.IDTipoPrenda,
			NombreTipoPrenda: prenda.TipoPrenda.NombreTipoPrenda,
		},
		Genero: models.GeneroDTO{
			IDGenero:     prenda.Genero.IDGenero,
			NombreGenero: prenda.Genero.NombreGenero,
		},
		Temporada: models.TemporadaBasicDTO{
			IDTemporada:     prenda.Temporada.IDTemporada,
			NombreTemporada: prenda.Temporada.NombreTemporada,
			FechaInicio:     prenda.Temporada.FechaInicio,
			FechaFin:        prenda.Temporada.FechaFin,
		},
		Proveedores:          proveedores,
		UniformesIncluyentes: uniformes,
	}

	middleware.SuccessResponse(c, http.StatusOK, response)
}

// CreatePrenda godoc
// @Summary Crear Prenda
// @Description Crea una nueva prenda en el catálogo
// @Tags Prendas
// @Accept json
// @Produce json
// @Param prenda body models.CreatePrendaDTO true "Datos de la prenda"
// @Success 201 {object} models.StandardResponse
// @Failure 400 {object} models.StandardResponse
// @Router /api/v1/prendas [post]
func CreatePrenda(c *gin.Context) {
	var dto models.CreatePrendaDTO
	if !middleware.ValidateRequest(c, &dto) {
		return
	}

	// Validar tallas
	if err := utils.ValidateTallas(dto.TallasDisponibles); err != nil {
		middleware.ErrorResponseWithCode(c, http.StatusBadRequest, err.Error(), models.ErrorInvalidTallas, nil)
		return
	}

	// Validar URL si está presente
	if dto.UrlImagen != "" && !utils.ValidateURL(dto.UrlImagen) {
		middleware.ValidationErrorResponse(c,
			fmt.Errorf("URL inválida"))
		return
	}

	// Normalizar tallas
	dto.TallasDisponibles = utils.NormalizeTallas(dto.TallasDisponibles)

	// Crear prenda
	prenda := models.Prenda{
		NombrePrenda:      dto.NombrePrenda,
		Descripcion:       dto.Descripcion,
		Maternal:          dto.Maternal,
		UrlImagen:         dto.UrlImagen,
		TallasDisponibles: dto.TallasDisponibles,
		IDTipoPrenda:      dto.IDTipoPrenda,
		IDGenero:          dto.IDGenero,
		IDTemporada:       dto.IDTemporada,
	}

	if err := initializers.DB.Create(&prenda).Error; err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "Error al crear prenda")
		return
	}

	middleware.SuccessMessageResponse(c, http.StatusCreated, "Prenda creada exitosamente", gin.H{
		"id_prenda":          prenda.IDPrenda,
		"nombre_prenda":      prenda.NombrePrenda,
		"descripcion":        prenda.Descripcion,
		"maternal":           prenda.Maternal,
		"tallas_disponibles": prenda.TallasDisponibles,
		"id_tipo_prenda":     prenda.IDTipoPrenda,
		"id_genero":          prenda.IDGenero,
		"id_temporada":       prenda.IDTemporada,
	})
}

// UpdatePrenda godoc
// @Summary Actualizar Prenda
// @Description Actualiza una prenda existente
// @Tags Prendas
// @Accept json
// @Produce json
// @Param id_prenda path int true "ID de la prenda"
// @Param prenda body models.UpdatePrendaDTO true "Datos a actualizar"
// @Success 200 {object} models.StandardResponse
// @Failure 404 {object} models.StandardResponse
// @Router /api/v1/prendas/{id_prenda} [put]
func UpdatePrenda(c *gin.Context) {
	idPrenda := c.Param("id_prenda")

	var dto models.UpdatePrendaDTO
	if !middleware.ValidateRequest(c, &dto) {
		return
	}

	var prenda models.Prenda
	if err := initializers.DB.First(&prenda, "id_prenda = ?", idPrenda).Error; err != nil {
		if middleware.HandleDBError(c, err, "Prenda") {
			return
		}
	}

	// Actualizar campos si están presentes
	if dto.NombrePrenda != "" {
		prenda.NombrePrenda = dto.NombrePrenda
	}
	if dto.Descripcion != "" {
		prenda.Descripcion = dto.Descripcion
	}
	if dto.Maternal != nil {
		prenda.Maternal = *dto.Maternal
	}
	if dto.UrlImagen != "" {
		if !utils.ValidateURL(dto.UrlImagen) {
			middleware.ValidationErrorResponse(c,
				fmt.Errorf("URL inválida"))
			return
		}
		prenda.UrlImagen = dto.UrlImagen
	}
	if dto.TallasDisponibles != "" {
		if err := utils.ValidateTallas(dto.TallasDisponibles); err != nil {
			middleware.ErrorResponseWithCode(c, http.StatusBadRequest, err.Error(), models.ErrorInvalidTallas, nil)
			return
		}
		prenda.TallasDisponibles = utils.NormalizeTallas(dto.TallasDisponibles)
	}
	if dto.IDTipoPrenda > 0 {
		prenda.IDTipoPrenda = dto.IDTipoPrenda
	}
	if dto.IDGenero > 0 {
		prenda.IDGenero = dto.IDGenero
	}
	if dto.IDTemporada > 0 {
		prenda.IDTemporada = dto.IDTemporada
	}

	if err := initializers.DB.Save(&prenda).Error; err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "Error al actualizar prenda")
		return
	}

	middleware.SuccessMessageResponse(c, http.StatusOK, "Prenda actualizada exitosamente", gin.H{
		"id_prenda":          prenda.IDPrenda,
		"nombre_prenda":      prenda.NombrePrenda,
		"descripcion":        prenda.Descripcion,
		"tallas_disponibles": prenda.TallasDisponibles,
	})
}

// DeletePrenda godoc
// @Summary Eliminar Prenda
// @Description Elimina una prenda del catálogo
// @Tags Prendas
// @Accept json
// @Produce json
// @Param id_prenda path int true "ID de la prenda"
// @Success 200 {object} models.StandardResponse
// @Failure 400 {object} models.StandardResponse
// @Router /api/v1/prendas/{id_prenda} [delete]
func DeletePrenda(c *gin.Context) {
	idPrenda := c.Param("id_prenda")

	var prenda models.Prenda
	if err := initializers.DB.First(&prenda, "id_prenda = ?", idPrenda).Error; err != nil {
		if middleware.HandleDBError(c, err, "Prenda") {
			return
		}
	}

	// Verificar si está en uso en uniformes
	var count int64
	initializers.DB.Model(&models.UniformePrenda{}).Where("id_prenda = ?", idPrenda).Count(&count)

	if count > 0 {
		middleware.ErrorResponseWithCode(c, http.StatusBadRequest,
			"No se puede eliminar la prenda",
			models.ErrorPrendaInUse,
			gin.H{"details": "La prenda está incluida en " + strconv.FormatInt(count, 10) + " uniformes activos"})
		return
	}

	// Eliminar prenda
	if err := initializers.DB.Delete(&prenda).Error; err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "Error al eliminar prenda")
		return
	}

	middleware.SuccessMessageResponse(c, http.StatusOK, "Prenda eliminada exitosamente", nil)
}

// SearchPrendas godoc
// @Summary Búsqueda Avanzada de Prendas
// @Description Búsqueda avanzada con múltiples filtros
// @Tags Prendas
// @Accept json
// @Produce json
// @Param search body models.SearchPrendasDTO true "Criterios de búsqueda"
// @Success 200 {object} models.StandardResponse
// @Router /api/v1/prendas/buscar [post]
func SearchPrendas(c *gin.Context) {
	var dto models.SearchPrendasDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		middleware.ValidationErrorResponse(c, err)
		return
	}

	// Defaults para paginación
	if dto.Page <= 0 {
		dto.Page = 1
	}
	if dto.Limit <= 0 {
		dto.Limit = 20
	}

	query := initializers.DB.Model(&models.Prenda{})

	// Aplicar filtros
	if dto.Nombre != "" {
		searchPattern := utils.BuildLikePattern(dto.Nombre)
		query = query.Where("LOWER(nombre_prenda) LIKE ?", searchPattern)
	}
	if len(dto.TiposPrenda) > 0 {
		query = query.Where("id_tipo_prenda IN ?", dto.TiposPrenda)
	}
	if len(dto.Generos) > 0 {
		query = query.Where("id_genero IN ?", dto.Generos)
	}
	if len(dto.Temporadas) > 0 {
		query = query.Where("id_temporada IN ?", dto.Temporadas)
	}
	if dto.Maternal != nil {
		query = query.Where("maternal = ?", *dto.Maternal)
	}
	if len(dto.Tallas) > 0 {
		// Buscar prendas que contengan al menos una de las tallas
		tallaConditions := make([]string, 0)
		tallaArgs := make([]interface{}, 0)
		for _, talla := range dto.Tallas {
			tallaConditions = append(tallaConditions, "tallas_disponibles LIKE ?")
			tallaArgs = append(tallaArgs, "%"+talla+"%")
		}
		query = query.Where(strings.Join(tallaConditions, " OR "), tallaArgs...)
	}
	if len(dto.Proveedores) > 0 {
		// Subquery para filtrar por proveedores
		query = query.Where("id_prenda IN (?)",
			initializers.DB.Model(&models.ProveedorPrenda{}).
				Select("id_prenda").
				Where("id_empresa IN ?", dto.Proveedores))
	}

	// Contar total
	var total int64
	query.Count(&total)

	// Obtener prendas
	var prendas []models.Prenda
	query.Preload("TipoPrenda").
		Preload("Genero").
		Preload("Temporada").
		Scopes(utils.Paginate(dto.Page, dto.Limit)).
		Find(&prendas)

	// Construir respuesta
	response := make([]gin.H, 0)
	for _, p := range prendas {
		response = append(response, gin.H{
			"id_prenda":          p.IDPrenda,
			"nombre_prenda":      p.NombrePrenda,
			"tipo_prenda":        p.TipoPrenda.NombreTipoPrenda,
			"genero":             p.Genero.NombreGenero,
			"tallas_disponibles": p.TallasDisponibles,
			"temporada":          p.Temporada.NombreTemporada,
			"maternal":           p.Maternal,
		})
	}

	meta := utils.CalculatePaginationMeta(dto.Page, dto.Limit, total)

	// Agregar información de filtros aplicados
	filtrosAplicados := gin.H{}
	if len(dto.TiposPrenda) > 0 {
		filtrosAplicados["tipos_prenda"] = len(dto.TiposPrenda)
	}
	if len(dto.Generos) > 0 {
		filtrosAplicados["generos"] = len(dto.Generos)
	}
	if len(dto.Tallas) > 0 {
		filtrosAplicados["tallas"] = len(dto.Tallas)
	}

	metaWithFilters := gin.H{
		"page":              meta.Page,
		"limit":             meta.Limit,
		"total":             meta.Total,
		"total_pages":       meta.TotalPages,
		"filtros_aplicados": filtrosAplicados,
	}

	middleware.SuccessResponseWithMeta(c, http.StatusOK, response, metaWithFilters)
}

// VincularProveedor godoc
// @Summary Vincular Proveedor a Prenda
// @Description Asocia un proveedor a una prenda
// @Tags Prendas
// @Accept json
// @Produce json
// @Param id_prenda path int true "ID de la prenda"
// @Param proveedor body models.VincularProveedorDTO true "ID del proveedor"
// @Success 201 {object} models.StandardResponse
// @Failure 400 {object} models.StandardResponse
// @Router /api/v1/prendas/{id_prenda}/proveedores [post]
func VincularProveedor(c *gin.Context) {
	idPrenda := c.Param("id_prenda")

	var dto models.VincularProveedorDTO
	if !middleware.ValidateRequest(c, &dto) {
		return
	}

	// Verificar que la prenda existe
	var prenda models.Prenda
	if err := initializers.DB.First(&prenda, "id_prenda = ?", idPrenda).Error; err != nil {
		if middleware.HandleDBError(c, err, "Prenda") {
			return
		}
	}

	// Verificar que la empresa existe y es proveedor
	var empresa models.Empresa
	if err := initializers.DB.First(&empresa, "id_empresa = ?", dto.IDEmpresa).Error; err != nil {
		if middleware.HandleDBError(c, err, "Proveedor") {
			return
		}
	}

	if empresa.IDTipoEmpresa != 2 && empresa.IDTipoEmpresa != 3 {
		middleware.ErrorResponseWithCode(c, http.StatusBadRequest,
			"La empresa no es un proveedor",
			models.ErrorInvalidEmpresaType,
			nil)
		return
	}

	// Verificar si ya está vinculado
	var existing models.ProveedorPrenda
	err := initializers.DB.Where("id_empresa = ? AND id_prenda = ?", dto.IDEmpresa, idPrenda).First(&existing).Error
	if err == nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "El proveedor ya está vinculado a esta prenda")
		return
	}

	// Crear vinculación
	proveedorPrenda := models.ProveedorPrenda{
		IDEmpresa: dto.IDEmpresa,
		IDPrenda:  prenda.IDPrenda,
	}

	if err := initializers.DB.Create(&proveedorPrenda).Error; err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "Error al vincular proveedor")
		return
	}

	middleware.SuccessMessageResponse(c, http.StatusCreated, "Proveedor vinculado exitosamente", gin.H{
		"id_prenda":      prenda.IDPrenda,
		"id_empresa":     empresa.IDEmpresa,
		"nombre_prenda":  prenda.NombrePrenda,
		"nombre_empresa": empresa.NombreEmpresa,
	})
}
