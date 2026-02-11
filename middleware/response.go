package middleware

import (
	"ms-catalogo/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// ErrorRecovery middleware para capturar panics
func ErrorRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				c.JSON(http.StatusInternalServerError, models.StandardResponse{
					Success: false,
					Message: "Error interno del servidor",
					Error: &models.ErrorInfo{
						Code:    "INTERNAL_SERVER_ERROR",
						Details: err,
					},
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

// SuccessResponse envía respuesta exitosa estándar
func SuccessResponse(c *gin.Context, status int, data interface{}) {
	c.JSON(status, models.StandardResponse{
		Success: true,
		Data:    data,
	})
}

// SuccessResponseWithMeta envía respuesta exitosa con metadata
func SuccessResponseWithMeta(c *gin.Context, status int, data interface{}, meta interface{}) {
	c.JSON(status, models.StandardResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

// SuccessMessageResponse envía respuesta exitosa con mensaje
func SuccessMessageResponse(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, models.StandardResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// ErrorResponse envía respuesta de error estándar
func ErrorResponse(c *gin.Context, status int, message string) {
	c.JSON(status, models.StandardResponse{
		Success: false,
		Message: message,
	})
}

// ErrorResponseWithCode envía respuesta de error con código personalizado
func ErrorResponseWithCode(c *gin.Context, status int, message string, code string, details interface{}) {
	c.JSON(status, models.StandardResponse{
		Success: false,
		Message: message,
		Error: &models.ErrorInfo{
			Code:    code,
			Details: details,
		},
	})
}

// ValidationErrorResponse envía respuesta de errores de validación
func ValidationErrorResponse(c *gin.Context, err error) {
	var validationErrors []models.ValidationError

	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			validationErrors = append(validationErrors, models.ValidationError{
				Field:   fe.Field(),
				Message: getValidationMessage(fe),
			})
		}
	} else {
		validationErrors = append(validationErrors, models.ValidationError{
			Field:   "general",
			Message: err.Error(),
		})
	}

	c.JSON(http.StatusBadRequest, models.StandardResponse{
		Success: false,
		Message: "Error en la validación",
		Errors:  validationErrors,
	})
}

// HandleDBError maneja errores de base de datos
func HandleDBError(c *gin.Context, err error, entityName string) bool {
	if err == nil {
		return false
	}

	if err == gorm.ErrRecordNotFound {
		ErrorResponseWithCode(c, http.StatusNotFound,
			entityName+" no encontrada",
			getNotFoundCode(entityName),
			nil)
		return true
	}

	ErrorResponse(c, http.StatusInternalServerError, "Error de base de datos")
	return true
}

// NotFoundResponse envía respuesta 404
func NotFoundResponse(c *gin.Context, entityName string) {
	ErrorResponseWithCode(c, http.StatusNotFound,
		entityName+" no encontrada",
		getNotFoundCode(entityName),
		nil)
}

// getValidationMessage obtiene mensaje de validación apropiado
func getValidationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "Campo requerido"
	case "min":
		return "Valor mínimo: " + fe.Param()
	case "max":
		return "Valor máximo: " + fe.Param()
	case "email":
		return "Email inválido"
	case "url":
		return "URL inválida"
	default:
		return "Validación fallida"
	}
}

// getNotFoundCode obtiene código de error para entidad no encontrada
func getNotFoundCode(entityName string) string {
	switch entityName {
	case "Prenda":
		return models.ErrorPrendaNotFound
	case "Uniforme":
		return models.ErrorUniformeNotFound
	case "Temporada":
		return models.ErrorTemporadaNotFound
	case "Proveedor":
		return models.ErrorProveedorNotFound
	default:
		return "NOT_FOUND"
	}
}

// ValidateRequest valida el request body y maneja errores
func ValidateRequest(c *gin.Context, dto interface{}) bool {
	if err := c.ShouldBindJSON(dto); err != nil {
		ValidationErrorResponse(c, err)
		return false
	}
	return true
}
