package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GETCases godoc
// @Summary Get all cases
// @Description Get all cases
// @Tags Cases
// @Accept json
// @Produce json
// @Success 200 {object} CasesResponse
// @Failure 500 {object} ErrorResponse
// @Router /cases [get]
func GETCases(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetAllCasesHandler"})
}

// GETSingleCase godoc
// @Summary Get single case
// @Description Get single case
// @Tags Cases
// @Accept json
// @Produce json
// @Param id path string true "Case ID"
// @Success 200 {object} CaseResponse
// @Failure 500 {object} ErrorResponse
// @Router /cases/{id} [get]
func GETSingleCase(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetSingleCaseHandler"})
}

// POSTCase godoc
// @Summary Create case
// @Description Create case
// @Tags Cases
// @Accept json
// @Produce json
// @Param body body CaseRequest true "Case"
// @Success 200 {object} CaseResponse
// @Failure 500 {object} ErrorResponse
// @Router /cases [post]
func POSTCase(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "CreateCaseHandler"})
}

// PATCHCase godoc
// @Summary Update single case
// @Description Update single case
// @Tags Cases
// @Accept json
// @Produce json
// @Param id path string true "Case ID"
// @Param body body CaseRequest true "Case"
// @Success 200 {object} CaseResponse
// @Failure 500 {object} ErrorResponse
// @Router /cases/{id} [patch]
func PATCHCase(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "SingleUpdateCaseHandler"})
}

// PUTCase godoc
// @Summary Update full case
// @Description Update full case
// @Tags Cases
// @Accept json
// @Produce json
// @Param id path string true "Case ID"
// @Param body body CaseRequest true "Case"
// @Success 200 {object} CaseResponse
// @Failure 500 {object} ErrorResponse
// @Router /cases/{id} [put]
func PUTCase(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "FullUpdateCaseHandler"})
}

// GETCaseReports godoc
// @Summary Get all case reports
// @Description Get all case reports
// @Tags Cases
// @Accept json
// @Produce json
// @Param id path string true "Case ID"
// @Success 200 {object} CaseReportsResponse
// @Failure 500 {object} ErrorResponse
// @Router /cases/{id}/reports [get]
func GETCaseReports(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetAllCaseReportsHandler"})
}

// GETRelatedCases godoc
// @Summary Get all related cases
// @Description Get all related cases
// @Tags Cases
// @Accept json
// @Produce json
// @Param id path string true "Case ID"
// @Success 200 {object} RelatedCasesResponse
// @Failure 500 {object} ErrorResponse
// @Router /cases/{id}/related [get]
func GETRelatedCases(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetAllRelatedCasesHandler"})
}

// GETSchemas4Case godoc
// @Summary Get all related schemas
// @Description Get all related schemas
// @Tags Cases
// @Accept json
// @Produce json
// @Param id path string true "Case ID"
// @Success 200 {object} RelatedSchemasResponse
// @Failure 500 {object} ErrorResponse
// @Router /cases/{id}/schemas [get]
func GETSchemas4Case(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetAllRelatedSchemasHandler"})
}

// GETDefinitions4Case godoc
// @Summary Get all related definitions
// @Description Get all related definitions
// @Tags Cases
// @Accept json
// @Produce json
// @Param id path string true "Case ID"
// @Success 200 {object} RelatedDefinitionsResponse
// @Failure 500 {object} ErrorResponse
// @Router /cases/{id}/definitions [get]
func GETDefinitions4Case(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetAllRelatedDefinitionsHandler"})
}

// POSTCaseReport godoc
// @Summary Create case report
// @Description Create case report
// @Tags Cases
// @Accept json
// @Produce json
// @Param id path string true "Case ID"
// @Param body body CaseReportRequest true "Case Report"
// @Success 200 {object} CaseReportResponse
// @Failure 500 {object} ErrorResponse
// @Router /cases/{id}/reports [post]
func POSTCaseReport(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "CreateCaseReportHandler"})
}
