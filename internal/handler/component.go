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

var initComponentApiOnce sync.Once

func InitComponentApi() {
	initComponentApiOnce.Do(func() {
		api := &componentAPI{}

		ezapi.RegisterGinApi(func(r ezapi.Router) {
			r.GET("/components/toast", api.getToast)
		})
	})
}

type componentAPI struct {
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
