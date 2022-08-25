package controller

import (
	"net/http"
	"time"

	"github.com/pilinux/gorest/database"
	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/lib/renderer"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// GetUsers - GET /users
func GetUsers(c *gin.Context) {
	db := database.GetDB()
	users := []model.User{}

	if err := db.Find(&users).Error; err != nil {
		renderer.Render(c, gin.H{"msg": "not found"}, http.StatusNotFound)
	} else {
		renderer.Render(c, users, http.StatusOK)
	}
}

// GetUser - GET /users/:id
func GetUser(c *gin.Context) {
	db := database.GetDB()
	id := c.Params.ByName("id")
	user := model.User{}

	if err := db.Where("user_id = ? ", id).First(&user).Error; err != nil {
		renderer.Render(c, gin.H{"msg": "not found"}, http.StatusNotFound)
	} else {
		renderer.Render(c, user, http.StatusOK)
	}
}

// CreateUser - POST /users
func CreateUser(c *gin.Context) {
	db := database.GetDB()
	user := model.User{}
	userFinal := model.User{}

	userIDAuth := c.GetUint64("authID")

	// does the user have an existing profile
	if err := db.Where("id_auth = ?", userIDAuth).First(&userFinal).Error; err == nil {
		renderer.Render(c, gin.H{"msg": "user profile found, no need to create a new one"}, http.StatusForbidden)
		return
	}

	// bind JSON
	if err := c.ShouldBindJSON(&user); err != nil {
		renderer.Render(c, gin.H{"msg": "bad request"}, http.StatusBadRequest)
		return
	}

	// user must not be able to manipulate all fields
	userFinal.FirstName = user.FirstName
	userFinal.LastName = user.LastName
	userFinal.IDAuth = userIDAuth

	tx := db.Begin()
	if err := tx.Create(&userFinal).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("error code: 1101")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
	} else {
		tx.Commit()
		renderer.Render(c, userFinal, http.StatusCreated)
	}
}

// UpdateUser - PUT /users
func UpdateUser(c *gin.Context) {
	db := database.GetDB()
	user := model.User{}
	userFinal := model.User{}

	userIDAuth := c.GetUint64("authID")

	// does the user have an existing profile
	if err := db.Where("id_auth = ?", userIDAuth).First(&userFinal).Error; err != nil {
		renderer.Render(c, gin.H{"msg": "no user profile found"}, http.StatusNotFound)
		return
	}

	// bind JSON
	if err := c.ShouldBindJSON(&user); err != nil {
		renderer.Render(c, gin.H{"msg": "bad request"}, http.StatusBadRequest)
		return
	}

	// user must not be able to manipulate all fields
	userFinal.UpdatedAt = time.Now()
	userFinal.FirstName = user.FirstName
	userFinal.LastName = user.LastName

	tx := db.Begin()
	if err := tx.Save(&userFinal).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("error code: 1111")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
	} else {
		tx.Commit()
		renderer.Render(c, userFinal, http.StatusOK)
	}
}
