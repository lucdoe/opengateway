package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GETPartners godoc
// @Summary Get all Partners
// @Description Get all Partners
// @Tags Partners
// @Accept json
// @Produce json
// @Success 200 {object} PartnersResponse
// @Failure 500 {object} ErrorResponse
// @Router /Partners [get]
func GETPartners(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetAllPartnersHandler"})
}

// GETSinglePartner godoc
// @Summary Get single Partner
// @Description Get single Partner
// @Tags Partners
// @Accept json
// @Produce json
// @Param id path string true "Partner ID"
// @Success 200 {object} PartnerResponse
// @Failure 500 {object} ErrorResponse
// @Router /Partners/{id} [get]
func GETSinglePartner(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetSinglePartnerHandler"})
}

// POSTPartner godoc
// @Summary Create Partner
// @Description Create Partner
// @Tags Partners
// @Accept json
// @Produce json
// @Param body body PartnerRequest true "Partner"
// @Success 200 {object} PartnerResponse
// @Failure 500 {object} ErrorResponse
// @Router /Partners [post]
func POSTPartner(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "CreatePartnerHandler"})
}

// PATCHPartner godoc
// @Summary Update single Partner
// @Description Update single Partner
// @Tags Partners
// @Accept json
// @Produce json
// @Param id path string true "Partner ID"
// @Param body body PartnerRequest true "Partner"
// @Success 200 {object} PartnerResponse
// @Failure 500 {object} ErrorResponse
// @Router /Partners/{id} [patch]
func PATCHPartner(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "SingleUpdatePartnerHandler"})
}

// PUTPartner godoc
// @Summary Update full Partner
// @Description Update full Partner
// @Tags Partners
// @Accept json
// @Produce json
// @Param id path string true "Partner ID"
// @Param body body PartnerRequest true "Partner"
// @Success 200 {object} PartnerResponse
// @Failure 500 {object} ErrorResponse
// @Router /Partners/{id} [put]
func PUTPartner(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "FullUpdatePartnerHandler"})
}
