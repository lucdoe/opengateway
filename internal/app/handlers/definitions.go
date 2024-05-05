package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GETSingleDefinitions godoc
// @Summary Get all definitions
// @Description Get all definitions
// @Tags Definitions
// @Accept json
// @Produce json
// @Success 200 {object} DefinitionsResponse
// @Failure 500 {object} ErrorResponse
// @Router /definitions [get]
func GETDefinitions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetAllDefinitionsHandler"})
}

// GETSingleDefinition godoc
// @Summary Get single definition
// @Description Get single definition
// @Tags Definitions
// @Accept json
// @Produce json
// @Param id path string true "Definition ID"
// @Success 200 {object} DefinitionResponse
// @Failure 500 {object} ErrorResponse
// @Router /definitions/{id} [get]
func GETSingleDefinition(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetSingleDefinitionHandler"})
}

// GETCases4Definition godoc
// @Summary Get all cases for definition
// @Description Get all cases for definition
// @Tags Definitions
// @Accept json
// @Produce json
// @Param id path string true "Definition ID"
// @Success 200 {object} CasesResponse
// @Failure 500 {object} ErrorResponse
// @Router /definitions/{id}/cases [get]
func GETCases4Definition(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetCases4DefinitionHandler"})
}

// POSTDefinition godoc
// @Summary Create definition
// @Description Create definition
// @Tags Definitions
// @Accept json
// @Produce json
// @Param body body DefinitionRequest true "Definition"
// @Success 200 {object} DefinitionResponse
// @Failure 500 {object} ErrorResponse
// @Router /definitions [post]
func POSTDefinition(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "CreateDefinitionHandler"})
}

// PATCHUpdateDefinition godoc
// @Summary Update single definition
// @Description Update single definition
// @Tags Definitions
// @Accept json
// @Produce json
// @Param id path string true "Definition ID"
// @Param body body DefinitionRequest true "Definition"
// @Success 200 {object} DefinitionResponse
// @Failure 500 {object} ErrorResponse
// @Router /definitions/{id} [patch]
func PATCHUpdateDefinition(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "SingleUpdateDefinitionHandler"})
}

// PUTUpdateDefinition godoc
// @Summary Update full definition
// @Description Update full definition
// @Tags Definitions
// @Accept json
// @Produce json
// @Param id path string true "Definition ID"
// @Param body body DefinitionRequest true "Definition"
// @Success 200 {object} DefinitionResponse
// @Failure 500 {object} ErrorResponse
// @Router /definitions/{id} [put]
func PUTUpdateDefinition(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "FullUpdateDefinitionHandler"})
}
