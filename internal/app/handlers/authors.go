package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GETAuthors godoc
// @Summary Get all authors
// @Description Get all authors
// @Tags Authors
// @Accept json
// @Produce json
// @Success 200 {object} AuthorsResponse
// @Failure 500 {object} ErrorResponse
// @Router /authors [get]
func GETAuthors(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetAllAuthorsHandler"})
}

// GETSingleAuthor godoc
// @Summary Get single author
// @Description Get single author
// @Tags Authors
// @Accept json
// @Produce json
// @Param id path string true "Author ID"
// @Success 200 {object} AuthorResponse
// @Failure 500 {object} ErrorResponse
// @Router /authors/{id} [get]
func GETSingleAuthor(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetSingleAuthorHandler"})
}

func GETCasesByAuthor(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetCasesByAuthorHandler"})
}
