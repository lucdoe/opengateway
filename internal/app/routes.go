package app

import (
	"github.com/gin-gonic/gin"
	h "github.com/lucdoe/capstone/internal/app/handlers"
)

const (
	relatedCasesPath = "/:id/cases"
	editsPath        = "/:id/cases/:caseID/edits"
)

func DefinitionRoutes(r *gin.RouterGroup) {
	r.GET("", h.GETDefinitions)
	r.GET("/:id", h.GETSingleDefinition)
	r.GET(relatedCasesPath, h.GETCases4Definition)
	r.POST("", h.POSTDefinition)
	r.PATCH("/:id", h.PATCHUpdateDefinition)
	r.PUT("/:id", h.PUTUpdateDefinition)
}

func SchemaRoutes(r *gin.RouterGroup) {
	r.GET("", h.GETSchemas)
	r.GET("/:id", h.GETSingleSchema)
	r.GET(relatedCasesPath, h.GETCases4Schema)
	r.POST("", h.POSTSchema)
	r.PATCH("/:id", h.PATCHSchema)
	r.PUT("/:id", h.PUTSchema)
}

func AuthorRoutes(r *gin.RouterGroup) {
	r.GET("", h.GETAuthors)
	r.GET("/:id", h.GETSingleAuthor)
	r.GET(relatedCasesPath, h.GETCasesByAuthor)
}

func PartnerRoutes(r *gin.RouterGroup) {
	r.GET("", h.GETPartners)
	r.GET("/:id", h.GETSinglePartner)
	r.POST("", h.POSTPartner)
	r.PATCH("/:id", h.PATCHPartner)
	r.PUT("/:id", h.PUTPartner)
}

func CaseRoutes(r *gin.RouterGroup) {
	r.GET("", h.GETCases)
	r.GET("/:id", h.GETSingleCase)
	r.GET("/:id/reports", h.GETCaseReports)
	r.GET("/:id/related", h.GETRelatedCases)
	r.GET("/:id/schemas", h.GETSchemas4Case)
	r.GET("/:id/definitions", h.GETDefinitions4Case)
	r.POST("", h.POSTCase)
	r.POST("/:id/reports", h.POSTCaseReport)
	r.PUT("/:id", h.PUTCase)
	r.PATCH(":id", h.PATCHCase)
}

func UserRoutes(r *gin.RouterGroup) {
	r.GET("/:id", h.GETSingleUser)
	r.GET(editsPath, h.GETUserCaseEdit)
	r.POST("", h.POSTUser)
	r.POST("/:id/cases/:caseID/finished", h.POSTUserCaseFinish)
	r.POST("/:id/cases/:caseID/saved", h.POSTUserSaveCase)
	r.POST(editsPath, h.POSTUserCaseEdit)
	r.PATCH("/:id", h.PATCHUser)
	r.PATCH(editsPath, h.PATCHUserCaseEdit)
	r.PUT("/:id", h.PUTUser)
	r.PUT(editsPath, h.PUTUserCaseEdit)
	r.PUT("/:id/cases/:caseID/saved", h.PUTUserSaveCase)
}
