package handler

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/94peter/vulpes/ezapi"
	"github.com/gin-gonic/gin"

	"seanAIgent/components/toast"
)

type componentAPI struct {
	once sync.Once
}

func (api *componentAPI) InitRouter(r ezapi.Router) {
	api.once.Do(func() {
		r.GET("/components/toast", api.getToast)
	})
}

func NewComponentApi() WebAPI {
	return &componentAPI{}
}

func (api *componentAPI) getToast(c *gin.Context) {
	title := c.Query("title")
	description := c.Query("description")
	variant := toast.Variant(c.Query("variant"))
	durationStr := c.Query("duration")

	duration, err := strconv.Atoi(durationStr)
	if err != nil {
		duration = 5000 // default duration
	}

	props := toast.Props{
		Title:       title,
		Description: description,
		Variant:     variant,
		Duration:    duration,
		Icon:        true,
		Dismissible: true,
		Position:    toast.PositionBottomRight,
	}

	comp := toast.Toast(props)
	r := newTemplRenderer(c.Request.Context(), http.StatusOK, comp)
	c.Render(http.StatusOK, r)
}

func addToastTrigger(c *gin.Context, title, description, variant string) {
	c.Header("HX-Trigger-After-Settle", fmt.Sprintf(`{"showToast":{"title":"%s","description":"%s","variant":"%s"}}`,
		base64.StdEncoding.EncodeToString([]byte(title)),
		base64.StdEncoding.EncodeToString([]byte(description)),
		variant))
}
