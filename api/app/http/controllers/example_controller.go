package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ExampleController handles example routes
type ExampleController struct{}

// NewExampleController creates a new example controller
func NewExampleController() *ExampleController {
	return &ExampleController{}
}

// Index handles the index route
func (c *ExampleController) Index(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Welcome to the API",
		"status":  "success",
	})
}

// Show handles the show route
func (c *ExampleController) Show(ctx *gin.Context) {
	id := ctx.Param("id")
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Showing item with ID: " + id,
		"status":  "success",
	})
}
