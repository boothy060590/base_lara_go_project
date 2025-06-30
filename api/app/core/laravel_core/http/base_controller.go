package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// BaseController provides common controller functionality
type BaseController struct{}

// SuccessResponse returns a success response
func (c *BaseController) SuccessResponse(ctx *gin.Context, data interface{}, message string) {
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
		"data":    data,
	})
}

// ErrorResponse returns an error response
func (c *BaseController) ErrorResponse(ctx *gin.Context, statusCode int, message string, errors interface{}) {
	ctx.JSON(statusCode, gin.H{
		"success": false,
		"message": message,
		"errors":  errors,
	})
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
	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Resource created successfully",
		"data":    resource,
	})
}

// UpdatedResponse returns an updated response (Laravel-style)
func (c *BaseController) UpdatedResponse(ctx *gin.Context, resource interface{}) {
	c.SuccessResponse(ctx, resource, "Resource updated successfully")
}

// DeletedResponse returns a deleted response (Laravel-style)
func (c *BaseController) DeletedResponse(ctx *gin.Context) {
	c.SuccessResponse(ctx, nil, "Resource deleted successfully")
}
