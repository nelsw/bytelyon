package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nelsw/bytelyon/internal/service"
	"gorm.io/gorm"
)

type SearchController interface {
	Delete(c *gin.Context)
}

type searchController struct {
	service.SearchService
}

func NewSearchController(db *gorm.DB) SearchController {
	return &searchController{service.NewSearchService(db)}
}

func (ctl searchController) Delete(c *gin.Context) {
	if c.Param("id") == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing id parameter"})
	} else if id, err := strconv.Atoi(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else if err = ctl.SearchService.Delete(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "search deleted successfully"})
	}
}
