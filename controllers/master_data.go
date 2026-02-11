package controllers

import (
	"fmt"
	"ms-catalogo/initializers"
	"ms-catalogo/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CampaignResponse struct {
	ID          string `json:"id"`
	Nombre      string `json:"nombre"`
	FechaInicio string `json:"fechaInicio"`
	FechaFin    string `json:"fechaFin"`
	Activa      bool   `json:"activa"`
}

type IDLabelResponse struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type MasterDataResponse struct {
	Campaign      CampaignResponse  `json:"campaign"`
	GarmentTypes  []IDLabelResponse `json:"garmentTypes"`
	Sizes         []IDLabelResponse `json:"sizes"`
	ChangeReasons []IDLabelResponse `json:"changeReasons"`
}

// GetMasterData godoc
// @Summary Get Master Data
// @Description Get active campaign, garment types, sizes, and change reasons
// @Tags Master Data
// @Accept  json
// @Produce  json
// @Success 200 {object} MasterDataResponse
// @Router /master-data [get]
func GetMasterData(c *gin.Context) {
	// 1. Campaign (Temporada Activa)
	var temporada models.Temporada
	// Assuming ID 1 is active based on request context.
	result := initializers.DB.Where("id_estado_temporada = ?", 1).First(&temporada)

	campaign := CampaignResponse{}
	if result.Error == nil {
		campaign = CampaignResponse{
			ID:          strconv.Itoa(temporada.IDTemporada),
			Nombre:      temporada.NombreTemporada,
			FechaInicio: temporada.FechaInicio,
			FechaFin:    temporada.FechaFin,
			Activa:      true,
		}
	} else {
		// Return empty inactive structure if no active season found
		campaign = CampaignResponse{Activa: false}
	}

	// 2. Garment Types
	var tipoPrendas []models.TipoPrenda
	initializers.DB.Find(&tipoPrendas)

	garmentTypes := make([]IDLabelResponse, 0)
	for _, t := range tipoPrendas {
		garmentTypes = append(garmentTypes, IDLabelResponse{
			ID:    strconv.Itoa(t.IDTipoPrenda),
			Label: t.NombreTipoPrenda,
		})
	}

	// 3. Sizes (Static)
	sizes := []IDLabelResponse{
		{ID: "S", Label: "Small"},
		{ID: "M", Label: "Medium"},
		{ID: "L", Label: "Large"},
		{ID: "XL", Label: "Extra Large"},
		{ID: "40", Label: "40 (Calzado)"},
		{ID: "42", Label: "42 (Calzado)"},
	}

	// 4. Change Reasons (Static)
	reasons := []IDLabelResponse{
		{ID: "SIZE", Label: "Talla Incorrecta"},
		{ID: "DEFECT", Label: "Producto Defectuoso"},
	}

	response := MasterDataResponse{
		Campaign:      campaign,
		GarmentTypes:  garmentTypes,
		Sizes:         sizes,
		ChangeReasons: reasons,
	}

	c.JSON(http.StatusOK, response)
}

// Helper to parse strings if necessary, though logic handles it.
func toString(id uint) string {
	return fmt.Sprintf("%d", id)
}
