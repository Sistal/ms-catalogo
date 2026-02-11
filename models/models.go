package models

// Estado: Tabla de estados para diversas entidades
type Estado struct {
	IDEstado     int    `gorm:"primaryKey;column:id_estado"`
	NombreEstado string `gorm:"column:nombre_estado"`
	TablaEstado  string `gorm:"column:tabla_estado"`
}

func (Estado) TableName() string { return "\"Estado\"" }

// Temporada: Mapeo exacto a tabla "Temporada"
type Temporada struct {
	IDTemporada       int    `gorm:"primaryKey;column:id_temporada"`
	NombreTemporada   string `gorm:"column:nombre_temporada"`
	FechaInicio       string `gorm:"column:fecha_inicio"`
	FechaFin          string `gorm:"column:fecha_fin"`
	IDEstadoTemporada int    `gorm:"column:id_estado_temporada"` // Nota el typo en DDL 'temporada'
	Estado            Estado `gorm:"foreignKey:IDEstadoTemporada;references:IDEstado"`
}

func (Temporada) TableName() string { return "\"Temporada\"" }

// Genero: Tabla de géneros
type Genero struct {
	IDGenero     int    `gorm:"primaryKey;column:id_genero"`
	NombreGenero string `gorm:"column:nombre_genero"`
}

func (Genero) TableName() string { return "\"Genero\"" }

// Segmento: Tabla de segmentos de negocio
type Segmento struct {
	IDSegmento     int    `gorm:"primaryKey;column:id_segmento"`
	NombreSegmento string `gorm:"column:nombre_segmento"`
	Descripcion    string `gorm:"column:descripcion"`
}

func (Segmento) TableName() string { return "\"Segmento\"" }

// TipoPrenda: Mapeo exacto a tabla "Tipo Prenda"
type TipoPrenda struct {
	IDTipoPrenda     int    `gorm:"primaryKey;column:id_tipo_prenda"`
	NombreTipoPrenda string `gorm:"column:nombre_tipo_prenda"`
}

func (TipoPrenda) TableName() string { return "\"Tipo Prenda\"" }

// Empresa: Tabla de empresas (clientes/proveedores)
type Empresa struct {
	IDEmpresa     int    `gorm:"primaryKey;column:id_empresa"`
	NombreEmpresa string `gorm:"column:nombre_empresa"`
	RutEmpresa    string `gorm:"column:rut_empresa"`
	Direccion     string `gorm:"column:direccion"`
	Telefono      string `gorm:"column:telefono"`
	Email         string `gorm:"column:email"`
	IDTipoEmpresa int    `gorm:"column:id_tipo_empresa"` // 1=Cliente, 2=Proveedor, 3=Ambos
}

func (Empresa) TableName() string { return "\"Empresa\"" }

// Prenda: Mapeo exacto a tabla "Prenda"
type Prenda struct {
	IDPrenda          int        `gorm:"primaryKey;column:id_prenda"`
	NombrePrenda      string     `gorm:"column:nombre_prenda"`
	Descripcion       string     `gorm:"column:descripcion"`
	Maternal          bool       `gorm:"column:maternal"`
	UrlImagen         string     `gorm:"column:url_imagen"`
	TallasDisponibles string     `gorm:"column:tallas_disponibles"` // String separado por comas en BD
	IDTipoPrenda      int        `gorm:"column:id_tipo_prenda"`
	IDGenero          int        `gorm:"column:id_genero"`
	IDTemporada       int        `gorm:"column:id_temporada"`
	TipoPrenda        TipoPrenda `gorm:"foreignKey:IDTipoPrenda;references:IDTipoPrenda"`
	Genero            Genero     `gorm:"foreignKey:IDGenero;references:IDGenero"`
	Temporada         Temporada  `gorm:"foreignKey:IDTemporada;references:IDTemporada"`
}

func (Prenda) TableName() string { return "\"Prenda\"" }

// ProveedorPrenda: Tabla intermedia Proveedor - Prenda
type ProveedorPrenda struct {
	IDEmpresa int     `gorm:"primaryKey;column:id_empresa"`
	IDPrenda  int     `gorm:"primaryKey;column:id_prenda"`
	Empresa   Empresa `gorm:"foreignKey:IDEmpresa;references:IDEmpresa"`
	Prenda    Prenda  `gorm:"foreignKey:IDPrenda;references:IDPrenda"`
}

func (ProveedorPrenda) TableName() string { return "\"Proveedor - Prenda\"" }

// Uniforme: Mapeo base
type Uniforme struct {
	IDUniforme     int      `gorm:"primaryKey;column:id_uniforme"`
	NombreUniforme string   `gorm:"column:nombre_uniforme"`
	Descripcion    string   `gorm:"column:descripcion"`
	IDSegmento     int      `gorm:"column:id_segmento"`
	Segmento       Segmento `gorm:"foreignKey:IDSegmento;references:IDSegmento"`

	// Relación many-to-many con Prenda
	Prendas []Prenda `gorm:"many2many:Uniforme - Prenda;foreignKey:IDUniforme;joinForeignKey:id_uniforme;References:IDPrenda;joinReferences:id_prenda"`
}

func (Uniforme) TableName() string { return "\"Uniforme\"" }

// UniformePrenda: JOIN Table "Uniforme - Prenda" con cantidad
type UniformePrenda struct {
	IDUniforme int `gorm:"primaryKey;column:id_uniforme"`
	IDPrenda   int `gorm:"primaryKey;column:id_prenda"`
	Cantidad   int `gorm:"column:cantidad"`
}

func (UniformePrenda) TableName() string { return "\"Uniforme - Prenda\"" }
