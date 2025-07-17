package http

import (
	app_core "base_lara_go_project/app/core/go_core"
	"net/http"

	"github.com/gin-gonic/gin"
)

// BaseController provides common controller functionality
type BaseController struct{}

// JSONResponse sends a JSON response using the optimized JSON processor
func (c *BaseController) JSONResponse(ctx *gin.Context, statusCode int, data interface{}) {
	// Use the optimized JSON processor if available
	if jsonBytes, err := app_core.EncodeJSON(data); err == nil {
		ctx.Data(statusCode, "application/json", jsonBytes)
	} else {
		// Fallback to standard JSON if optimization fails
		ctx.JSON(statusCode, data)
	}
}

// SuccessResponse returns a success response
func (c *BaseController) SuccessResponse(ctx *gin.Context, data interface{}, message string) {
	// Try to use schema-based response first
	if schemaBytes, err := app_core.RenderGlobalSchema("api_success", map[string]interface{}{
		"message": message,
		"data":    data,
	}); err == nil {
		ctx.Data(http.StatusOK, "application/json", schemaBytes)
		return
	}
	
	// Fall back to pooled response
	resp := app_core.GetPooledGenericResponse()
	defer app_core.PutPooledGenericResponse(resp)
	
	resp.Success = true
	resp.Message = message
	resp.Data = data
	
	c.JSONResponse(ctx, http.StatusOK, resp)
}

// ErrorResponse returns an error response
func (c *BaseController) ErrorResponse(ctx *gin.Context, statusCode int, message string, errors interface{}) {
	// Try to use schema-based response first
	if schemaBytes, err := app_core.RenderGlobalSchema("api_error", map[string]interface{}{
		"message": message,
		"errors":  errors,
	}); err == nil {
		ctx.Data(statusCode, "application/json", schemaBytes)
		return
	}
	
	// Fall back to pooled response
	resp := app_core.GetPooledErrorResponse()
	defer app_core.PutPooledErrorResponse(resp)
	
	resp.Success = false
	resp.Message = message
	resp.Errors = errors
	
	c.JSONResponse(ctx, statusCode, resp)
}

// ValidationErrorResponse returns a validation error response
func (c *BaseController) ValidationErrorResponse(ctx *gin.Context, errors interface{}) {
	c.ErrorResponse(ctx, http.StatusUnprocessableEntity, "Validation failed", errors)
}

// NotFoundResponse returns a not found response
func (c *BaseController) NotFoundResponse(ctx *gin.Context, message string) {
	if message == "" {
		message = "Resource not found"
	}
	c.ErrorResponse(ctx, http.StatusNotFound, message, nil)
}

// UnauthorizedResponse returns an unauthorized response
func (c *BaseController) UnauthorizedResponse(ctx *gin.Context, message string) {
	if message == "" {
		message = "Unauthorized"
	}
	c.ErrorResponse(ctx, http.StatusUnauthorized, message, nil)
}

// ForbiddenResponse returns a forbidden response
func (c *BaseController) ForbiddenResponse(ctx *gin.Context, message string) {
	if message == "" {
		message = "Forbidden"
	}
	c.ErrorResponse(ctx, http.StatusForbidden, message, nil)
}

// ServerErrorResponse returns a server error response
func (c *BaseController) ServerErrorResponse(ctx *gin.Context, message string) {
	if message == "" {
		message = "Internal server error"
	}
	c.ErrorResponse(ctx, http.StatusInternalServerError, message, nil)
}

// ResourceResponse returns a resource response (Laravel-style)
func (c *BaseController) ResourceResponse(ctx *gin.Context, resource interface{}) {
	c.SuccessResponse(ctx, resource, "Resource retrieved successfully")
}

// CollectionResponse returns a collection response (Laravel-style)
func (c *BaseController) CollectionResponse(ctx *gin.Context, collection interface{}) {
	c.SuccessResponse(ctx, collection, "Collection retrieved successfully")
}

// CreatedResponse returns a created response (Laravel-style)
func (c *BaseController) CreatedResponse(ctx *gin.Context, resource interface{}) {
	resp := app_core.GetPooledGenericResponse()
	defer app_core.PutPooledGenericResponse(resp)
	
	resp.Success = true
	resp.Message = "Resource created successfully"
	resp.Data = resource
	
	c.JSONResponse(ctx, http.StatusCreated, resp)
}

// UpdatedResponse returns an updated response (Laravel-style)
func (c *BaseController) UpdatedResponse(ctx *gin.Context, resource interface{}) {
	c.SuccessResponse(ctx, resource, "Resource updated successfully")
}

// DeletedResponse returns a deleted response (Laravel-style)
func (c *BaseController) DeletedResponse(ctx *gin.Context) {
	c.SuccessResponse(ctx, nil, "Resource deleted successfully")
}
