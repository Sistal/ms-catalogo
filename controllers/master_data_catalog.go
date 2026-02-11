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
// @Param limit query int false "Límite por página" default(20)
// @Param id_tipo_empresa query int false "Filtrar por tipo (1=Cliente, 2=Proveedor, 3=Ambos)"
// @Param search query string false "Buscar en nombre o RUT"
// @Success 200 {object} models.StandardResponse
// @Router /api/v1/empresas [get]
func GetEmpresas(c *gin.Context) {
	page := utils.ParseQueryInt(c.Query("page"), 1)
	limit := utils.ParseQueryInt(c.Query("limit"), 20)
	idTipoEmpresa := utils.ParseQueryInt(c.Query("id_tipo_empresa"), 0)
	search := c.Query("search")

	query := initializers.DB.Model(&models.Empresa{})

	// Filtros
	if idTipoEmpresa > 0 {
		query = query.Where("id_tipo_empresa = ?", idTipoEmpresa)
	}
	if search != "" {
		searchPattern := utils.BuildLikePattern(search)
		query = query.Where("LOWER(nombre_empresa) LIKE ? OR LOWER(rut_empresa) LIKE ?", searchPattern, searchPattern)
	}

	// Contar total
	var total int64
	query.Count(&total)

	// Obtener empresas
	var empresas []models.Empresa
	query.Scopes(utils.Paginate(page, limit)).
		Order("nombre_empresa ASC").
		Find(&empresas)

	// Construir respuesta
	response := make([]models.EmpresaResponseDTO, 0)
	for _, e := range empresas {
		// Contar prendas suministradas si es proveedor
		var prendasSuministradas int64
		if e.IDTipoEmpresa == 2 || e.IDTipoEmpresa == 3 {
			initializers.DB.Model(&models.ProveedorPrenda{}).
				Where("id_empresa = ?", e.IDEmpresa).
				Count(&prendasSuministradas)
		}

		response = append(response, models.EmpresaResponseDTO{
			IDEmpresa:     e.IDEmpresa,
			NombreEmpresa: e.NombreEmpresa,
			RutEmpresa:    e.RutEmpresa,
			Direccion:     e.Direccion,
			Telefono:      e.Telefono,
			Email:         e.Email,
			TipoEmpresa: models.TipoEmpresaDTO{
				IDTipoEmpresa:     e.IDTipoEmpresa,
				NombreTipoEmpresa: utils.GetTipoEmpresaNombre(e.IDTipoEmpresa),
			},
			PrendasSuministradas: int(prendasSuministradas),
		})
	}

	meta := utils.CalculatePaginationMeta(page, limit, total)
	middleware.SuccessResponseWithMeta(c, http.StatusOK, response, meta)
}
