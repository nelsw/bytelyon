package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nelsw/bytelyon/internal/model"
	"github.com/nelsw/bytelyon/internal/service"
	"gorm.io/gorm"
)

type JobController interface {
	List(c *gin.Context)
	Save(c *gin.Context)
	Delete(c *gin.Context)
}

type jobController struct {
	service.JobService
}

func NewJobController(db *gorm.DB) JobController {
	return &jobController{service.NewJobService(db)}
}

func (ctl jobController) List(c *gin.Context) {
	if arr, err := ctl.JobService.List(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else if len(arr) == 0 {
		c.JSON(http.StatusNoContent, nil)
	} else {
		c.JSON(http.StatusOK, arr)
	}
}

func (ctl jobController) Save(c *gin.Context) {
	var a model.Job
	if err := c.ShouldBind(&a); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else if err = ctl.JobService.Save(&a); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusCreated, a)
	}

}

func (ctl jobController) Delete(c *gin.Context) {
	if c.Param("id") == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing id parameter"})
	} else if id, err := strconv.Atoi(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else if err = ctl.JobService.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusNoContent, nil)
	}
}
