package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GETSchemas godoc
// @Summary Get all Schemas
// @Description Get all Schemas
// @Tags Schemas
// @Accept json
// @Produce json
// @Success 200 {object} SchemasResponse
// @Failure 500 {object} ErrorResponse
// @Router /Schemas [get]
func GETSchemas(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetAllSchemasHandler"})
}

// GETSingleSchema godoc
// @Summary Get single Schema
// @Description Get single Schema
// @Tags Schemas
// @Accept json
// @Produce json
// @Param id path string true "Schema ID"
// @Success 200 {object} SchemaResponse
// @Failure 500 {object} ErrorResponse
// @Router /Schemas/{id} [get]
func GETSingleSchema(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetSingleSchemaHandler"})
}

func GETCases4Schema(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetCases4SchemaHandler"})
}

// POSTSchema godoc
// @Summary Create Schema
// @Description Create Schema
// @Tags Schemas
// @Accept json
// @Produce json
// @Param body body SchemaRequest true "Schema"
// @Success 200 {object} SchemaResponse
// @Failure 500 {object} ErrorResponse
// @Router /Schemas [post]
func POSTSchema(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "CreateSchemaHandler"})
}

// PATCHSchema godoc
// @Summary Update single Schema
// @Description Update single Schema
// @Tags Schemas
// @Accept json
// @Produce json
// @Param id path string true "Schema ID"
// @Param body body SchemaRequest true "Schema"
// @Success 200 {object} SchemaResponse
// @Failure 500 {object} ErrorResponse
// @Router /Schemas/{id} [patch]
func PATCHSchema(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "SingleUpdateSchemaHandler"})
}

// PUTSchema godoc
// @Summary Update full Schema
// @Description Update full Schema
// @Tags Schemas
// @Accept json
// @Produce json
// @Param id path string true "Schema ID"
// @Param body body SchemaRequest true "Schema"
// @Success 200 {object} SchemaResponse
// @Failure 500 {object} ErrorResponse
// @Router /Schemas/{id} [put]
func PUTSchema(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "FullUpdateSchemaHandler"})
}
