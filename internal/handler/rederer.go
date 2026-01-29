package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"

	"github.com/a-h/templ"
)

var Default = &HTMLTemplRenderer{}

type HTMLTemplRenderer struct {
	FallbackHtmlRenderer render.HTMLRender
}

func (r *HTMLTemplRenderer) Instance(s string, d any) render.Render {
	templData, ok := d.(templ.Component)
	if !ok {
		if r.FallbackHtmlRenderer != nil {
			return r.FallbackHtmlRenderer.Instance(s, d)
		}
	}
	return &Renderer{
		Ctx:       context.Background(),
		Status:    -1,
		Component: templData,
	}
}

func newTemplRenderer(ctx context.Context, status int, component templ.Component) render.Render {
	return &Renderer{
		Ctx:       ctx,
		Status:    status,
		Component: component,
	}
}

type Renderer struct {
	Ctx       context.Context
	Component templ.Component
	Status    int
}

func (t Renderer) Render(w http.ResponseWriter) error {
	t.WriteContentType(w)
	if t.Status != -1 {
		w.WriteHeader(t.Status)
	}
	if t.Component != nil {
		return t.Component.Render(t.Ctx, w)
	}
	return nil
}

func (t Renderer) WriteContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
}

func getUserID(c *gin.Context) string {
	return c.GetString("line.liff.userid")
}

func getUserDisplayName(c *gin.Context) string {
	return c.GetString("line.liff.username")
}

func isAdmin(c *gin.Context) bool {
	return c.GetBool("isAdmin")
}
