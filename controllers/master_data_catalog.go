package controllers

import (
	"ms-catalogo/initializers"
	"ms-catalogo/middleware"
	"ms-catalogo/models"
	"ms-catalogo/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetTiposPrenda godoc
// @Summary Listar Tipos de Prenda
// @Description Obtiene lista de tipos de prenda disponibles
// @Tags Master Data
// @Accept json
// @Produce json
// @Success 200 {object} models.StandardResponse
// @Router /api/v1/tipos-prenda [get]
func GetTiposPrenda(c *gin.Context) {
	var tiposPrenda []models.TipoPrenda
	initializers.DB.Find(&tiposPrenda)

	response := make([]models.TipoPrendaDTO, 0)
	for _, t := range tiposPrenda {
		response = append(response, models.TipoPrendaDTO{
			IDTipoPrenda:     t.IDTipoPrenda,
			NombreTipoPrenda: t.NombreTipoPrenda,
		})
	}

	meta := gin.H{
		"total": len(response),
	}

	middleware.SuccessResponseWithMeta(c, http.StatusOK, response, meta)
}

// GetSegmentos godoc
// @Summary Listar Segmentos
// @Description Obtiene lista de segmentos de negocio
// @Tags Master Data
// @Accept json
// @Produce json
// @Success 200 {object} models.StandardResponse
// @Router /api/v1/segmentos [get]
func GetSegmentos(c *gin.Context) {
	var segmentos []models.Segmento
	initializers.DB.Find(&segmentos)

	response := make([]models.SegmentoDTO, 0)
	for _, s := range segmentos {
		// Contar uniformes por segmento
		var totalUniformes int64
		initializers.DB.Model(&models.Uniforme{}).Where("id_segmento = ?", s.IDSegmento).Count(&totalUniformes)

		response = append(response, models.SegmentoDTO{
			IDSegmento:     s.IDSegmento,
			NombreSegmento: s.NombreSegmento,
			Descripcion:    s.Descripcion,
			TotalUniformes: int(totalUniformes),
		})
	}

	meta := gin.H{
		"total": len(response),
	}

	middleware.SuccessResponseWithMeta(c, http.StatusOK, response, meta)
}

// GetEmpresas godoc
// @Summary Listar Empresas
// @Description Obtiene lista de empresas (proveedores/clientes) con paginación
// @Tags Master Data
// @Accept json
// @Produce json
// @Param page query int false "Número de página" default(1)
// @Param limit query int false "Límite por página (usar -1 para traer todas)" default(20)
// @Param id_tipo_empresa query int false "Filtrar por tipo (1=Cliente, 2=Proveedor, 3=Ambos)"
// @Param search query string false "Buscar en nombre o RUT"
// @Success 200 {object} models.StandardResponse
// @Router /api/v1/empresas [get]
func GetEmpresas(c *gin.Context) {
	page := utils.ParseQueryInt(c.Query("page"), 1)
	limitParam := c.Query("limit")
	limit := 20
	if limitParam != "" {
		limit = utils.ParseQueryInt(limitParam, 20)
	}

	idTipoEmpresa := utils.ParseQueryInt(c.Query("id_tipo_empresa"), 0)
	search := c.Query("search")

	query := initializers.DB.Model(&models.Empresa{})

	// Filtros
	if idTipoEmpresa > 0 {
		query = query.Where("id_tipo_empresa = ?", idTipoEmpresa)
	}
	if search != "" {
		searchPattern := utils.BuildLikePattern(search)
		query = query.Where("LOWER(razon_social) LIKE ? OR LOWER(identificador_tributario) LIKE ?", searchPattern, searchPattern)
	}

	// Contar total
	var total int64
	query.Count(&total)

	// Obtener empresas
	var empresas []models.Empresa

	if limit == -1 {
		// Traer todas sin paginación
		query.Order("razon_social ASC").Find(&empresas)
	} else {
		// Con paginación
		query.Scopes(utils.Paginate(page, limit)).
			Order("razon_social ASC").
			Find(&empresas)
	}

	// Construir respuesta
	response := make([]models.EmpresaResponseDTO, 0)
	for _, e := range empresas {
		response = append(response, models.EmpresaResponseDTO{
			IDEmpresa:               e.IDEmpresa,
			RazonSocial:             e.NombreEmpresa,
			IdentificadorTributario: e.RutEmpresa,
			Direccion:               e.Direccion,
			Telefono:                e.Telefono,
			Email:                   e.Email,
			TipoEmpresa: models.TipoEmpresaDTO{
				IDTipoEmpresa:     e.IDTipoEmpresa,
				NombreTipoEmpresa: utils.GetTipoEmpresaNombre(e.IDTipoEmpresa),
			},
		})
	}

	meta := utils.CalculatePaginationMeta(page, limit, total)
	if limit == -1 {
		meta = models.PaginationMeta{
			Page:       1,
			Limit:      int(total),
			Total:      int(total),
			TotalPages: 1,
		}
	}

	middleware.SuccessResponseWithMeta(c, http.StatusOK, response, meta)
}

// GetSegmentosByEmpresa godoc
// @Summary Listar Segmentos por Empresa
// @Description Obtiene lista de segmentos de negocio filtrados por empresa
// @Tags Master Data
// @Accept json
// @Produce json
// @Param id path int true "ID de Empresa"
// @Success 200 {object} models.StandardResponse
// @Router /api/v1/segmentos/{id} [get]
func GetSegmentosByEmpresa(c *gin.Context) {
	idEmpresa := utils.ParseQueryInt(c.Param("id"), 0)
	if idEmpresa <= 0 {
		middleware.ErrorResponseWithCode(c, http.StatusBadRequest, "Datos de entrada inválidos", models.ErrorValidation, "ID de empresa inválido")
		return
	}

	var segmentos []models.Segmento
	if err := initializers.DB.Where("id_empresa = ?", idEmpresa).Find(&segmentos).Error; err != nil {
		middleware.ErrorResponseWithCode(c, http.StatusInternalServerError, "Error al consultar segmentos", "DB_ERROR", err.Error())
		return
	}

	response := make([]models.SegmentoDTO, 0)
	for _, s := range segmentos {
		response = append(response, models.SegmentoDTO{
			IDSegmento:     s.IDSegmento,
			NombreSegmento: s.NombreSegmento,
			Descripcion:    s.Descripcion,
		})
	}

	meta := gin.H{
		"total": len(response),
	}

	middleware.SuccessResponseWithMeta(c, http.StatusOK, response, meta)
}

// GetSucursalesByEmpresa godoc
// @Summary Listar Sucursales por Empresa
// @Description Obtiene lista de sucursales asociadas a una empresa
// @Tags Master Data
// @Accept json
// @Produce json
// @Param id_empresa path int true "ID de Empresa"
// @Success 200 {object} models.StandardResponse
// @Router /api/v1/sucursales/{id_empresa} [get]
func GetSucursalesByEmpresa(c *gin.Context) {
	idEmpresa := utils.ParseQueryInt(c.Param("id_empresa"), 0)
	if idEmpresa <= 0 {
		middleware.ErrorResponseWithCode(c, http.StatusBadRequest, "Datos de entrada inválidos", models.ErrorValidation, "ID de empresa inválido")
		return
	}

	var sucursales []models.Sucursal
	// JOIN con tabla intermedia Sucursal - Empresa
	err := initializers.DB.Joins("JOIN \"Sucursal - Empresa\" se ON se.id_sucursal = \"Sucursal\".id_sucursal").
		Where("se.id_empresa = ?", idEmpresa).
		Find(&sucursales).Error

	if err != nil {
		middleware.ErrorResponseWithCode(c, http.StatusInternalServerError, "Error al consultar sucursales", "DB_ERROR", err.Error())
		return
	}

	response := make([]models.SucursalDTO, 0)
	for _, s := range sucursales {
		response = append(response, models.SucursalDTO{
			IDSucursal:     s.IDSucursal,
			NombreSucursal: s.NombreSucursal,
			Direccion:      s.Direccion,
		})
	}

	meta := gin.H{
		"total": len(response),
	}

	middleware.SuccessResponseWithMeta(c, http.StatusOK, response, meta)
}
