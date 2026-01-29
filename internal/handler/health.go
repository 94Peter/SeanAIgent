package handler

import (
	"net/http"
	"sync"

	"github.com/94peter/vulpes/db/mgo"
	"github.com/94peter/vulpes/ezapi"
	"github.com/gin-gonic/gin"
)

type healthAPI struct{}

var initHealthApiOnce sync.Once

func initHealthApi(r ezapi.Router) {
	initHealthApiOnce.Do(func() {
		api := &healthAPI{}

		r.GET("/health", api.getHealth)
	})
}

func (api *healthAPI) getHealth(c *gin.Context) {

	err := mgo.IsHealth(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}
