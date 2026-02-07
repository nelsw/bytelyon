package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nelsw/bytelyon/internal/model"
	"gorm.io/gorm"
)

type JobController interface {
	List(c *gin.Context)
	ListWhereType(c *gin.Context)
	Save(c *gin.Context)
	Delete(c *gin.Context)
}

type jobController struct {
	*gorm.DB
}

func NewJobController(db *gorm.DB) JobController {
	return &jobController{db}
}

func (ctl jobController) List(c *gin.Context) {

	var arr []*model.Bot
	if err := ctl.DB.Find(&arr).Error; err != nil {
		panic(err)
	}

	if len(arr) == 0 {
		c.Status(http.StatusNoContent)
	} else {
		c.JSON(http.StatusOK, arr)
	}
}

func (ctl jobController) ListWhereType(c *gin.Context) {

	var arr []*model.Bot
	if err := ctl.DB.Scopes(Type(c.Param("type"))).Find(&arr).Error; err != nil {
		panic(err)
	}

	if len(arr) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	c.JSON(http.StatusOK, arr)
}

func (ctl jobController) Save(c *gin.Context) {
	var a model.Bot
	if err := c.Bind(&a); err != nil {
		return
	}

	if err := ctl.DB.Save(&a).Error; err != nil {
		panic(err)
	}

	c.JSON(http.StatusCreated, &a)

}

func (ctl jobController) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err = ctl.DB.Where("id = ?", id).Delete(&model.Bot{}).Error; err != nil {
		panic(err)
	}

	c.Status(http.StatusNoContent)
}

func Type(s string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("type = ?", model.BotType(s))
	}
}
