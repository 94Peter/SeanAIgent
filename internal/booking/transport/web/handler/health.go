package handler

import (
	"net/http"
	"sync"

	"github.com/94peter/vulpes/db/mgo"
	"github.com/94peter/vulpes/ezapi"
	"github.com/gin-gonic/gin"
)

func NewHealthApi() WebAPI {
	return &healthAPI{
		once: sync.Once{},
	}
}

type healthAPI struct {
	once sync.Once
}

func (api *healthAPI) InitRouter(r ezapi.Router) {
	api.once.Do(func() {
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
