package main

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	_ "poggadaj-shared"

	"github.com/labstack/echo/v5"
)

type TemplateRenderer struct {
	templates map[string]*template.Template
}

func newTemplateRenderer() (*TemplateRenderer, error) {
	templates := map[string]*template.Template{}
	templateNames := []string{"html/home.html", "html/downloads.html"}
	for _, v := range templateNames {
		tmpl, err := template.ParseFiles("html/base.html", v)
		if err != nil {
			return nil, fmt.Errorf("failed to render template %s: %w", v, err)
		}
		templates[v] = tmpl
	}

	return &TemplateRenderer{templates: templates}, nil
}

func (t *TemplateRenderer) Render(c *echo.Context, w io.Writer, name string, data any) error {
	if tmpl, ok := t.templates[name]; ok {
		return tmpl.ExecuteTemplate(w, "base.html", data)

	}
	return errors.New("couldn't find template")
}

func main() {
	e := echo.New()

	var err error
	e.Renderer, err = newTemplateRenderer()
	if err != nil {
		panic(err)
	}

	e.GET("/", func(c *echo.Context) error {
		return c.Render(http.StatusOK, "html/home.html", nil)
	})

	e.GET("/download", func(c *echo.Context) error {
		return c.Render(http.StatusOK, "html/downloads.html", nil)
	})

	if err := e.Start(":3000"); err != nil {
		e.Logger.Error("shutting down the server", "error", err)
	}
}
