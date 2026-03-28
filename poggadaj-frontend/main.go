package main

import (
	"html/template"
	"io"
	"net/http"
	_ "poggadaj-shared"

	"github.com/labstack/echo/v5"
)

type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(c *echo.Context, w io.Writer, name string, data any) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	e := echo.New()
	e.Renderer = &TemplateRenderer{
		templates: template.Must(template.ParseGlob("html/*.html")),
	}

	e.GET("/", func(c *echo.Context) error {
		return c.Render(http.StatusOK, "index.html", nil)
	})

	if err := e.Start(":3000"); err != nil {
		e.Logger.Error("shutting down the server", "error", err)
	}
}
