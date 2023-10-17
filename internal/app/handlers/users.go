package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GETSingleUser godoc
// @Summary Get single user
// @Description Get single user
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} UserResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id} [get]
func GETSingleUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetSingleUserHandle"})
}

// POSTUser godoc
// @Summary Create user
// @Description Create user
// @Tags Users
// @Accept json
// @Produce json
// @Param body body UserRequest true "User"
// @Success 200 {object} UserResponse
// @Failure 500 {object} ErrorResponse
// @Router /users [post]
func POSTUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "CreateUserHandler"})
}

// PATCHUser godoc
// @Summary Update single user
// @Description Update single user
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param body body UserRequest true "User"
// @Success 200 {object} UserResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id} [patch]
func PATCHUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "SingleUpdateUserHandler"})
}

// PUTUser godoc
// @Summary Full update single user
// @Description Full update single user
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param body body UserRequest true "User"
// @Success 200 {object} UserResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id} [put]
func PUTUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "FullUpdateUserHandler"})
}

// GETUserCaseEdit
func GETUserCaseEdit(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetSingleUserCaseEditsHandler"})
}

// POSTUserCaseFinish
func POSTUserCaseFinish(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "CreateUserCaseFinishedHandler"})
}

// POSTUserSaveCase
func POSTUserSaveCase(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "CreateUserCaseSavedHandler"})
}

// POSTUserCaseEdit
func POSTUserCaseEdit(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "CreateUserCaseEditHandler"})
}

// PATCHUserCaseEdit
func PATCHUserCaseEdit(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "UpdateUserCaseEditHandler"})
}

// PUTUserCaseEdit
func PUTUserCaseEdit(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "FullUpdateUserCaseEditHandler"})
}

// PUTUserSaveCase
func PUTUserSaveCase(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "UpdateUserCaseSavedHandler"})
}
