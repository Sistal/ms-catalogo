package main

import (
	"log"
	"ms-catalogo/controllers"
	"ms-catalogo/initializers"
	"ms-catalogo/middleware"
	"os"

	_ "ms-catalogo/docs"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title MS-Catalogo API
// @version 1.0
// @description API del Microservicio de Catálogo de Prendas y Uniformes
// @description Gestiona el catálogo de prendas, uniformes, temporadas y datos maestros

// @contact.name Soporte API
// @contact.email soporte@sistal.com

// @license.name MIT

// @host localhost:8084
// @BasePath /

// @schemes http https

// @tag.name Prendas
// @tag.description Gestión de prendas del catálogo

// @tag.name Uniformes
// @tag.description Gestión de uniformes y conjuntos de prendas

// @tag.name Temporadas
// @tag.description Gestión de temporadas/campañas

// @tag.name Master Data
// @tag.description Datos maestros (tipos de prenda, segmentos, empresas)

// @tag.name Uniforms (Legacy)
// @tag.description Endpoints legacy para compatibilidad

// @tag.name Garments (Legacy)
// @tag.description Endpoints legacy para compatibilidad

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	// Connect to Database
	initializers.ConnectToDB()
}

func main() {
	r := gin.Default()

	// Middleware global
	r.Use(middleware.ErrorRecovery())

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Legacy routes (mantener para compatibilidad)
	r.GET("/master-data", controllers.GetMasterData)
	r.GET("/uniforms", controllers.GetUniforms)
	r.GET("/garments/:id", controllers.GetGarment)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// ========== PRENDAS ==========
		prendas := v1.Group("/prendas")
		{
			prendas.GET("", controllers.ListPrendas)                               // GET /api/v1/prendas
			prendas.GET("/:id_prenda", controllers.GetPrenda)                      // GET /api/v1/prendas/:id
			prendas.POST("", controllers.CreatePrenda)                             // POST /api/v1/prendas
			prendas.PUT("/:id_prenda", controllers.UpdatePrenda)                   // PUT /api/v1/prendas/:id
			prendas.DELETE("/:id_prenda", controllers.DeletePrenda)                // DELETE /api/v1/prendas/:id
			prendas.POST("/buscar", controllers.SearchPrendas)                     // POST /api/v1/prendas/buscar
			prendas.POST("/:id_prenda/proveedores", controllers.VincularProveedor) // POST /api/v1/prendas/:id/proveedores
		}

		// ========== UNIFORMES ==========
		uniformes := v1.Group("/uniformes")
		{
			uniformes.GET("", controllers.ListUniformes)                                             // GET /api/v1/uniformes
			uniformes.GET("/:id_uniforme", controllers.GetUniforme)                                  // GET /api/v1/uniformes/:id
			uniformes.POST("", controllers.CreateUniforme)                                           // POST /api/v1/uniformes
			uniformes.PUT("/:id_uniforme", controllers.UpdateUniforme)                               // PUT /api/v1/uniformes/:id
			uniformes.POST("/:id_uniforme/prendas", controllers.AgregarPrendaUniforme)               // POST /api/v1/uniformes/:id/prendas
			uniformes.PUT("/:id_uniforme/prendas/:id_prenda", controllers.ActualizarCantidadPrenda)  // PUT /api/v1/uniformes/:id/prendas/:id_prenda
			uniformes.DELETE("/:id_uniforme/prendas/:id_prenda", controllers.EliminarPrendaUniforme) // DELETE /api/v1/uniformes/:id/prendas/:id_prenda
		}

		// ========== TEMPORADAS ==========
		temporadas := v1.Group("/temporadas")
		{
			temporadas.GET("", controllers.ListTemporadas)                           // GET /api/v1/temporadas
			temporadas.GET("/activa", controllers.GetTemporadaActiva)                // GET /api/v1/temporadas/activa
			temporadas.POST("", controllers.CreateTemporada)                         // POST /api/v1/temporadas
			temporadas.PATCH("/:id_temporada/activar", controllers.ActivarTemporada) // PATCH /api/v1/temporadas/:id/activar
		}

		// ========== MASTER DATA ==========
		v1.GET("/tipos-prenda", controllers.GetTiposPrenda) // GET /api/v1/tipos-prenda
		v1.GET("/segmentos", controllers.GetSegmentos)      // GET /api/v1/segmentos
		v1.GET("/empresas", controllers.GetEmpresas)        // GET /api/v1/empresas
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}

	log.Printf("Starting server on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
