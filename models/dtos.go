package models

// StandardResponse estructura estándar para todas las respuestas
type StandardResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

// ErrorInfo información detallada de errores
type ErrorInfo struct {
	Code    string      `json:"code"`
	Details interface{} `json:"details,omitempty"`
}

// PaginationMeta metadata para respuestas paginadas
type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// Códigos de error personalizados
const (
	ErrorPrendaNotFound          = "PRENDA_NOT_FOUND"
	ErrorUniformeNotFound        = "UNIFORME_NOT_FOUND"
	ErrorTemporadaNotFound       = "TEMPORADA_NOT_FOUND"
	ErrorPrendaInUse             = "PRENDA_IN_USE"
	ErrorDuplicatePrendaUniforme = "DUPLICATE_PRENDA_UNIFORME"
	ErrorInvalidDateRange        = "INVALID_DATE_RANGE"
	ErrorMultipleActiveSeasons   = "MULTIPLE_ACTIVE_SEASONS"
	ErrorInvalidTallas           = "INVALID_TALLAS"
	ErrorProveedorNotFound       = "PROVEEDOR_NOT_FOUND"
	ErrorInvalidEmpresaType      = "INVALID_EMPRESA_TYPE"
	ErrorPermissionDenied        = "PERMISSION_DENIED"
	ErrorValidation              = "VALIDATION_ERROR"
)

// =============== DTOs para Prenda ===============

// PrendaResponseDTO respuesta detallada de prenda
type PrendaResponseDTO struct {
	IDPrenda             int                   `json:"id_prenda"`
	NombrePrenda         string                `json:"nombre_prenda"`
	Descripcion          string                `json:"descripcion,omitempty"`
	Maternal             bool                  `json:"maternal"`
	UrlImagen            string                `json:"url_imagen,omitempty"`
	TallasDisponibles    string                `json:"tallas_disponibles"`
	TipoPrenda           TipoPrendaDTO         `json:"tipo_prenda"`
	Genero               GeneroDTO             `json:"genero"`
	Temporada            TemporadaBasicDTO     `json:"temporada"`
	Proveedores          []EmpresaBasicDTO     `json:"proveedores,omitempty"`
	UniformesIncluyentes []UniformeEnPrendaDTO `json:"uniformes_que_incluyen,omitempty"`
}

// CreatePrendaDTO datos para crear prenda
type CreatePrendaDTO struct {
	NombrePrenda      string `json:"nombre_prenda" binding:"required"`
	Descripcion       string `json:"descripcion"`
	Maternal          bool   `json:"maternal"`
	UrlImagen         string `json:"url_imagen"`
	TallasDisponibles string `json:"tallas_disponibles" binding:"required"`
	IDTipoPrenda      int    `json:"id_tipo_prenda" binding:"required"`
	IDGenero          int    `json:"id_genero" binding:"required"`
	IDTemporada       int    `json:"id_temporada" binding:"required"`
}

// UpdatePrendaDTO datos para actualizar prenda
type UpdatePrendaDTO struct {
	NombrePrenda      string `json:"nombre_prenda"`
	Descripcion       string `json:"descripcion"`
	Maternal          *bool  `json:"maternal"`
	UrlImagen         string `json:"url_imagen"`
	TallasDisponibles string `json:"tallas_disponibles"`
	IDTipoPrenda      int    `json:"id_tipo_prenda"`
	IDGenero          int    `json:"id_genero"`
	IDTemporada       int    `json:"id_temporada"`
}

// SearchPrendasDTO búsqueda avanzada de prendas
type SearchPrendasDTO struct {
	Nombre      string   `json:"nombre"`
	TiposPrenda []int    `json:"tipos_prenda"`
	Generos     []int    `json:"generos"`
	Temporadas  []int    `json:"temporadas"`
	Maternal    *bool    `json:"maternal"`
	Tallas      []string `json:"tallas"`
	Proveedores []int    `json:"proveedores"`
	Page        int      `json:"page"`
	Limit       int      `json:"limit"`
}

// VincularProveedorDTO vincular proveedor a prenda
type VincularProveedorDTO struct {
	IDEmpresa int `json:"id_empresa" binding:"required"`
}

// =============== DTOs para Uniforme ===============

// UniformeResponseDTO respuesta detallada de uniforme
type UniformeResponseDTO struct {
	IDUniforme           int                   `json:"id_uniforme"`
	NombreUniforme       string                `json:"nombre_uniforme"`
	Descripcion          string                `json:"descripcion,omitempty"`
	Segmento             SegmentoDTO           `json:"segmento"`
	Prendas              []PrendaEnUniformeDTO `json:"prendas"`
	TotalPrendasTipos    int                   `json:"total_prendas_tipos"`
	TotalPrendasUnidades int                   `json:"total_prendas_unidades"`
}

// UniformeListDTO respuesta para lista de uniformes
type UniformeListDTO struct {
	IDUniforme     int         `json:"id_uniforme"`
	NombreUniforme string      `json:"nombre_uniforme"`
	Descripcion    string      `json:"descripcion"`
	Segmento       SegmentoDTO `json:"segmento"`
	TotalPrendas   int         `json:"total_prendas"`
	PrendasPreview []string    `json:"prendas_preview"`
}

// CreateUniformeDTO datos para crear uniforme
type CreateUniformeDTO struct {
	NombreUniforme string                      `json:"nombre_uniforme" binding:"required"`
	Descripcion    string                      `json:"descripcion"`
	IDSegmento     int                         `json:"id_segmento" binding:"required"`
	Prendas        []PrendaUniformeCantidadDTO `json:"prendas" binding:"required,min=1"`
}

// UpdateUniformeDTO datos para actualizar uniforme
type UpdateUniformeDTO struct {
	NombreUniforme string `json:"nombre_uniforme"`
	Descripcion    string `json:"descripcion"`
	IDSegmento     int    `json:"id_segmento"`
}

// PrendaUniformeCantidadDTO prenda con cantidad para uniforme
type PrendaUniformeCantidadDTO struct {
	IDPrenda int `json:"id_prenda" binding:"required"`
	Cantidad int `json:"cantidad" binding:"required,min=1"`
}

// PrendaEnUniformeDTO prenda dentro de un uniforme
type PrendaEnUniformeDTO struct {
	IDPrenda          int           `json:"id_prenda"`
	NombrePrenda      string        `json:"nombre_prenda"`
	TipoPrenda        TipoPrendaDTO `json:"tipo_prenda"`
	Genero            GeneroDTO     `json:"genero,omitempty"`
	Cantidad          int           `json:"cantidad"`
	TallasDisponibles string        `json:"tallas_disponibles"`
	UrlImagen         string        `json:"url_imagen,omitempty"`
}

// AgregarPrendaUniformeDTO agregar prenda a uniforme
type AgregarPrendaUniformeDTO struct {
	IDPrenda int `json:"id_prenda" binding:"required"`
	Cantidad int `json:"cantidad" binding:"required,min=1"`
}

// ActualizarCantidadDTO actualizar cantidad de prenda en uniforme
type ActualizarCantidadDTO struct {
	Cantidad int `json:"cantidad" binding:"required,min=1"`
}

// UniformeEnPrendaDTO uniforme que incluye una prenda
type UniformeEnPrendaDTO struct {
	IDUniforme     int    `json:"id_uniforme"`
	NombreUniforme string `json:"nombre_uniforme"`
	Cantidad       int    `json:"cantidad"`
}

// =============== DTOs para Temporada ===============

// TemporadaResponseDTO respuesta detallada de temporada
type TemporadaResponseDTO struct {
	IDTemporada          int       `json:"id_temporada"`
	NombreTemporada      string    `json:"nombre_temporada"`
	FechaInicio          string    `json:"fecha_inicio"`
	FechaFin             string    `json:"fecha_fin"`
	Estado               EstadoDTO `json:"estado"`
	TotalPrendas         int       `json:"total_prendas,omitempty"`
	DiasRestantes        int       `json:"dias_restantes,omitempty"`
	DiasTranscurridos    int       `json:"dias_transcurridos,omitempty"`
	PorcentajeCompletado float64   `json:"porcentaje_completado,omitempty"`
	PrendasDisponibles   int       `json:"prendas_disponibles,omitempty"`
}

// CreateTemporadaDTO datos para crear temporada
type CreateTemporadaDTO struct {
	NombreTemporada   string `json:"nombre_temporada" binding:"required"`
	FechaInicio       string `json:"fecha_inicio" binding:"required"`
	FechaFin          string `json:"fecha_fin" binding:"required"`
	IDEstadoTemporada int    `json:"id_estado_temporada"`
}

// =============== DTOs Auxiliares ===============

// TipoPrendaDTO información básica de tipo de prenda
type TipoPrendaDTO struct {
	IDTipoPrenda     int    `json:"id_tipo_prenda"`
	NombreTipoPrenda string `json:"nombre_tipo_prenda"`
}

// GeneroDTO información de género
type GeneroDTO struct {
	IDGenero     int    `json:"id_genero"`
	NombreGenero string `json:"nombre_genero"`
}

// SegmentoDTO información de segmento
type SegmentoDTO struct {
	IDSegmento     int    `json:"id_segmento"`
	NombreSegmento string `json:"nombre_segmento"`
	Descripcion    string `json:"descripcion,omitempty"`
	TotalUniformes int    `json:"total_uniformes,omitempty"`
}

// EstadoDTO información de estado
type EstadoDTO struct {
	IDEstado     int    `json:"id_estado"`
	NombreEstado string `json:"nombre_estado"`
	TablaEstado  string `json:"tabla_estado,omitempty"`
}

// TemporadaBasicDTO información básica de temporada
type TemporadaBasicDTO struct {
	IDTemporada     int    `json:"id_temporada"`
	NombreTemporada string `json:"nombre_temporada"`
	FechaInicio     string `json:"fecha_inicio,omitempty"`
	FechaFin        string `json:"fecha_fin,omitempty"`
}

// EmpresaBasicDTO información básica de empresa
type EmpresaBasicDTO struct {
	IDEmpresa     int    `json:"id_empresa"`
	NombreEmpresa string `json:"nombre_empresa"`
	RutEmpresa    string `json:"rut_empresa,omitempty"`
	Email         string `json:"email,omitempty"`
}

// EmpresaResponseDTO respuesta detallada de empresa
type EmpresaResponseDTO struct {
	IDEmpresa            int            `json:"id_empresa"`
	NombreEmpresa        string         `json:"nombre_empresa"`
	RutEmpresa           string         `json:"rut_empresa"`
	Direccion            string         `json:"direccion"`
	Telefono             string         `json:"telefono"`
	Email                string         `json:"email"`
	TipoEmpresa          TipoEmpresaDTO `json:"tipo_empresa"`
	PrendasSuministradas int            `json:"prendas_suministradas,omitempty"`
}

// TipoEmpresaDTO tipo de empresa
type TipoEmpresaDTO struct {
	IDTipoEmpresa     int    `json:"id_tipo_empresa"`
	NombreTipoEmpresa string `json:"nombre_tipo_empresa"`
}

// ValidationError error de validación de campo
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
